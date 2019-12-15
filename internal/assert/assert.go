// Package assert contains helpers for test assertions.
package assert

import (
	"reflect"
	"testing"
)

// Equal asserts exp == act.
func Equal(t testing.TB, name string, exp, act interface{}) {
	t.Helper()
	if !reflect.DeepEqual(exp, act) {
		t.Fatalf("unexpected %v: exp: %q but got %q", name, exp, act)
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
