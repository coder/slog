package slogcore

import (
	"sort"
)

type Value interface {
	isFieldValue()
}

// field represents a log field.
type Field struct {
	Name  string
	Value Value
}

type String string

func (f String) isFieldValue() {}

type Int int64

func (f Int) isFieldValue() {}

type Uint uint64

func (f Uint) isFieldValue() {}

type Float float64

func (f Float) isFieldValue() {}

type Bool bool

func (f Bool) isFieldValue() {}

type Map []Field

func (f Map) isFieldValue() {}

type List []Value

func (f List) isFieldValue() {}

func (f Map) Clone() Map {
	f2 := make(Map, len(f))
	copy(f2, f)
	return f2
}

func (f Map) Append(key string, val Value) Map {
	return append(f, Field{
		key,
		val,
	})
}

func (f Map) AppendFields(f2 Map) Map {
	if len(f2) == 0 {
		return f
	}

	f = f.Clone()
	return append(f, f2...)
}

// sort sorts the fields by name.
// Only used when the fields represent a map to ensure
// stable key order.
func (f Map) Sort() {
	sort.Slice(f, func(i, j int) bool {
		return f[i].Name < f[j].Name
	})
}
