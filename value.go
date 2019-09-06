package slog

import (
	"sort"
)

type fieldValue interface {
	isFieldValue()
}

type fieldString string

func (f fieldString) isFieldValue() {}

type fieldInt int64

func (f fieldInt) isFieldValue() {}

type fieldUint uint64

func (f fieldUint) isFieldValue() {}

type fieldFloat float64

func (f fieldFloat) isFieldValue() {}

type fieldBool bool

func (f fieldBool) isFieldValue() {}

type fieldMap []field

func (f fieldMap) isFieldValue() {}

type fieldList []fieldValue

func (f fieldList) isFieldValue() {}

func (f fieldMap) clone() fieldMap {
	f2 := make(fieldMap, len(f))
	copy(f2, f)
	return f2
}

func (f fieldMap) append(key string, val fieldValue) fieldMap {
	return append(f, field{
		key,
		val,
	})
}

func (f fieldMap) appendFields(f2 fieldMap) fieldMap {
	if len(f2) == 0 {
		return f
	}

	f = f.clone()
	return append(f, f2...)
}

// sort sorts the fields by name.
// Only used when the fields represent a map to ensure
// stable key order.
func (f fieldMap) sort() {
	sort.Slice(f, func(i, j int) bool {
		return f[i].name < f[j].name
	})
}
