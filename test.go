package testlog

import (
	"context"
	"fmt"
	"testing"

	"go.coder.com/m/lib/log"
	"go.coder.com/m/lib/log/internal/core"
)

var globalCtx = func() context.Context {
	ctx := context.Background()
	ctx = core.WithSkip(ctx, 1)
	return ctx
}()

// Debug logs the given msg and fields to t via t.Log at the debug level.
func Debug(t *testing.T, msg string, fields log.F) {
	t.Helper()
	Make(t).Debug(globalCtx, msg, fields)
}

// Info logs the given msg and fields to t via t.Log at the info level.
func Info(t *testing.T, msg string, fields log.F) {
	t.Helper()
	Make(t).Info(globalCtx, msg, fields)
}

// Error logs the given msg and fields to t via t.Error at the error level.
func Error(t *testing.T, msg string, fields log.F) {
	t.Helper()
	Make(t).Error(globalCtx, msg, fields)
}

// Fatal logs the given msg and fields to t via t.Fatal at the fatal level.
func Fatal(t *testing.T, msg string, fields log.F) {
	t.Helper()
	Make(t).Fatal(globalCtx, msg, fields)
}

// Inspect is useful for one off debug statements.
func Inspect(t *testing.T, v ...interface{}) {
	t.Helper()
	core.Inspect(globalCtx, Make(t), v)
}

// Option describes an option for the test logger.
type Option func(testLogger) testLogger

// IgnoreError makes the logger write
// all logs via t.Log and never t.Error or t.Fatal.
//
// Note: With this option there is no reasonable way
// to handle Fatal logs and so the library panics.
func IgnoreError() Option {
	return func(tl testLogger) testLogger {
		tl.ignoreError = true
		return tl
	}
}

// Test creates a logger for a test.
//
// By default, if the logger receives an Error log
// or a Fatal log, then the log will be written via
// t.Error or t.Fatal respectively. If you want to disable
// this behaviour, use the IgnoreError option.
//
// Note: using IgnoreError means you cannot have
// Fatal logs. See the docs on it.
func Make(t testing.TB, opts ...Option) log.Logger {
	tl := testLogger{
		t,
		false,
		core.Logger{},
	}
	for _, opt := range opts {
		tl = opt(tl)
	}
	return tl
}

type testLogger struct {
	tb          testing.TB
	ignoreError bool

	l core.Logger
}

func (tl testLogger) Debug(ctx context.Context, msg string, fields log.F) {
	tl.tb.Helper()
	tl.log(ctx, core.Debug, msg, fields)
}

func (tl testLogger) Info(ctx context.Context, msg string, fields log.F) {
	tl.tb.Helper()
	tl.log(ctx, core.Info, msg, fields)
}

func (tl testLogger) Warn(ctx context.Context, msg string, fields log.F) {
	tl.tb.Helper()
	tl.log(ctx, core.Warn, msg, fields)
}

func (tl testLogger) Error(ctx context.Context, msg string, fields log.F) {
	tl.tb.Helper()
	tl.log(ctx, core.Error, msg, fields)
}

func (tl testLogger) Critical(ctx context.Context, msg string, fields log.F) {
	tl.tb.Helper()
	tl.log(ctx, core.Critical, msg, fields)
}

func (tl testLogger) Fatal(ctx context.Context, msg string, fields log.F) {
	tl.tb.Helper()
	tl.log(ctx, core.Fatal, msg, fields)
}

func (tl testLogger) With(fields log.F) log.Logger {
	tl.l = tl.l.With(fields)
	return tl
}

func (tl testLogger) log(ctx context.Context, sev core.Severity, msg string, fields log.F) {
	tl.tb.Helper()

	ent := tl.l.Make(ctx, core.EntryConfig{
		Sev:    sev,
		Msg:    msg,
		Fields: fields,
		Skip:   2,
	})
	if !core.IsStdlib(ctx) {
		// We do not want to print the file or line number ourselves.
		// The testing framework handles it for us.
		// But we do want the function name.
		// However, if the test package is being used with the stdlib log adapter, then we do want
		// the line/file number because we cannot put t.Helper calls in stdlib log.
		ent.Loc.File = ""
		ent.Loc.Line = 0
	}

	tl.write(ent)
}

func (tl testLogger) write(ent core.Entry) {
	tl.tb.Helper()

	s := ent.String()

	switch ent.Sev {
	case core.Debug, core.Info, core.Warn:
		tl.tb.Log(s)
	case core.Error, core.Critical:
		if tl.ignoreError {
			tl.tb.Log(s)
		} else {
			tl.tb.Error(s)
		}
	case core.Fatal:
		if tl.ignoreError {
			panicf("cannot fatal in tests when IgnoreError option is set")
		}
		tl.tb.Fatal(s)
	}
}

// See Stdlib in the log package for how this is used.
func (tl testLogger) TestLogger() testing.TB {
	return tl.tb
}

func panicf(f string, v ...interface{}) {
	f = "testlog: " + f
	s := fmt.Sprintf(f, v...)
	panic(s)
}
