// Package assert is a helper package for test assertions.
//
// On failure, every assertion will fatal the test.
//
// The name parameter is available in each assertion for easier debugging.
package assert // import "cdr.dev/slog/sloggers/slogtest/assert"

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"cdr.dev/slog"
	"cdr.dev/slog/internal/assert"
	"cdr.dev/slog/sloggers/slogtest"
)

// Equal asserts exp == act.
//
// If they are not equal, it will fatal the test with a diff of the
// two objects.
//
// Errors will be compared with errors.Is.
func Equal(t testing.TB, name string, exp, act interface{}, opts ...cmp.Option) {
	slog.Helper()

	if diff := assert.Diff(exp, act, opts...); diff != "" {
		slogtest.Fatal(t, "unexpected value",
			slog.F("name", name),
			slog.F("diff", diff),
		)
	}
}

// Success asserts err == nil.
func Success(t testing.TB, name string, err error) {
	slog.Helper()
	if err != nil {
		slogtest.Fatal(t, "unexpected error",
			slog.F("name", name),
			slog.Error(err),
		)
	}
}

// True asserts act == true.
func True(t testing.TB, name string, act bool) {
	slog.Helper()
	Equal(t, name, true, act)
}

// False asserts act == false.
func False(t testing.TB, name string, act bool) {
	slog.Helper()
	Equal(t, name, false, act)
}

// Error asserts err != nil.
func Error(t testing.TB, name string, err error) {
	slog.Helper()
	if err == nil {
		slogtest.Fatal(t, "expected error",
			slog.F("name", name),
		)
	}
}

// ErrorContains asserts err != nil and err.Error() contains sub.
//
// The match will be case insensitive.
func ErrorContains(t testing.TB, name string, err error, sub string) {
	slog.Helper()

	Error(t, name, err)

	errs := err.Error()
	if !stringContainsFold(errs, sub) {
		slogtest.Fatal(t, "unexpected error string",
			slog.F("name", name),
			slog.F("error_string", errs),
			slog.F("expected_contains", sub),
		)
	}
}

func stringContainsFold(errs, sub string) bool {
	errs = strings.ToLower(errs)
	sub = strings.ToLower(sub)

	return strings.Contains(errs, sub)

}
