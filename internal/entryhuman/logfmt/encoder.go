package logfmt

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
	"unicode"
)

type Encoder struct {
	w           io.Writer
	FormatKey   func(key string) string
	// FormatPrimitiveValue is used to format primitive values (strings, ints,
	// floats, etc). It is not used for arrays or objects.
	FormatPrimitiveValue func(value interface{}) string
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		FormatKey: func(key string) string { return key },
		FormatPrimitiveValue: func(value interface{}) string { return fmt.Sprintf("%+v", value) },
		w: w,
	}
}

func isPrimitive(typ reflect.Type) bool {
	switch typ.Kind() {
	case reflect.Bool:
		return true
	case reflect.String:
		return true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	case reflect.Float32, reflect.Float64:
		return true
	case reflect.Complex64, reflect.Complex128:
		return true
	default:
		return false
	}
}

// Encode encodes the given message to the writer. For flat objects, the
// output resembles key=value pairs. For nested objects, a surrounding { } is
// used. For arrays, a surrounding [ ] is used.
func (e *Encoder) Encode(m interface{}) error {
	typ := reflect.TypeOf(m)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	if isPrimitive(typ) {
		e.w.Write([]byte(e.FormatPrimitiveValue(m)))
		return nil
	}

	switch typ.Kind() {
	case reflect.Struct:
		v := reflect.ValueOf(m)
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			value := v.Field(i)
			if !value.CanInterface() {
				continue
			}
			if value.IsZero() {
				continue
			}
			if field.Anonymous {
				if err := e.Encode(value.Interface()); err != nil {
					return err
				}
				continue
			}
			if e.FormatKey != nil {
				e.w.Write([]byte(e.FormatKey(field.Name)))
			} else {
				e.w.Write([]byte(field.Name))
			}
			e.w.Write([]byte("="))
			if e.FormatPrimitiveValue != nil {
				e.w.Write([]byte(e.FormatPrimitiveValue(value.Interface())))
			} else {
				e.w.Write([]byte(formatValue(value.Interface())))
			}
	default:
		return fmt.Errorf("unsupported type %T", m)
	}
}

// quotes quotes a string so that it is suitable
// as a key for a map or in general some output that
// cannot span multiple lines or have weird characters.
func quote(key string) string {
	// strconv.Quote does not quote an empty string so we need this.
	if key == "" {
		return `""`
	}

	var hasSpace bool
	for _, r := range key {
		if unicode.IsSpace(r) {
			hasSpace = true
			break
		}
	}
	quoted := strconv.Quote(key)
	// If the key doesn't need to be quoted, don't quote it.
	// We do not use strconv.CanBackquote because it doesn't
	// account tabs.
	if !hasSpace && quoted[1:len(quoted)-1] == key {
		return key
	}
	return quoted
}
