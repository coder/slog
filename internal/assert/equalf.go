// Package assert is a helper package for asserting equality and
// generating diffs in tests.
package assert

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// Equalf compares exp to act. If they are not equal, it will fatal the test
// with the passed msg, args and also a diff of the differences.
func Equalf(t *testing.T, exp, act interface{}, msg string, v ...interface{}) {
	if diff := diff(exp, act); diff != "" {
		t.Fatalf(msg+": %v", append(v, diff)...)
	}
}

// diff returns a diff between exp and act.
// The empty string is returned if they are identical.
// See https://github.com/google/go-cmp/issues/40#issuecomment-328615283
// Copied from https://github.com/nhooyr/websocket/blob/1b874731eab56c69c8bb3ebf8a029020c7863fc9/cmp_test.go
func diff(exp, act interface{}) string {
	return cmp.Diff(exp, act, deepAllowUnexported(exp, act))
}

func deepAllowUnexported(vs ...interface{}) cmp.Option {
	m := make(map[reflect.Type]struct{})
	for _, v := range vs {
		structTypes(reflect.ValueOf(v), m)
	}
	var typs []interface{}
	for t := range m {
		typs = append(typs, reflect.New(t).Elem().Interface())
	}
	return cmp.AllowUnexported(typs...)
}

func structTypes(v reflect.Value, m map[reflect.Type]struct{}) {
	if !v.IsValid() {
		return
	}
	switch v.Kind() {
	case reflect.Ptr:
		if !v.IsNil() {
			structTypes(v.Elem(), m)
		}
	case reflect.Interface:
		if !v.IsNil() {
			structTypes(v.Elem(), m)
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			structTypes(v.Index(i), m)
		}
	case reflect.Map:
		for _, k := range v.MapKeys() {
			structTypes(v.MapIndex(k), m)
		}
	case reflect.Struct:
		m[v.Type()] = struct{}{}
		for i := 0; i < v.NumField(); i++ {
			structTypes(v.Field(i), m)
		}
	}
}
