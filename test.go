package slog

import (
	"context"
	"os"
	"testing"

	"go.coder.com/slog/internal/humanfmt"
	"go.coder.com/slog/slogcore"
)

// TestOptions represents the options for the logger returned
// by Test.
type TestOptions struct {
	// IgnoreErrors causes the test logger to not fatal the test
	// on Fatal and not error the test on Error or Critical.
	IgnoreErrors bool
}

// Test creates a Logger that writes logs to tb.
func Test(tb testing.TB, opts *TestOptions) Logger {
	if opts == nil {
		opts = &TestOptions{}
	}
	return testLogger{
		tb:   tb,
		opts: opts,
	}
}

type testLogger struct {
	tb   testing.TB
	opts *TestOptions

	stdlib bool
	p      parsedFields
}

func (tl testLogger) Debug(ctx context.Context, msg string, fields ...Field) {
	tl.tb.Helper()
	tl.log(ctx, slogcore.Debug, msg, fields)
}

func (tl testLogger) Info(ctx context.Context, msg string, fields ...Field) {
	tl.tb.Helper()
	tl.log(ctx, slogcore.Info, msg, fields)
}

func (tl testLogger) Warn(ctx context.Context, msg string, fields ...Field) {
	tl.tb.Helper()
	tl.log(ctx, slogcore.Warn, msg, fields)
}

func (tl testLogger) Error(ctx context.Context, msg string, fields ...Field) {
	tl.tb.Helper()
	tl.log(ctx, slogcore.Error, msg, fields)
}

func (tl testLogger) Critical(ctx context.Context, msg string, fields ...Field) {
	tl.tb.Helper()
	tl.log(ctx, slogcore.Critical, msg, fields)
}

func (tl testLogger) Fatal(ctx context.Context, msg string, fields ...Field) {
	tl.tb.Helper()
	tl.log(ctx, slogcore.Fatal, msg, fields)
}

func (tl testLogger) With(fields ...Field) Logger {
	tl.p = tl.p.withFields(fields)
	return tl
}

func (tl testLogger) log(ctx context.Context, level slogcore.Level, msg string, fields []Field) {
	tl.tb.Helper()

	ent := tl.p.entry(ctx, entryParams{
		level:  level,
		msg:    msg,
		fields: fields,
		skip:   2,
	})
	if !tl.stdlib {
		// We do not want to print the file or line number ourselves.
		// The testing framework handles it for us.
		// But we do want the function name.
		// However, if the test package is being used with the stdlib log adapter, then we do want
		// the line/file number because we cannot put t.Helper calls in stdlib log.
		ent.File = ""
		ent.Line = 0
	}

	tl.write(ent)
}

var stderrColor = isTTY(os.Stderr)

func (tl testLogger) write(ent slogcore.Entry) {
	tl.tb.Helper()

	s := humanfmt.Entry(ent, stderrColor)

	switch ent.Level {
	case slogcore.Debug, slogcore.Info, slogcore.Warn:
		tl.tb.Log(s)
	case slogcore.Error, slogcore.Critical:
		if tl.opts.IgnoreErrors {
			tl.tb.Log(s)
		} else {
			tl.tb.Error(s)
		}
	case slogcore.Fatal:
		if tl.opts.IgnoreErrors {
			panicf("cannot fatal in tests when IgnoreErrors option is set")
		}
		tl.tb.Fatal(s)
	}
}
