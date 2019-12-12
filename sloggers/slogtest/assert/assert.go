// Package assert is a helper package for test assertions.
//
// On failure, every assertion will fatal the test.
//
// The name parameter is available in each assertion for easier debugging.
package assert // import "cdr.dev/slog/sloggers/slogtest/assert"

import (
	"errors"
	"testing"

	"cdr.dev/slog"
	"cdr.dev/slog/internal/assert"
	"cdr.dev/slog/sloggers/slogtest"
)

// Equal asserts exp == act.
//
// If they are not equal, it will fatal the test with a diff of the
// two objects.
//
// If act is an error it will be unwrapped.
func Equal(t testing.TB, exp, act interface{}, name string) {
	slog.Helper()

	if err, ok := act.(error); ok {
		act = unwrapErr(err)
	}

	if diff := assert.CmpDiff(exp, act); diff != "" {
		slogtest.Fatal(t, "unexpected value",
			slog.F("name", name),
			slog.F("diff", diff),
		)
	}
}

// Success asserts err == nil.
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
func True(t testing.TB, act bool, name string) {
	slog.Helper()
	Equal(t, true, act, name)
}

// Error asserts err != nil.
func Error(t testing.TB, err error, name string) {
	slog.Helper()
	if err == nil {
		slogtest.Fatal(t, "expected error",
			slog.F("name", name),
		)
	}
}

func unwrapErr(err error) error {
	uerr := errors.Unwrap(err)
	for uerr != nil {
		err = uerr
		uerr = errors.Unwrap(uerr)
	}
	return err
}
