// Package slogtest contains the slogger for use
// with Go's testing package.
package slogtest // import "go.coder.com/slog/sloggers/slogtest"

import (
	"context"
	"os"
	"testing"

	"go.coder.com/slog"
	"go.coder.com/slog/internal/humanfmt"
)

// TestOptions represents the options for the logger returned
// by Make.
type TestOptions struct {
	// IgnoreErrors causes the test logger to not fatal the test
	// on Fatal and not error the test on Error or Critical.
	IgnoreErrors bool
}

// Make creates a Logger that writes logs to tb in a human readable format.
func Make(tb testing.TB, opts *TestOptions) slog.Logger {
	if opts == nil {
		opts = &TestOptions{}
	}
	return slog.Make(testSink{
		tb:   tb,
		opts: opts,
	})
}

type testSink struct {
	tb     testing.TB
	opts   *TestOptions
	stdlib bool
}

var stderrColor = humanfmt.IsTTY(os.Stderr)

func (ts testSink) LogEntry(ctx context.Context, ent slog.Entry) {
	s := humanfmt.Entry(ent, stderrColor)

	switch ent.Level {
	case slog.LevelDebug, slog.LevelInfo, slog.LevelWarn:
		ts.tb.Log(s)
	case slog.LevelError, slog.LevelCritical:
		if ts.opts.IgnoreErrors {
			ts.tb.Log(s)
		} else {
			ts.tb.Error(s)
		}
	case slog.LevelFatal:
		if ts.opts.IgnoreErrors {
			panic("slogtest: cannot fatal in tests when IgnoreErrors option is set")
		}
		ts.tb.Fatal(s)
	}
}

var ctx = context.Background()

// Debug logs the given msg and fields to t via t.Log at the debug level.
func Debug(t testing.TB, msg string, fields ...slog.Field) {
	l := Make(t, nil)
	l.Helper()
	l.Debug(ctx, msg, fields...)
}

// Info logs the given msg and fields to t via t.Log at the info level.
func Info(t testing.TB, msg string, fields ...slog.Field) {
	l := Make(t, nil)
	l.Helper()
	l.Info(ctx, msg, fields...)
}

// Error logs the given msg and fields to t via t.Error at the error level.
func Error(t testing.TB, msg string, fields ...slog.Field) {
	l := Make(t, nil)
	l.Helper()
	l.Error(ctx, msg, fields...)
}

// Fatal logs the given msg and fields to t via t.Fatal at the fatal level.
func Fatal(t testing.TB, msg string, fields ...slog.Field) {
	l := Make(t, nil)
	l.Helper()
	l.Fatal(ctx, msg, fields...)
}
