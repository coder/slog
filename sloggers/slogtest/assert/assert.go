// Package assert is a helper package for test assertions.
package assert

import (
	"testing"

	"cdr.dev/slog"
	"cdr.dev/slog/internal/assert"
	"cdr.dev/slog/sloggers/slogtest"
)

// Equal compares exp to act. If they are not equal, it will fatal the test
// with the passed msg, fields and also a diff of the differences.
func Equal(t testing.TB, exp, act interface{}, name string) {
	slog.Helper()
	equal(t, exp, act, name)
}

// Success calls Equal with exp = nil and act = err.
func Success(t testing.TB, err error, name string) {
	slog.Helper()
	if err != nil {
		slogtest.Fatal(t, "unexpected error",
			slog.F("name", name),
			slog.Error(err),
		)
	}
}

// True asserts that act is true.
func True(t testing.TB, act bool, name string) {
	slog.Helper()
	equal(t, true, act, name)
}

func equal(t testing.TB, exp, act interface{}, name string) {
	slog.Helper()
	if diff := assert.CmpDiff(exp, act); diff != "" {
		slogtest.Fatal(t, "equal assertion failed",
			slog.F("name", name),
			slog.F("diff", diff),
		)
	}
}
