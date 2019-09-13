package slog

import (
	"encoding"
	"fmt"
	"reflect"
	"strings"

	"github.com/fatih/camelcase"
	"github.com/golang/protobuf/proto"
	"golang.org/x/xerrors"

	"go.coder.com/slog/slogval"
)

// Encode encodes the interface to a structured and easily
// introspectable slogval.Value.
func Encode(v interface{}) slogval.Value {
	return encode(reflect.ValueOf(v))
}

func encode(rv reflect.Value) slogval.Value {
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
		return encode(rv.Elem())
	}

	typ := rv.Type()
	switch {
	case implements(typ, (*slogval.Value)(nil)):
		// We cannot use rv.Interface() directly, instead
		// we rely on SlogValue() to return the slogval.Value.
		v := rv.MethodByName("SlogValue").Call(nil)
		return v[0].Interface().(slogval.Value)
	case typ == reflect.TypeOf(jsonValue{}):
		rv = rv.FieldByName("V").Elem()
		typ = rv.Type()
		if rv.Kind() != reflect.Struct {
			return encode(rv)
		}
		m := make(slogval.Map, 0, typ.NumField())
		m = reflectStruct(m, rv, typ, true)
		return m
	case implements(typ, (*Value)(nil)):
		m := rv.MethodByName("LogValue")
		lv := m.Call(nil)
		return encode(lv[0])
	case implements(typ, (*encoding.TextMarshaler)(nil)):
		return marshalText(rv)
	case implements(typ, (*xerrors.Formatter)(nil)):
		return extractXErrorChain(rv)
	case implements(typ, (*error)(nil)):
		m := rv.MethodByName("Error")
		s := m.Call(nil)
		return slogval.String(s[0].String())
	case implements(typ, (*fmt.Stringer)(nil)):
		if implements(typ, (*proto.Message)(nil)) {
			// We do not want a flat string for protobufs.
			// The reflection based struct handler below will ensure
			// protobufs values have structure in logs.
			break
		}
		s := rv.MethodByName("String").Call(nil)
		return slogval.String(s[0].String())
	}

	// Fallback to reflection.
	return reflectValue(rv)
}

func marshalText(rv reflect.Value) slogval.Value {
	rs := rv.MethodByName("MarshalText").Call(nil)
	bytes := rs[0]
	err := rs[1]

	if !err.IsNil() {
		return Encode(map[string]reflect.Value{
			"marshalTextError": err,
		})
	}
	return Encode(string(bytes.Bytes()))
}

func reflectValue(rv reflect.Value) slogval.Value {
	if !rv.IsValid() {
		// reflect.ValueOf(nil).IsValid == false
		return nil
	}

	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		if rv.IsNil() {
			return nil
		}
	}

	typ := rv.Type()
	switch rv.Kind() {
	case reflect.String:
		return slogval.String(rv.String())
	case reflect.Bool:
		return slogval.Bool(rv.Bool())
	case reflect.Float32, reflect.Float64:
		return slogval.Float(rv.Float())
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		return slogval.Int(rv.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return slogval.Uint(rv.Uint())
	case reflect.Ptr:
		return encode(rv.Elem())
	case reflect.Slice, reflect.Array:
		// Ordered map.
		if typ == reflect.TypeOf([]Field(nil)) {
			m := make(slogval.Map, 0, rv.Len())
			for i := 0; i < rv.Len(); i++ {
				f := rv.Index(i)
				key := f.MethodByName("LogKey").Call(nil)[0].String()
				val := f.MethodByName("LogValue").Call(nil)[0]
				m = m.Append(key, encode(val))
			}
			return m
		}
		list := make(slogval.List, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			list[i] = encode(rv.Index(i))
		}
		return list
	case reflect.Map:
		m := make(slogval.Map, 0, rv.Len())
		for _, k := range rv.MapKeys() {
			mv := rv.MapIndex(k)
			m = m.Append(fmt.Sprintf("%v", k), encode(mv))
		}
		m.Sort()
		return m
	case reflect.Struct:
		m := make(slogval.Map, 0, typ.NumField())
		m = reflectStruct(m, rv, typ, false)
		return m
	default:
		return slogval.String(fmt.Sprintf("%+v", rv))
	}
}

