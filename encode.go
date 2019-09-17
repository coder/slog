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
	return encodeInterface(v)
}

func encode(rv reflect.Value, pure bool) slogval.Value {
	if pure {
		return encodeReflect(rv, true)
	}
	return encodeInterface(rv.Interface())
}

func encodeInterface(v interface{}) slogval.Value {
	if v == nil {
		return nil
	}

	switch v := v.(type) {
	case jsonValue:
		rv := reflect.ValueOf(v.V)
		if rv.Kind() != reflect.Struct {
			return Encode(v.V)
		}
		m := make(slogval.Map, 0, rv.NumField())
		m = reflectStruct(m, rv, rv.Type(), true, false)
		return m
	case []Field:
		m := make(slogval.Map, 0, len(v))
		for _, f := range v {
			m = m.Append(f.LogKey(), Encode(f.LogValue()))
		}
		return m
	case Value:
		return Encode(v.LogValue())
	case proto.Message:
		return encodeReflect(reflect.ValueOf(v), false)
	case encoding.TextMarshaler:
		return marshalText(v)
	case xerrors.Formatter:
		return extractXErrorChain(v)
	case error, fmt.Stringer:
		// Cannot use %+v here as if its not a xerrors.Formatter
		// then fmt will use its reflection encoder for the value
		// instead of v.String() or v.Error().
		return Encode(fmt.Sprintf("%v", v))
	default:
		return encodeReflect(reflect.ValueOf(v), false)
	}
}

func marshalText(v encoding.TextMarshaler) slogval.Value {
	b, err := v.MarshalText()
	if err != nil {
		return Encode(map[string]error{
			"marshalTextError": err,
		})
	}
	return Encode(string(b))
}

func encodeReflect(rv reflect.Value, pure bool) slogval.Value {
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
		return encode(rv.Elem(), pure)
	case reflect.Interface:
		return encode(rv.Elem(), pure)
	case reflect.Slice, reflect.Array:
		// Ordered map.
		if typ == reflect.TypeOf(slogval.Map(nil)) {
			m := make(slogval.Map, 0, rv.Len())
			for i := 0; i < rv.Len(); i++ {
				f := rv.Index(i)
				key := f.FieldByName("Name").String()
				val := f.FieldByName("Value")
				m = m.Append(key, encode(val, pure))
			}
			return m
		}
		list := make(slogval.List, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			list[i] = encode(rv.Index(i), pure)
		}
		return list
	case reflect.Map:
		m := make(slogval.Map, 0, rv.Len())
		for _, k := range rv.MapKeys() {
			mv := rv.MapIndex(k)
			m = m.Append(fmt.Sprintf("%v", k), encode(mv, pure))
		}
		m.Sort()
		return m
	case reflect.Struct:
		m := make(slogval.Map, 0, typ.NumField())
		m = reflectStruct(m, rv, typ, false, pure)
		return m
	default:
		return slogval.String(fmt.Sprintf("%+v", rv))
	}
}

func reflectStruct(m slogval.Map, rv reflect.Value, structTyp reflect.Type, json, pure bool) slogval.Map {
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

		tagFieldName := snakecase(typ.Name)
		if json {
			tag := typ.Tag.Get("json")
			if tag == "-" {
				continue
			}

			var opts map[string]struct{}
			tagFieldName, opts = parseTag(tag)
			if _, ok := opts["omitempty"]; ok && isEmptyValue(rv) {
				continue
			}
		}

		pure := pure
		if typ.PkgPath != "" {
			// We need to use pure reflection as the field is unexported.
			pure = true
		}

		if typ.Anonymous {
			m = reflectStruct(m, rv, typ.Type, json, pure)
		} else {
			m = m.Append(tagFieldName, encode(rv, pure))
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
func extractXErrorChain(f xerrors.Formatter) slogval.List {
	errs := slogval.List{}

	for {
		p := &xerrorPrinter{}
		next := f.FormatError(p)
		errs = append(errs, Encode(p.e))

		if next == nil {
			return errs
		}

		var ok bool
		f, ok = next.(xerrors.Formatter)
		if !ok {
			errs = append(errs, Encode(next))
			return errs
		}
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
