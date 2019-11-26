// Package assert is a helper package for test assertions.
package assert

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"

	"go.coder.com/slog"
	"go.coder.com/slog/sloggers/slogtest"
)

// Equal compares exp to act. If they are not equal, it will fatal the test
// with the passed msg, fields and also a diff of the differences.
func Equal(t testing.TB, exp, act interface{}, msg string, fields ...slog.F) {
	slog.Helper()
	equal(t, exp, act, msg, fields)
}

// Success calls Equal with exp = nil and act = err.
func Success(t testing.TB, err error, msg string, fields ...slog.F) {
	slog.Helper()
	if err != nil {
		fields = append(fields, slog.Error(err))
		slogtest.Fatal(t, "error: "+msg, fields...)
	}
}

// True asserts that act is true.
func True(t testing.TB, act bool, msg string, fields ...slog.F) {
	slog.Helper()
	equal(t, true, act, msg, fields)
}

func equal(t testing.TB, exp, act interface{}, msg string, fields slog.Map, opts ...cmp.Option) {
	slog.Helper()
	if diff := diff(exp, act); diff != "" {
		fields = append(fields, slog.F{"diff", diff})
		slogtest.Fatal(t, "equal assertion failed: "+msg, fields...)
	}
}

// diff returns a diff between exp and act.
// The empty string is returned if they are identical.
// See https://github.com/google/go-cmp/issues/40#issuecomment-328615283
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
