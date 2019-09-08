package slogtest

import (
	"context"
	"os"
	"testing"

	"go.coder.com/slog"
	"go.coder.com/slog/internal/humanfmt"
	"go.coder.com/slog/internal/stdlibctx"
)

// TestOptions represents the options for the logger returned
// by Test.
type TestOptions struct {
	*slog.Options

	// IgnoreErrors causes the test logger to not fatal the test
	// on Fatal and not error the test on Error or Critical.
	IgnoreErrors bool
}

// Test creates a Logger that writes logs to tb.
func Make(tb testing.TB, opts *TestOptions) slog.Logger {
	if opts == nil {
		opts = &TestOptions{}
	}
	return slog.Make(testSink{
		tb:   tb,
		opts: opts,
	}, opts.Options)
}

type testSink struct {
	tb   testing.TB
	opts *TestOptions
}

func (ts testSink) XXX_slogTestingHelper() func() {
	return ts.tb.Helper
}

var stderrColor = humanfmt.IsTTY(os.Stderr)

func (ts testSink) LogEntry(ctx context.Context, ent slog.Entry) {
	ts.tb.Helper()

	if !stdlibctx.From(ctx) {
		// We do not want to print the file or line number ourselves.
		// The testing framework handles it for us.
		// But we do want the function name.
		// However, if the test package is being used with the stdlib log adapter, then we do want
		// the line/file number because we cannot put t.Helper calls in stdlib log.
		ent.File = ""
		ent.Line = 0
	}

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
