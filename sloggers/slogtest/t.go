// Package slogtest contains the slogger for use
// with Go's testing package.
package slogtest // import "go.coder.com/slog/sloggers/slogtest"

import (
	"context"
	"log"
	"os"
	"testing"

	"go.coder.com/slog"
	"go.coder.com/slog/internal/humanfmt"
	"go.coder.com/slog/sloggers/sloghuman"
)

// Ensure all stdlib logs go through slog.
func init() {
	l := sloghuman.Make(os.Stderr)
	log.SetOutput(slog.Stdlib(context.Background(), l).Writer())
}

// Options represents the options for the logger returned
// by Make.
type Options struct {
	// IgnoreErrors causes the test logger to not fatal the test
	// on Fatal and not error the test on Error or Critical.
	IgnoreErrors bool
}

// Make creates a Logger that writes logs to tb in a human readable format.
func Make(tb testing.TB, opts *Options) slog.Logger {
	if opts == nil {
		opts = &Options{}
	}
	return slog.Make(testSink{
		tb:   tb,
		opts: opts,
	})
}

type testSink struct {
	tb     testing.TB
	opts   *Options
	stdlib bool
}

// The testing package logs to stdout and not stderr.
var stdoutColor = humanfmt.IsTTY(os.Stdout) || os.Getenv("FORCE_COLOR") != ""

func (ts testSink) LogEntry(ctx context.Context, ent slog.SinkEntry) error {
	s := humanfmt.Entry(ent, stdoutColor)

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
	return nil
}

func (ts testSink) Sync() error {
	return nil
}

var ctx = context.Background()

func l(t testing.TB) slog.Logger {
	return Make(t, nil)
}

// Debug logs the given msg and fields to t via t.Log at the debug level.
func Debug(t testing.TB, msg string, fields ...slog.F) {
	slog.Helper()
	l(t).Debug(ctx, msg, fields...)
}

// Info logs the given msg and fields to t via t.Log at the info level.
func Info(t testing.TB, msg string, fields ...slog.F) {
	slog.Helper()
	l(t).Info(ctx, msg, fields...)
}

// Error logs the given msg and fields to t via t.Error at the error level.
func Error(t testing.TB, msg string, fields ...slog.F) {
	slog.Helper()
	l(t).Error(ctx, msg, fields...)
}

// Fatal logs the given msg and fields to t via t.Fatal at the fatal level.
func Fatal(t testing.TB, msg string, fields ...slog.F) {
	slog.Helper()
	l(t).Fatal(ctx, msg, fields...)
}
