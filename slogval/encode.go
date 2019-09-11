package slogval

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"gitlab.com/c0b/go-ordered-json"
	"golang.org/x/xerrors"

	"go.coder.com/slog"
)

// VisitFunc is used to customize the representation of any field visited
// by the Encode.
type VisitFunc func(v interface{}, fn VisitFunc) (_ Value, ok bool)

// Encode encodes the interface to Value.
func Encode(v interface{}, visit VisitFunc) Value {
	if visit != nil {
		rv, ok := visit(v, visit)
		if ok {
			return rv
		}
	}

	switch v := v.(type) {
	case interface {
		LogValueJSON() interface{}
	}:
		return fromJSON(v.LogValueJSON(), visit)
	case slog.Value:
		return Encode(v.LogValue(), visit)
	case Value:
		return v
	case []slog.Field:
		var m Map
		for _, f := range v {
			val := Encode(f.LogValue(), visit)
			m = m.appendVal(f.LogKey(), val)
		}
		return m
	case json.Marshaler:
		return fromJSON(v, visit)
	case fmt.Stringer:
		return Encode(fmt.Sprintf("%+v", v), visit)
	case xerrors.Formatter:
		return extractXErrorChain(v, visit)
	case error:
		return Encode(fmt.Sprintf("%+v", v), visit)
	case string:
		return String(v)
	case bool:
		return Bool(v)
	}

	rv, ok := reflectEncode(v, visit)
	if ok {
		return rv
	}

	return Encode(fmt.Sprintf("%+v", v), visit)
}

func reflectEncode(v interface{}, visit VisitFunc) (Value, bool) {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Float32, reflect.Float64:
		return Float(rv.Float()), true
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		return Int(rv.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return Uint(rv.Uint()), true
	case reflect.Slice, reflect.Array:
		list := make(List, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			list[i] = Encode(rv.Index(i).Interface(), visit)
		}
		return list, true
	case reflect.Map:
		m := make(Map, 0, rv.Len())
		for _, k := range rv.MapKeys() {
			mv := rv.MapIndex(k)
			m = m.appendVal(fmt.Sprintf("%v", k), Encode(mv, visit))
		}
		m.sort()
		return m, true
	}
	return nil, false
}

func fromJSON(v interface{}, visit VisitFunc) Value {
	b, err := json.Marshal(map[string]interface{}{
		"val": v,
	})
	if err != nil {
		err = xerrors.Errorf("failed to marshal JSON: %w", err)
		return Encode(err, visit)
	}

	m := ordered.NewOrderedMap()
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	err = d.Decode(m)
	if err != nil {
		err = xerrors.Errorf("failed to unmarshal valid JSON: %w", err)
		return Encode(err, visit)
	}

	jsonVal := m.Get("val")
	return unmarshalJSONVal(jsonVal)
}

type wrapError struct {
	Msg string `json:"msg"`
	Fun string `json:"fun"`
	// file:line `json
	Loc string `json:"loc"`
}

func (e wrapError) Error() string {
	return fmt.Sprintf("wrap error: %+v", e)
}

func (e wrapError) LogValue() interface{} {
	return slog.JSON(e)
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

func extractXErrorChain(f xerrors.Formatter, visit VisitFunc) List {
	var l List

	for {
		p := &xerrorPrinter{}
		next := f.FormatError(p)

		l = append(l, Encode(p.e, visit))

		if next != nil {
			var ok bool
			f, ok = next.(xerrors.Formatter)
			if ok {
				continue
			}
			l = append(l, Encode(next, visit))
		}
		return l
	}
}

func panicf(f string, v ...interface{}) {
	f = "slogval: " + f
	s := fmt.Sprintf(f, v...)
	panic(s)
}

func unmarshalJSONVal(v interface{}) Value {
	switch v := v.(type) {
	case string:
		return String(v)
	case json.Number:
		i, err := v.Int64()
		if err == nil {
			return Int(i)
		}
		f, err := v.Float64()
		if err == nil {
			return Float(f)
		}
		return String(err.Error())
	case bool:
		return Bool(v)
	case []interface{}:
		l := make(List, 0, len(v))
		for _, v := range v {
			l = append(l, unmarshalJSONVal(v))
		}
		return l
	case *ordered.OrderedMap:
		var m Map
		it := v.EntriesIter()
		for {
			f, ok := it()
			if !ok {
				return m
			}
			m = m.appendVal(f.Key, unmarshalJSONVal(f.Value))
		}
	}
	panic("slogval: unexpected JSON type %T" + reflect.TypeOf(v).String())
}
