package slog_test

import (
	"context"
	"io"
	"runtime"
	"testing"

	"go.opencensus.io/trace"

	"cdr.dev/slog"
	"cdr.dev/slog/internal/assert"
)

var _, slogTestFile, _, _ = runtime.Caller(0)

type fakeSink struct {
	entries     []slog.SinkEntry
	logEntryErr error

	synced  bool
	syncErr error
}

func (s *fakeSink) LogEntry(_ context.Context, e slog.SinkEntry) error {
	s.entries = append(s.entries, e)
	return s.logEntryErr
}

func (s *fakeSink) Sync() error {
	s.synced = true
	return s.syncErr
}

var bg = context.Background()

func TestLogger(t *testing.T) {
	t.Parallel()

	t.Run("basic", func(t *testing.T) {
		t.Parallel()

		s1 := &fakeSink{}
		s2 := &fakeSink{}
		l := slog.Tee(slog.Make(s1), slog.Make(s2))

		l.SetLevel(slog.LevelError)

		l.Info(bg, "wow", slog.Error(io.EOF))
		l.Error(bg, "meow", slog.Error(io.ErrUnexpectedEOF))

		assert.True(t, s1.synced, "synced")
		assert.Equal(t, 1, len(s1.entries), "len(entries)")

		assert.Equal(t, s1, s2, "sinks")
	})

	t.Run("helper", func(t *testing.T) {
		t.Parallel()

		s := &fakeSink{}
		l := slog.Make(s)
		h := func(ctx context.Context) {
			slog.Helper()
			l.Debug(ctx, "logging in helper")
		}

		ctx := slog.Context(bg, slog.F(
			"ctx", 1024),
		)
		h(ctx)

		assert.Equal(t, slog.SinkEntry{
			Time: s.entries[0].Time,

			Level:   slog.LevelDebug,
			Message: "logging in helper",

			File: slogTestFile,
			Func: "cdr.dev/slog_test.TestLogger.func2",
			Line: 71,

			Fields: slog.M(
				slog.F("ctx", 1024),
			),
		}, s.entries[0], "entry")
	})

	t.Run("entry", func(t *testing.T) {
		t.Parallel()

		s := &fakeSink{}
		l := slog.Make(s)
		l = l.Named("hello")
		l = l.Named("hello2")

		ctx, span := trace.StartSpan(bg, "trace")
		ctx = slog.Context(ctx, slog.F("ctx", io.EOF))
		l = l.With(slog.F("with", 2))

		l.Info(ctx, "meow", slog.F("hi", "xd"))

		assert.Equal(t, slog.SinkEntry{
			Time: s.entries[0].Time,

			Level:   slog.LevelInfo,
			Message: "meow",

			LoggerName: "hello.hello2",

			File: slogTestFile,
			Func: "cdr.dev/slog_test.TestLogger.func3",
			Line: 101,

			SpanContext: span.SpanContext(),

			Fields: slog.M(
				slog.F("with", 2),
				slog.F("ctx", io.EOF),
				slog.F("hi", "xd"),
			),
		}, s.entries[0], "entry")
	})
}

func TestLevel_String(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "slog.Level(12)", slog.Level(12).String(), "level string")
}
