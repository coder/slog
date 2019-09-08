package slogval // import "go.coder.com/slog/slogval"

import (
	"sort"
)

type Value interface {
	// This returns the Value so that we do not need
	// to reconstruct the field ourselves as we cannot
	// access it directly without an accessor method
	// in case its on an unexported struct.
	isSlogCoreValue() Value
}

type Field struct {
	Name  string
	Value Value
}

type String string

func (f String) isSlogCoreValue() Value {
	return f
}

type Int int64

func (f Int) isSlogCoreValue() Value {
	return f
}

type Uint uint64

func (f Uint) isSlogCoreValue() Value {
	return f
}

type Float float64

func (f Float) isSlogCoreValue() Value {
	return f
}

type Bool bool

func (f Bool) isSlogCoreValue() Value {
	return f
}

type Map []Field

func (f Map) isSlogCoreValue() Value {
	return f
}

type List []Value

func (f List) isSlogCoreValue() Value {
	return f
}

func (f Map) Clone() Map {
	f2 := make(Map, len(f))
	copy(f2, f)
	return f2
}

// TODO make sure this is used correctly.
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
