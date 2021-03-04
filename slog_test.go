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
		l := slog.Make(s1)
		l = l.Leveled(slog.LevelError)
		l = l.AppendSinks(s2)

		l.Info(bg, "wow", slog.Error(io.EOF))
		l.Error(bg, "meow", slog.Error(io.ErrUnexpectedEOF))

		assert.Equal(t, "syncs", 1, s1.syncs)
		assert.Len(t, "entries", 1, s1.entries)

		assert.Equal(t, "sinks", s1, s2)
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

		assert.Len(t, "entries", 1, s.entries)
		assert.Equal(t, "entry", slog.SinkEntry{
			Time: s.entries[0].Time,

			Level:   slog.LevelInfo,
			Message: "logging in helper",

			File: slogTestFile,
			Func: "cdr.dev/slog_test.TestLogger.func2",
			Line: 67,

			Fields: slog.M(
				slog.F("ctx", 1024),
			),
		}, s.entries[0])
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

		assert.Len(t, "entries", 1, s.entries)
		assert.Equal(t, "entry", slog.SinkEntry{
			Time: s.entries[0].Time,

			Level:   slog.LevelInfo,
			Message: "meow",

			LoggerNames: []string{"hello", "hello2"},

			File: slogTestFile,
			Func: "cdr.dev/slog_test.TestLogger.func3",
			Line: 98,

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
