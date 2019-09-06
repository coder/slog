package slog

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"reflect"
	"strings"

	"github.com/fatih/camelcase"
	"golang.org/x/xerrors"
)

// pbval is called from within the logger to convert each field value
// into a field.
func reflectFieldValue(rv reflect.Value) fieldValue {
	if !rv.IsValid() {
		// reflect.ValueOf(nil).IsValid == false
		return nil
	}

	// We always want to look at the actual type in the interface.
	// Otherwise we cannot check if e.g. an error variable implements
	// the Value interface. If this statement was not here, we would see
	// the error variable always does not implement the Value interface
	// but does implement Error. With this, we check the concrete value instead.
	if rv.Kind() == reflect.Interface {
		return reflectFieldValue(rv.Elem())
	}

	typ := rv.Type()
	switch {
	case implements(typ, (*Value)(nil)):
		m := rv.MethodByName("LogValue")
		lv := m.Call(nil)
		return reflectFieldValue(lv[0])
	case implements(typ, (*xerrors.Formatter)(nil)):
		return extractErrorChain(rv)
	case implements(typ, (*error)(nil)):
		m := rv.MethodByName("Error")
		s := m.Call(nil)
		return fieldString(s[0].String())
	case implements(typ, (*fmt.Stringer)(nil)):
		if implements(typ, (*proto.Message)(nil)) {
			// We do not want a flat string for protobufs.
			// The reflection based struct handler below will ensure
			// protobufs values have structure in logs.
			break
		}
		m := rv.MethodByName("String")
		s := m.Call(nil)
		return fieldString(s[0].String())
	}

	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		if rv.IsNil() {
			return nil
		}
	}

	if typ == reflect.TypeOf((func() interface{})(nil)) {
		lv := rv.Call(nil)
		return reflectFieldValue(lv[0])
	}

	switch rv.Kind() {
	case reflect.String:
		return fieldString(rv.String())
	case reflect.Bool:
		return fieldBool(rv.Bool())
	case reflect.Float32, reflect.Float64:
		return fieldFloat(rv.Float())
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		return fieldInt(rv.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fieldUint(rv.Uint())
	case reflect.Ptr:
		return reflectFieldValue(rv.Elem())
	case reflect.Slice, reflect.Array:
		list := make(fieldList, rv.Len())

		for i := 0; i < rv.Len(); i++ {
			list[i] = reflectFieldValue(rv.Index(i))
		}
		return list
	case reflect.Map:
		f := make(fieldMap, 0, rv.Len())
		for _, k := range rv.MapKeys() {
			mv := rv.MapIndex(k)
			f = f.append(fmt.Sprintf("%v", k), reflectFieldValue(mv))
		}
		f.sort()
		return f
	case reflect.Struct:
		typ := rv.Type()

		f := make(fieldMap, 0, typ.NumField())

		for i := 0; i < typ.NumField(); i++ {
			ft := typ.Field(i)
			fv := rv.Field(i)

			if ft.Tag.Get("log") == "-" {
				continue
			}
			if implements(typ, (*proto.Message)(nil)) && strings.HasPrefix(ft.Name, "XXX_") {
				// Have to ignore XXX_ fields for protobuf messages.
				continue
			}

			v := reflectFieldValue(fv)
			name := ft.Tag.Get("log")
			if name == "" {
				name = snakecase(ft.Name)
			}
			f = f.append(name, v)

		}

		return f
	default:
		return fieldString(fmt.Sprintf("%v", rv))
	}
}

func snakecase(s string) string {
	splits := camelcase.Split(s)
	for i, s := range splits {
		splits[i] = strings.ToLower(s)
	}
	return strings.Join(splits, "_")
}

// This function requires that variable v be a pointer
// to the interface value that typ may implement.
// See https://blog.golang.org/laws-of-reflection
// Its the only way to grab the reflect type of an interface
// as using reflect.TypeOf on an interface grabs the
// type of its underlying value. So we pass a pointer
// to an interface and then use the Elem() method on the
// pointer reflect type to grab the type of the interface.
func implements(typ reflect.Type, v interface{}) bool {
	return typ.Implements(reflect.TypeOf(v).Elem())
}

type xerrorPrinter struct {
	msg      string
	function string
	// file:line
	loc string
}

func (p *xerrorPrinter) Print(v ...interface{}) {
	s := fmt.Sprint(v...)
	p.write(s)
}

func (p *xerrorPrinter) write(s string) {
	s = strings.TrimSpace(s)
	switch {
	case p.msg == "":
		p.msg = s
	case p.function == "":
		p.function = s
	case p.loc == "":
		p.loc = s
	default:
		panicf("unexpected String from xerrors.FormatError: %q", s)
	}
}

func (p *xerrorPrinter) Printf(f string, v ...interface{}) {
	s := fmt.Sprintf(f, v...)
	p.write(s)
}

func (p *xerrorPrinter) Detail() bool {
	return true
}

func (p *xerrorPrinter) fields() fieldMap {
	f := fieldMap{}
	f = f.append("msg", fieldString(p.msg))
	if p.loc != "" {
		f = f.append("loc", fieldString(p.loc))
	}
	if p.function != "" {
		f = f.append("fun", fieldString(p.function))
	}
	return f
}

// The value passed in must implement xerrors.Formatter.
func extractErrorChain(rv reflect.Value) fieldList {
	errs := fieldList{}

	formatError := func(p xerrors.Printer) error {
		m := rv.MethodByName("FormatError")
		err := m.Call([]reflect.Value{reflect.ValueOf(p)})
		next, _ := err[0].Interface().(error)
		return next
	}
	for {
		p := &xerrorPrinter{}
		err := formatError(p)

		errs = append(errs, p.fields())

		if err != nil {
			var ok bool
			f, ok := err.(xerrors.Formatter)
			if ok {
				formatError = func(p xerrors.Printer) error {
					return f.FormatError(p)
				}
				continue
			}
			errs = append(errs, reflectFieldValue(reflect.ValueOf(err)))
		}
		return errs
	}
}
