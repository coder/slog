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
	entries []slog.SinkEntry

	syncs int
}

func (s *fakeSink) LogEntry(_ context.Context, e slog.SinkEntry) {
	s.entries = append(s.entries, e)
}

func (s *fakeSink) Sync() {
	s.syncs++
}

func TestLogger(t *testing.T) {
	t.Parallel()

	t.Run("basic", func(t *testing.T) {
		t.Parallel()

		s1 := &fakeSink{}
		s2 := &fakeSink{}
		var ctx context.Context
		ctx = slog.Make(context.Background(), s1, s2)
		ctx = slog.Leveled(ctx, slog.LevelError)

		slog.Info(ctx, "wow", slog.Err(io.EOF))
		slog.Error(ctx, "meow", slog.Err(io.ErrUnexpectedEOF))

		assert.Equal(t, "syncs", 1, s1.syncs)
		assert.Len(t, "entries", 1, s1.entries)

		assert.Equal(t, "sinks", s1, s2)
	})

	t.Run("helper", func(t *testing.T) {
		t.Parallel()

		s := &fakeSink{}
		var ctx context.Context
		ctx = slog.Make(context.Background(), s)
		h := func(ctx context.Context) {
			slog.Helper()
			slog.Info(ctx, "logging in helper")
		}

		ctx = slog.With(ctx, slog.F(
			"ctx", 1024),
		)
		h(ctx)

		assert.Len(t, "entries", 1, s.entries)
		assert.Equal(t, "entry", slog.SinkEntry{
			Time: s.entries[0].Time,

			Level:   slog.LevelInfo,
			Message: "logging in helper",

			File: slogTestFile,
			Func: "cdr.dev/slog_test.TestLogger.func2",
			Line: 66,

			Fields: slog.M(
				slog.F("ctx", 1024),
			),
		}, s.entries[0])
	})

	t.Run("entry", func(t *testing.T) {
		t.Parallel()

		s := &fakeSink{}
		var ctx context.Context
		ctx = slog.Make(context.Background(), s)
		ctx = slog.Named(ctx, "hello")
		ctx = slog.Named(ctx, "hello2")

		ctx, span := trace.StartSpan(ctx, "trace")
		ctx = slog.With(ctx, slog.F("ctx", io.EOF))
		ctx = slog.With(ctx, slog.F("with", 2))

		slog.Info(ctx, "meow", slog.F("hi", "xd"))

		assert.Len(t, "entries", 1, s.entries)
		assert.Equal(t, "entry", slog.SinkEntry{
			Time: s.entries[0].Time,

			Level:   slog.LevelInfo,
			Message: "meow",

			LoggerNames: []string{"hello", "hello2"},

			File: slogTestFile,
			Func: "cdr.dev/slog_test.TestLogger.func3",
			Line: 97,

			SpanContext: span.SpanContext(),

			Fields: slog.M(
				slog.F("with", 2),
				slog.F("ctx", io.EOF),
				slog.F("hi", "xd"),
			),
		}, s.entries[0])
	})

	t.Run("levels", func(t *testing.T) {
		t.Parallel()

		s := &fakeSink{}
		var ctx context.Context
		ctx = slog.Make(context.Background(), s)

		exits := 0
		ctx = slog.SetExit(ctx, func(int) {
			exits++
		})

		ctx = slog.Leveled(ctx, slog.LevelDebug)
		slog.Debug(ctx, "")
		slog.Info(ctx, "")
		slog.Warn(ctx, "")
		slog.Error(ctx, "")
		slog.Critical(ctx, "")
		slog.Fatal(ctx, "")

		assert.Len(t, "entries", 6, s.entries)
		assert.Equal(t, "syncs", 3, s.syncs)
		assert.Equal(t, "level", slog.LevelDebug, s.entries[0].Level)
		assert.Equal(t, "level", slog.LevelInfo, s.entries[1].Level)
		assert.Equal(t, "level", slog.LevelWarn, s.entries[2].Level)
		assert.Equal(t, "level", slog.LevelError, s.entries[3].Level)
		assert.Equal(t, "level", slog.LevelCritical, s.entries[4].Level)
		assert.Equal(t, "level", slog.LevelFatal, s.entries[5].Level)
		assert.Equal(t, "exits", 1, exits)
	})
}

func TestLevel_String(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "level string", "slog.Level(12)", slog.Level(12).String())
}