func reflectStruct(m slogval.Map, rv reflect.Value, structTyp reflect.Type, json bool) slogval.Map {
	for i := 0; i < structTyp.NumField(); i++ {
		typ := structTyp.Field(i)
		rv := rv.Field(i)

		if implements(structTyp, (*proto.Message)(nil)) && strings.HasPrefix(typ.Name, "XXX_") {
			// Have to ignore XXX_ fields for protobuf messages.
			continue
		}

		// Ignore unexported fields in JSON mode.
		if json && typ.PkgPath != "" {
			continue
		}

		tag := typ.Tag.Get("json")
		if tag == "-" {
			continue
		}

		tagFieldName, opts := parseTag(tag)
		if _, ok := opts["omitempty"]; ok && isEmptyValue(rv) {
			continue
		}

		if typ.Anonymous {
			m = reflectStruct(m, rv, typ.Type, json)
			continue
		}
		if tagFieldName == "" {
			tagFieldName = snakecase(typ.Name)
		}

		// If the field is unexported, we want to use
		// reflectValue to encode as we cannot call any
		// methods on it so the interfaces are useless.
		if typ.PkgPath == "" {
			m = m.Append(tagFieldName, encode(rv))
		} else {
			m = m.Append(tagFieldName, reflectValue(rv))
		}

	}
	return m
}

func parseTag(tag string) (name string, opts map[string]struct{}) {
	s := strings.Split(tag, ",")
	if len(s) == 1 {
		return s[0], nil
	}
	opts = make(map[string]struct{})
	for _, opt := range s[1:] {
		opts[opt] = struct{}{}
	}
	return s[0], opts
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

type wrapError struct {
	Msg string `json:"msg"`
	Fun string `json:"fun"`
	// file:line
	Loc string `json:"loc"`
}

func (e wrapError) Error() string {
	return fmt.Sprintf("wrap error: %+v", e)
}

func (e wrapError) LogValue() interface{} {
	return JSON(e)
}

type xerrorPrinter struct {
	e wrapError
}

func (p *xerrorPrinter) Print(v ...interface{}) {
	s := fmt.Sprint(v...)
	p.write(s)
}

func (p *xerrorPrinter) write(s string) {
	s = strings.TrimSpace(s)
	switch {
	case p.e.Msg == "":
		p.e.Msg = s
	case p.e.Fun == "":
		p.e.Fun = s
	case p.e.Loc == "":
		p.e.Loc = s
	default:
		panic(fmt.Sprintf("slogval: unexpected String from xerrors.FormatError: %q", s))
	}
}

func (p *xerrorPrinter) Printf(f string, v ...interface{}) {
	s := fmt.Sprintf(f, v...)
	p.write(s)
}

func (p *xerrorPrinter) Detail() bool {
	return true
}

// The value passed in must implement xerrors.Formatter.
func extractXErrorChain(rv reflect.Value) slogval.List {
	errs := slogval.List{}

	formatError := func(p xerrors.Printer) error {
		m := rv.MethodByName("FormatError")
		err := m.Call([]reflect.Value{reflect.ValueOf(p)})
		next, _ := err[0].Interface().(error)
		return next
	}
	for {
		p := &xerrorPrinter{}
		next := formatError(p)

		errs = append(errs, Encode(p.e))

		if next != nil {
			var ok bool
			f, ok := next.(xerrors.Formatter)
			if ok {
				formatError = func(p xerrors.Printer) error {
					return f.FormatError(p)
				}
				continue
			}
			errs = append(errs, Encode(next))
		}
		return errs
	}
}

// Copied from encode.go in encoding/json.
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}
