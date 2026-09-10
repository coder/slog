package slog_test

import (
	"testing"

	"cdr.dev/slog/v3"
	"cdr.dev/slog/v3/internal/assert"
)

func TestBuffer(t *testing.T) {
	t.Parallel()

	debugEntry := func(msg string) slog.SinkEntry {
		return slog.SinkEntry{Level: slog.LevelDebug, Message: msg}
	}
	infoEntry := func(msg string) slog.SinkEntry {
		return slog.SinkEntry{Level: slog.LevelInfo, Message: msg}
	}

	t.Run("ForwardsAtOrAboveLevel", func(t *testing.T) {
		t.Parallel()

		s := &fakeSink{}
		b := slog.NewBuffer(slog.LevelInfo, 8, s)

		b.LogEntry(bg, infoEntry("info"))
		b.LogEntry(bg, slog.SinkEntry{Level: slog.LevelError, Message: "error"})

		assert.Len(t, "forwarded", 2, s.entries)
		assert.Equal(t, "first", "info", s.entries[0].Message)
		assert.Equal(t, "second", "error", s.entries[1].Message)
	})

	t.Run("BuffersBelowLevelUntilFlush", func(t *testing.T) {
		t.Parallel()

		s := &fakeSink{}
		b := slog.NewBuffer(slog.LevelInfo, 8, s)

		b.LogEntry(bg, debugEntry("debug1"))
		b.LogEntry(bg, debugEntry("debug2"))
		// Nothing below the level is forwarded yet.
		assert.Len(t, "held", 0, s.entries)

		b.Flush(bg)
		assert.Len(t, "flushed", 2, s.entries)
		assert.Equal(t, "first", "debug1", s.entries[0].Message)
		assert.Equal(t, "second", "debug2", s.entries[1].Message)
	})

	t.Run("InterleavesImmediateAndBuffered", func(t *testing.T) {
		t.Parallel()

		s := &fakeSink{}
		b := slog.NewBuffer(slog.LevelInfo, 8, s)

		b.LogEntry(bg, debugEntry("debug"))
		b.LogEntry(bg, infoEntry("info"))
		// The info entry is forwarded immediately; the debug entry waits.
		assert.Len(t, "immediate", 1, s.entries)
		assert.Equal(t, "immediate", "info", s.entries[0].Message)

		b.Flush(bg)
		assert.Len(t, "after flush", 2, s.entries)
		assert.Equal(t, "buffered", "debug", s.entries[1].Message)
	})

	t.Run("EvictsOldestWhenFull", func(t *testing.T) {
		t.Parallel()

		s := &fakeSink{}
		b := slog.NewBuffer(slog.LevelInfo, 2, s)

		b.LogEntry(bg, debugEntry("debug1"))
		b.LogEntry(bg, debugEntry("debug2"))
		b.LogEntry(bg, debugEntry("debug3"))

		b.Flush(bg)
		// debug1 was evicted; the two newest remain in order.
		assert.Len(t, "flushed", 2, s.entries)
		assert.Equal(t, "first", "debug2", s.entries[0].Message)
		assert.Equal(t, "second", "debug3", s.entries[1].Message)
	})

	t.Run("FlushIsIdempotent", func(t *testing.T) {
		t.Parallel()

		s := &fakeSink{}
		b := slog.NewBuffer(slog.LevelInfo, 8, s)

		b.LogEntry(bg, debugEntry("debug"))
		b.Flush(bg)
		b.Flush(bg)

		assert.Len(t, "flushed once", 1, s.entries)
	})

	t.Run("ZeroSizeDropsBelowLevel", func(t *testing.T) {
		t.Parallel()

		s := &fakeSink{}
		b := slog.NewBuffer(slog.LevelInfo, 0, s)

		b.LogEntry(bg, debugEntry("debug"))
		b.LogEntry(bg, infoEntry("info"))
		b.Flush(bg)

		// Only the info entry is forwarded; the debug entry was dropped.
		assert.Len(t, "forwarded", 1, s.entries)
		assert.Equal(t, "only", "info", s.entries[0].Message)
	})

	t.Run("SyncForwardsToWrappedSinks", func(t *testing.T) {
		t.Parallel()

		s := &fakeSink{}
		b := slog.NewBuffer(slog.LevelInfo, 8, s)

		b.Sync()
		assert.Equal(t, "syncs", 1, s.syncs)
	})
}

// TestBufferWithLogger exercises Buffer through a Logger to confirm the logger
// must run at the buffered level for lower-level entries to reach the sink.
func TestBufferWithLogger(t *testing.T) {
	t.Parallel()

	s := &fakeSink{}
	b := slog.NewBuffer(slog.LevelInfo, 8, s)
	// The logger must be at LevelDebug so it does not drop debug entries before
	// they reach the buffer.
	log := slog.Make(b).Leveled(slog.LevelDebug)

	log.Debug(bg, "debug")
	log.Info(bg, "info")
	assert.Len(t, "immediate", 1, s.entries)
	assert.Equal(t, "immediate", "info", s.entries[0].Message)

	b.Flush(bg)
	assert.Len(t, "after flush", 2, s.entries)
	assert.Equal(t, "buffered", "debug", s.entries[1].Message)
}
