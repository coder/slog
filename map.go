package slog

import (
	"bytes"
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"unicode"

	"golang.org/x/xerrors"
)

// Map represents an ordered map of fields.
type Map []Field

var _ json.Marshaler = Map(nil)

// MarshalJSON implements json.Marshaler.
func (m Map) MarshalJSON() ([]byte, error) {
	m = encodeInterface(m).(Map)

	b := &bytes.Buffer{}
	b.WriteString("{")
	for i, f := range m {
		fieldName, err := json.Marshal(f.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal field name: %w", err)
		}

		fieldValue, err := json.Marshal(f.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal field value: %w", err)
		}

		b.WriteString("\n")
		b.Write(fieldName)
		b.WriteString(":")
		b.Write(fieldValue)

		if i < len(m)-1 {
			b.WriteString(",")
		}
	}
	b.WriteString(`}`)

	return b.Bytes(), nil
}

func encode(rv reflect.Value, pure bool) interface{} {
	if pure {
		return encodeReflect(rv, true)
	}
	return encodeInterface(rv.Interface())
}

func encodeInterface(v interface{}) interface{} {
	if v == nil {
		return nil
	}

	switch v := v.(type) {
	case JSON:
		rv := reflect.ValueOf(v.V)
		if rv.Kind() != reflect.Struct {
			return encodeInterface(v.V)
		}
		m := make(Map, 0, rv.NumField())
		m = reflectStruct(m, rv, rv.Type(), true, false)
		return m
	case Map:
		m := make(Map, 0, len(v))
		for _, f := range v {
			m = append(m, F(f.Name, encodeInterface(f.Value)))
		}
		return m
	case Value:
		return encodeInterface(v.LogValue())
	case protoMessage:
		return encodeReflect(reflect.ValueOf(v), false)
	case encoding.TextMarshaler:
		return marshalText(v)
	case xerrors.Formatter:
		return extractXErrorChain(v)
	case error, fmt.Stringer:
		// Cannot use %+v here as if its not a xerrors.Formatter
		// then fmt will use its reflection encoder for the value
		// instead of v.String() or v.Error().
		return encodeInterface(fmt.Sprintf("%v", v))
	default:
		return encodeReflect(reflect.ValueOf(v), false)
	}
}

func marshalText(v encoding.TextMarshaler) interface{} {
	b, err := v.MarshalText()
	if err != nil {
		return encodeInterface(map[string]error{
			"marshalTextError": err,
		})
	}
	return encodeInterface(string(b))
}

func encodeReflect(rv reflect.Value, pure bool) interface{} {
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
		return rv.String()
	case reflect.Bool:
		return rv.Bool()
	case reflect.Float32, reflect.Float64:
		return rv.Float()
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		return rv.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rv.Uint()
	case reflect.Ptr:
		return encode(rv.Elem(), pure)
	case reflect.Interface:
		return encode(rv.Elem(), pure)
	case reflect.Slice, reflect.Array:
		// Ordered map.
		if typ == reflect.TypeOf(Map(nil)) {
			m := make(Map, 0, rv.Len())
			for i := 0; i < rv.Len(); i++ {
				f := rv.Index(i)
				key := f.FieldByName("Name").String()
				val := f.FieldByName("Value")
				m = append(m, F(key, encode(val, pure)))
			}
			return m
		}
		list := make([]interface{}, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			list[i] = encode(rv.Index(i), pure)
		}
		return list
	case reflect.Map:
		m := make(Map, 0, rv.Len())
		for _, k := range rv.MapKeys() {
			mv := rv.MapIndex(k)
			m = append(m, F(fmt.Sprintf("%v", k), encode(mv, pure)))
		}
		// Ensure stable key order.
		sort.Slice(m, func(i, j int) bool {
			return m[i].Name < m[j].Name
		})
		return m
	case reflect.Struct:
		m := make(Map, 0, typ.NumField())
		m = reflectStruct(m, rv, typ, false, pure)
		return m
	default:
		return fmt.Sprintf("%+v", rv)
	}
}

func reflectStruct(m Map, rv reflect.Value, structTyp reflect.Type, json, pure bool) Map {
	for i := 0; i < structTyp.NumField(); i++ {
		typ := structTyp.Field(i)
		rv := rv.Field(i)

		if implements(structTyp, (*protoMessage)(nil)) && strings.HasPrefix(typ.Name, "XXX_") {
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

		v := encode(rv, pure)
		if sm, ok := v.(Map); ok {
			m = append(m, sm...)
		} else {
			m = append(m, F(tagFieldName, v))
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

// Adapted from https://github.com/fatih/camelcase/blob/9db1b65eb38bb28986b93b521af1b7891ee1b04d/camelcase.go#L10
func snakecase(s string) string {
	const (
		lower = iota + 1
		upper
		digit
	)

	class := func(r rune) int {
		// https://golang.org/ref/spec#Identifiers
		switch {
		case unicode.IsLower(r):
			return lower
		case unicode.IsUpper(r):
			return upper
		default:
			return digit
		}
	}

	var res strings.Builder
	runes := []rune(s)
	var prevClass int
	for i, r := range runes {
		c := class(r)

		// If this rune is uppercase and the next one is lowercase,
		// then we have a new field.
		if i+1 < len(runes) && c == upper && class(runes[i+1]) == lower {
			// This indicates to the below code that this is a new field
			// and then on the next iteration, it ensures another field
			// is not detected due to the class change.
			prevClass = upper
			c = lower
		}

		// If the class has changed then we have a new field.
		if prevClass != c {
			if i > 0 {
				res.WriteRune('_')
			}
		}

		res.WriteRune(unicode.ToLower(r))

		prevClass = c
	}

	return res.String()
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
	return typ.Implements(reflect.TypeOf(v).Elem()) || reflect.PtrTo(typ).Implements(reflect.TypeOf(v).Elem())
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
	return JSON{e}
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
		panic(fmt.Sprintf("slog: unexpected String from xerrors.FormatError: %q", s))
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
func extractXErrorChain(f xerrors.Formatter) []interface{} {
	var errs []interface{}

	for {
		p := &xerrorPrinter{}
		next := f.FormatError(p)
		errs = append(errs, encodeInterface(p.e))

		if next == nil {
			return errs
		}

		var ok bool
		f, ok = next.(xerrors.Formatter)
		if !ok {
			errs = append(errs, encodeInterface(next))
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

type protoMessage interface {
	// Avoids the dependency on github.com/golang/protobuf/proto to be explicit here.
	ProtoMessage()
}
