// Package assert contains helpers for test assertions.
package assert

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// Diff returns a diff between exp and act.
func Diff(exp, act interface{}, opts ...cmp.Option) string {
	opts = append(opts, cmpopts.EquateErrors(), cmp.Exporter(func(r reflect.Type) bool {
		return true
	}))
	return cmp.Diff(exp, act, opts...)
}

// Equal asserts exp == act.
func Equal(t testing.TB, name string, exp, act interface{}) {
	t.Helper()
	if diff := Diff(exp, act); diff != "" {
		t.Fatalf(`unexpected %v: diff:
%v`, name, diff)
	}
}

// Success asserts err == nil.
func Success(t testing.TB, name string, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error for %v: %+v", name, err)
	}
}

// Error asserts exp != nil.
func Error(t testing.TB, name string, err error) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error from %v", name)
	}
}

// True asserts true == act.
func True(t testing.TB, name string, act bool) {
	t.Helper()
	Equal(t, name, true, act)
}

// False asserts false == act.
func False(t testing.TB, name string, act bool) {
	t.Helper()
	Equal(t, name, false, act)
}

// Len asserts n == len(a).
func Len(t testing.TB, name string, n int, a interface{}) {
	t.Helper()
	act := reflect.ValueOf(a).Len()
	if n != act {
		t.Fatalf("expected len(%v) == %v but got %v", name, n, act)
	}
}
