package core

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/fatih/camelcase"
	"github.com/golang/protobuf/proto"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"golang.org/x/xerrors"

	"go.coder.com/m/lib/log"
)

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

// pbval is called from within the logger to convert each field value
// into a *structpb.Value.
func pbval(rv reflect.Value) *structpb.Value {
	if !rv.IsValid() {
		// reflect.ValueOf(nil).IsValid == false
		return pbnull()
	}

	// We always want to look at the actual type in the interface.
	// Otherwise we cannot check if e.g. an error variable implements
	// the Value interface. If this statement was not here, we would see
	// the error variable always does not implement the Value interface
	// but does implement Error. With this, we check the concrete value instead.
	if rv.Kind() == reflect.Interface {
		return pbval(rv.Elem())
	}

	typ := rv.Type()
	switch {
	case implements(typ, (*log.Value)(nil)):
		m := rv.MethodByName("LogValue")
		lv := m.Call(nil)
		return pbval(lv[0])
	case implements(typ, (*xerrors.Formatter)(nil)):
		chain := extractErrorChain(rv)
		return pblist(chain)
	case implements(typ, (*error)(nil)):
		m := rv.MethodByName("Error")
		s := m.Call(nil)
		return PBString(s[0].String())
	case implements(typ, (*fmt.Stringer)(nil)):
		if implements(typ, (*proto.Message)(nil)) {
			// We do not want a flat string for protobufs.
			// The reflection based struct handler below will ensure
			// protobufs values have structure in logs.
			break
		}
		m := rv.MethodByName("String")
		s := m.Call(nil)
		return PBString(s[0].String())
	}

	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		if rv.IsNil() {
			return pbnull()
		}
	}

	if typ == reflect.TypeOf((func() interface{})(nil)) {
		lv := rv.Call(nil)
		return pbval(lv[0])
	}

	switch rv.Kind() {
	case reflect.String:
		return PBString(rv.String())
	case reflect.Bool:
		return pbbool(rv.Bool())
	case reflect.Float32, reflect.Float64:
		return pbfloat64(rv.Float())
	case reflect.Ptr:
		return pbval(rv.Elem())
	case reflect.Slice, reflect.Array:
		list := &structpb.ListValue{
			Values: make([]*structpb.Value, rv.Len()),
		}

		for i := 0; i < rv.Len(); i++ {
			el := rv.Index(i)
			list.Values[i] = pbval(el)
		}

		return pblist(list)
	case reflect.Map:
		s := &structpb.Struct{
			Fields: make(map[string]*structpb.Value, rv.Len()),
		}

		for _, k := range rv.MapKeys() {
			mv := rv.MapIndex(k)
			s.Fields[fmt.Sprintf("%v", k)] = pbval(mv)
		}

		return pbstruct(s)
	case reflect.Struct:
		typ := rv.Type()

		s := &structpb.Struct{
			Fields: make(map[string]*structpb.Value, typ.NumField()),
		}

		for i := 0; i < typ.NumField(); i++ {
			ft := typ.Field(i)
			fv := rv.Field(i)

			if strings.HasPrefix(ft.Name, "XXX_") {
				// Extremely likely the field is a protobuf internal field that is exported.
				continue
			}

			v := pbval(fv)
			s.Fields[snakecase(ft.Name)] = v
		}

		return pbstruct(s)
	default:
		return PBString(fmt.Sprintf("%v", rv))
	}
}

func snakecase(s string) string {
	splits := camelcase.Split(s)
	for i, s := range splits {
		splits[i] = strings.ToLower(s)
	}
	return strings.Join(splits, "_")
}

func pbnull() *structpb.Value {
	return &structpb.Value{
		Kind: &structpb.Value_NullValue{
			NullValue: structpb.NullValue_NULL_VALUE,
		},
	}
}

func PBString(s string) *structpb.Value {
	return &structpb.Value{
		Kind: &structpb.Value_StringValue{
			StringValue: s,
		},
	}
}

func pbbool(b bool) *structpb.Value {
	return &structpb.Value{
		Kind: &structpb.Value_BoolValue{
			BoolValue: b,
		},
	}
}

func pbfloat64(f float64) *structpb.Value {
	return &structpb.Value{
		Kind: &structpb.Value_NumberValue{
			NumberValue: f,
		},
	}
}

func pbstruct(s *structpb.Struct) *structpb.Value {
	return &structpb.Value{
		Kind: &structpb.Value_StructValue{
			StructValue: s,
		},
	}
}

func pblist(list *structpb.ListValue) *structpb.Value {
	return &structpb.Value{
		Kind: &structpb.Value_ListValue{
			list,
		},
	}
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
		panicf("unexpected String from xerror.FormatError: %q", s)
	}
}

func (p *xerrorPrinter) Printf(f string, v ...interface{}) {
	s := fmt.Sprintf(f, v...)
	p.write(s)
}

func (p *xerrorPrinter) Detail() bool {
	return true
}

func (p *xerrorPrinter) pbstruct() *structpb.Struct {
	s := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"err": PBString(p.msg),
		},
	}

	if p.loc != "" {
		s.Fields["loc"] = PBString(p.loc)
	}

	if p.function != "" {
		s.Fields["fun"] = PBString(p.function)
	}

	return s
}

// The value passed in must implement xerrors.Formatter.
func extractErrorChain(rv reflect.Value) *structpb.ListValue {
	list := &structpb.ListValue{}

	formatError := func(p xerrors.Printer) error {
		m := rv.MethodByName("FormatError")
		err := m.Call([]reflect.Value{reflect.ValueOf(p)})
		next, _ := err[0].Interface().(error)
		return next
	}
	for {
		p := &xerrorPrinter{}
		err := formatError(p)

		list.Values = append(list.Values, pbstruct(p.pbstruct()))

		if err != nil {
			var ok bool
			f, ok := err.(xerrors.Formatter)
			if ok {
				formatError = func(p xerrors.Printer) error {
					return f.FormatError(p)
				}
				continue
			}
			list.Values = append(list.Values, pbval(reflect.ValueOf(err)))
		}
		return list
	}
}
