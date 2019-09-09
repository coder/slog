// Package slogval is used by the default sloggers to take a []slog.Field
// and convert it into a easily marshable slogval.Value.
package slogval // import "go.coder.com/slog/slogval"

import (
	"bytes"
	"encoding/json"
	"sort"

	"golang.org/x/xerrors"
)

// Value represents a primitive value for structured logging.
type Value interface {
	// This returns the Value so that we do not need
	// to reconstruct the field ourselves as we cannot
	// access it directly without an accessor method
	// in case its on an unexported struct.
	isSlogCoreValue() Value
}

// Field represents a field in the Map.
type Field struct {
	Name  string
	Value Value
}

// String represents a string.
type String string

func (f String) isSlogCoreValue() Value {
	return f
}

// Int represents an integer.
type Int int64

func (f Int) isSlogCoreValue() Value {
	return f
}

// Uint represents an unsigned integer.
type Uint uint64

func (f Uint) isSlogCoreValue() Value {
	return f
}

// Float represents a floating point number.
type Float float64

func (f Float) isSlogCoreValue() Value {
	return f
}

// Bool represents a boolean.
type Bool bool

func (f Bool) isSlogCoreValue() Value {
	return f
}

// Map represents a ordered map.
type Map []Field

func (m Map) isSlogCoreValue() Value {
	return m
}

// List represents a list of values.
type List []Value

func (f List) isSlogCoreValue() Value {
	return f
}

// appendVal appends an entry with the given key
// and val to the map.
func (m Map) appendVal(key string, val Value) Map {
	return append(m, Field{
		key,
		val,
	})
}

// sort sorts the fields by name.
// Only used when the fields represent a Go map to ensure
// stable key order.
func (m Map) sort() {
	sort.Slice(m, func(i, j int) bool {
		return m[i].Name < m[j].Name
	})
}

var _ json.Marshaler = Map(nil)

// MarshalJSON implements json.Marshaler.
func (m Map) MarshalJSON() ([]byte, error) {
	b := &bytes.Buffer{}
	b.WriteString("{")
	for i, f := range m {
		fieldName, err := json.Marshal(f.Name)
		if err != nil {
			return nil, xerrors.Errorf("failed to marshal field name: %w", err)
		}

		fieldValue, err := json.Marshal(f.Value)
		if err != nil {
			return nil, xerrors.Errorf("failed to marshal field value: %w", err)
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
