// Package assert is a helper package for test assertions.
package assert // import "cdr.dev/slog/sloggers/slogtest/assert"

import (
	"testing"

	"cdr.dev/slog"
	"cdr.dev/slog/internal/assert"
	"cdr.dev/slog/sloggers/slogtest"
)

// Equal asserts exp == act.
//
// If they are not equal, it will fatal the test with a diff of the
// two objects.
func Equal(t testing.TB, exp, act interface{}, name string) {
	slog.Helper()
	if diff := assert.CmpDiff(exp, act); diff != "" {
		slogtest.Fatal(t, "unexpected value",
			slog.F("name", name),
			slog.F("diff", diff),
		)
	}
}

// Success asserts err == nil.
//
// If err isn't nil, it will fatal the test with the error.
func Success(t testing.TB, err error, name string) {
	slog.Helper()
	if err != nil {
		slogtest.Fatal(t, "unexpected error",
			slog.F("name", name),
			slog.Error(err),
		)
	}
}

// True asserts act == true.
//
// If act isn't true, it will fatal the test.
func True(t testing.TB, act bool, name string) {
	slog.Helper()
	Equal(t, true, act, name)
}
