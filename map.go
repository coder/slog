package slog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"golang.org/x/xerrors"
)

// Map represents an ordered map of fields.
type Map []Field

// SlogValue implements Value.
func (m Map) SlogValue() interface{} {
	return ForceJSON(m)
}

var _ json.Marshaler = Map(nil)

// MarshalJSON implements json.Marshaler.
//
// It is guaranteed to return a nil error.
// Any error marshalling a field will become the field's value.
//
// Every field value is encoded with the following process:
//
// 1. slog.Value is handled to allow any type to replace its representation for logging.
//
// 2. xerrors.Formatter is handled.
//
// 3. error and fmt.Stringer are handled.
//
// 4. slices and arrays are handled to go through the encode function for every value.
//
// 5. json.Marshal is invoked as the default case.
func (m Map) MarshalJSON() ([]byte, error) {
	b := &bytes.Buffer{}
	b.WriteByte('{')
	for i, f := range m {
		b.WriteByte('\n')
		b.Write(encode(f.Name))
		b.WriteByte(':')
		b.Write(encode(f.Value))

		if i < len(m)-1 {
			b.WriteByte(',')
		}
	}
	b.WriteByte('}')

	return b.Bytes(), nil
}

// ForceJSON ensures the value is logged via json.Marshal even
// if it implements fmt.Stringer or error.
func ForceJSON(v interface{}) interface{} {
	return jsonVal{v: v}
}

type jsonVal struct {
	v interface{}
}

var _ json.Marshaler = jsonVal{}

// MarshalJSON implements json.Marshaler.
func (v jsonVal) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.v)
}

func marshalList(rv reflect.Value) []byte {
	b := &bytes.Buffer{}
	b.WriteByte('[')
	for i := 0; i < rv.Len(); i++ {
		b.WriteByte('\n')
		b.Write(encode(rv.Index(i).Interface()))

		if i < rv.Len()-1 {
			b.WriteByte(',')
		}
	}
	b.WriteByte(']')

	return b.Bytes()
}

func encode(v interface{}) []byte {
	switch v := v.(type) {
	case Value:
		return encode(v.SlogValue())
	case xerrors.Formatter:
		return encode(errorChain(v))
	case error, fmt.Stringer:
		return encode(fmt.Sprint(v))
	default:
		rv := reflect.Indirect(reflect.ValueOf(v))
		if rv.IsValid() {
			switch rv.Type().Kind() {
			case reflect.Slice:
				if rv.IsNil() {
					break
				}
				fallthrough
			case reflect.Array:
				return marshalList(rv)
			}
		}

		b, err := json.Marshal(v)
		if err != nil {
			return encode(M(
				Error(xerrors.Errorf("failed to marshal to JSON: %w", err)),
				F("type", reflect.TypeOf(v)),
				F("value", fmt.Sprintf("%+v", v)),
			))
		}
		return b
	}
}

func errorChain(f xerrors.Formatter) []interface{} {
	var errs []interface{}

	next := error(f)
	for {
		f, ok := next.(xerrors.Formatter)
		if !ok {
			errs = append(errs, next)
			return errs
		}

		p := &xerrorPrinter{}
		next = f.FormatError(p)
		errs = append(errs, p.e)
	}
}

type wrapError struct {
	Msg string `json:"msg"`
	Fun string `json:"fun"`
	// file:line
	Loc string `json:"loc"`
}

type xerrorPrinter struct {
	e wrapError
}

func (p *xerrorPrinter) Print(v ...interface{}) {
	s := fmt.Sprint(v...)
	p.write(s)
}

func (p *xerrorPrinter) Printf(f string, v ...interface{}) {
	s := fmt.Sprintf(f, v...)
	p.write(s)
}

func (p *xerrorPrinter) Detail() bool {
	return true
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
	}
}

func (m Map) append(m2 Map) Map {
	m3 := make(Map, 0, len(m)+len(m2))
	m3 = append(m3, m...)
	m3 = append(m3, m2...)
	return m3
}
