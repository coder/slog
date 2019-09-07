// Package testslog is a helper around slog.Test().
package testslog // import "go.coder.com/slog/testslog"

import (
	"context"
	"testing"

	"go.coder.com/slog"
)

var ctx = context.Background()

// Debug logs the given msg and fields to t via t.Log at the debug level.
func Debug(t testing.TB, msg string, fields ...slog.Field) {
	t.Helper()
	slog.Test(t, nil).Debug(ctx, msg, fields...)
}

// Info logs the given msg and fields to t via t.Log at the info level.
func Info(t testing.TB, msg string, fields ...slog.Field) {
	t.Helper()
	slog.Test(t, nil).Info(ctx, msg, fields...)
}

// Error logs the given msg and fields to t via t.Error at the error level.
func Error(t testing.TB, msg string, fields ...slog.Field) {
	t.Helper()
	slog.Test(t, nil).Error(ctx, msg, fields...)
}

// Fatal logs the given msg and fields to t via t.Fatal at the fatal level.
func Fatal(t testing.TB, msg string, fields ...slog.Field) {
	t.Helper()
	slog.Test(t, nil).Fatal(ctx, msg, fields...)
}
