// Package assert contains helpers for test assertions.
package assert

import (
	"strings"
	"testing"
)

// Equal asserts exp == act.
func Equal(t testing.TB, exp, act interface{}, name string) {
	t.Helper()
	diff := CmpDiff(exp, act)
	if diff != "" {
		t.Fatalf("unexpected %v: %v", name, diff)
	}
}

// NotEqual asserts exp != act.
func NotEqual(t testing.TB, exp, act interface{}, name string) {
	t.Helper()
	if CmpDiff(exp, act) == "" {
		t.Fatalf("expected different %v: %+v", name, act)
	}
}

// Success asserts exp == nil.
func Success(t testing.TB, err error, name string) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error for %v: %+v", name, err)
	}
}

// Error asserts exp != nil.
func Error(t testing.TB, err error, name string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error from %v", name)
	}
}

// ErrorContains asserts the error string from err contains sub.
func ErrorContains(t testing.TB, err error, sub, name string) {
	t.Helper()
	Error(t, err, name)
	errs := err.Error()
	if !strings.Contains(errs, sub) {
		t.Fatalf("error string %q from %v does not contain %q", errs, name, sub)
	}
}

// True asserts true == act.
func True(t testing.TB, act interface{}, name string) {
	t.Helper()
	Equal(t, true, act, name)
}
