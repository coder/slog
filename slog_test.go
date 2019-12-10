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

var bg = context.Background()

func TestLogger(t *testing.T) {
	t.Parallel()

	t.Run("basic", func(t *testing.T) {
		t.Parallel()

		s1 := &fakeSink{}
		s2 := &fakeSink{}
		l := slog.Make(s1, s2)
		l = l.Leveled(slog.LevelError)

		l.Info(bg, "wow", slog.Error(io.EOF))
		l.Error(bg, "meow", slog.Error(io.ErrUnexpectedEOF))

		assert.Equal(t, 1, s1.syncs, "syncs")
		assert.Len(t, 1, s1.entries, "entries")

		assert.Equal(t, s1, s2, "sinks")
	})

	t.Run("helper", func(t *testing.T) {
		t.Parallel()

		s := &fakeSink{}
		l := slog.Make(s)
		h := func(ctx context.Context) {
			slog.Helper()
			l.Info(ctx, "logging in helper")
		}

		ctx := slog.With(bg, slog.F(
			"ctx", 1024),
		)
		h(ctx)

		assert.Len(t, 1, s.entries, "entries")
		assert.Equal(t, slog.SinkEntry{
			Time: s.entries[0].Time,

			Level:   slog.LevelInfo,
			Message: "logging in helper",

			File: slogTestFile,
			Func: "cdr.dev/slog_test.TestLogger.func2",
			Line: 66,

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
		ctx = slog.With(ctx, slog.F("ctx", io.EOF))
		l = l.With(slog.F("with", 2))

		l.Info(ctx, "meow", slog.F("hi", "xd"))

		assert.Len(t, 1, s.entries, "entries")
		assert.Equal(t, slog.SinkEntry{
			Time: s.entries[0].Time,

			Level:   slog.LevelInfo,
			Message: "meow",

			Names: []string{"hello", "hello2"},

			File: slogTestFile,
			Func: "cdr.dev/slog_test.TestLogger.func3",
			Line: 97,

			SpanContext: span.SpanContext(),

			Fields: slog.M(
				slog.F("with", 2),
				slog.F("ctx", io.EOF),
				slog.F("hi", "xd"),
			),
		}, s.entries[0], "entry")
	})

	t.Run("levels", func(t *testing.T) {
		t.Parallel()

		s := &fakeSink{}
		l := slog.Make(s)

		exits := 0
		l.SetExit(func(int) {
			exits++
		})

		l = l.Leveled(slog.LevelDebug)
		l.Debug(bg, "")
		l.Info(bg, "")
		l.Warn(bg, "")
		l.Error(bg, "")
		l.Critical(bg, "")
		l.Fatal(bg, "")

		assert.Len(t, 6, s.entries, "entries")
		assert.Equal(t, 3, s.syncs, "syncs")
		assert.Equal(t, slog.LevelDebug, s.entries[0].Level, "level")
		assert.Equal(t, slog.LevelInfo, s.entries[1].Level, "level")
		assert.Equal(t, slog.LevelWarn, s.entries[2].Level, "level")
		assert.Equal(t, slog.LevelError, s.entries[3].Level, "level")
		assert.Equal(t, slog.LevelCritical, s.entries[4].Level, "level")
		assert.Equal(t, slog.LevelFatal, s.entries[5].Level, "level")
		assert.Equal(t, 1, exits, "exits")
	})
}

func TestLevel_String(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "slog.Level(12)", slog.Level(12).String(), "level string")
}
