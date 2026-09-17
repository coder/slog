package slog_test

import (
	"context"
	"sync"
	"testing"

	"cdr.dev/slog/v3"
	"cdr.dev/slog/v3/internal/assert"
)

func TestFlightRecorder(t *testing.T) {
	t.Parallel()

	t.Run("ForwardsAtOrAboveLevel", func(t *testing.T) {
		t.Parallel()

		s := &fakeSink{}
		log := slog.Make(s).Leveled(slog.LevelInfo).FlightRecorder(8)

		log.Info(bg, "info")
		log.Warn(bg, "warn")

		assert.Len(t, "forwarded", 2, s.entries)
		assert.Equal(t, "first", "info", s.entries[0].Message)
		assert.Equal(t, "second", "warn", s.entries[1].Message)
	})

	t.Run("BuffersBelowLevelUntilFlush", func(t *testing.T) {
		t.Parallel()

		s := &fakeSink{}
		log := slog.Make(s).Leveled(slog.LevelInfo).FlightRecorder(8)

		log.Debug(bg, "debug1")
		log.Debug(bg, "debug2")
		// Nothing below the level is forwarded yet.
		assert.Len(t, "held", 0, s.entries)

		log.Flush(bg)
		assert.Len(t, "flushed", 2, s.entries)
		assert.Equal(t, "first", "debug1", s.entries[0].Message)
		assert.Equal(t, "second", "debug2", s.entries[1].Message)
	})

	t.Run("InterleavesImmediateAndBuffered", func(t *testing.T) {
		t.Parallel()

		s := &fakeSink{}
		log := slog.Make(s).Leveled(slog.LevelInfo).FlightRecorder(8)

		log.Debug(bg, "debug")
		log.Info(bg, "info")
		// The info entry is forwarded immediately; the debug entry waits.
		assert.Len(t, "immediate", 1, s.entries)
		assert.Equal(t, "immediate", "info", s.entries[0].Message)

		log.Flush(bg)
		assert.Len(t, "after flush", 2, s.entries)
		assert.Equal(t, "buffered", "debug", s.entries[1].Message)
	})

	t.Run("FlushesAutomaticallyBeforeError", func(t *testing.T) {
		t.Parallel()

		s := &fakeSink{}
		log := slog.Make(s).Leveled(slog.LevelInfo).FlightRecorder(8)

		log.Debug(bg, "debug1")
		log.Debug(bg, "debug2")
		// The buffered history is still held until the error triggers a flush.
		assert.Len(t, "held", 0, s.entries)

		log.Error(bg, "boom")
		// The buffered debug entries precede the error that triggered the flush.
		assert.Len(t, "flushed with error", 3, s.entries)
		assert.Equal(t, "first", "debug1", s.entries[0].Message)
		assert.Equal(t, "second", "debug2", s.entries[1].Message)
		assert.Equal(t, "error last", "boom", s.entries[2].Message)
	})

	t.Run("CriticalFlushesBufferedHistory", func(t *testing.T) {
		t.Parallel()

		s := &fakeSink{}
		log := slog.Make(s).Leveled(slog.LevelInfo).FlightRecorder(8)

		log.Debug(bg, "debug")
		log.Critical(bg, "critical")

		assert.Len(t, "flushed", 2, s.entries)
		assert.Equal(t, "first", "debug", s.entries[0].Message)
		assert.Equal(t, "second", "critical", s.entries[1].Message)
	})

	t.Run("EvictsOldestWhenFull", func(t *testing.T) {
		t.Parallel()

		s := &fakeSink{}
		log := slog.Make(s).Leveled(slog.LevelInfo).FlightRecorder(2)

		log.Debug(bg, "debug1")
		log.Debug(bg, "debug2")
		log.Debug(bg, "debug3")

		log.Flush(bg)
		// debug1 was evicted; the two newest remain in order.
		assert.Len(t, "flushed", 2, s.entries)
		assert.Equal(t, "first", "debug2", s.entries[0].Message)
		assert.Equal(t, "second", "debug3", s.entries[1].Message)
	})

	t.Run("EvictsOldestAcrossMultipleWraps", func(t *testing.T) {
		t.Parallel()

		s := &fakeSink{}
		log := slog.Make(s).Leveled(slog.LevelInfo).FlightRecorder(2)

		log.Debug(bg, "debug1")
		log.Debug(bg, "debug2")
		log.Debug(bg, "debug3")
		log.Debug(bg, "debug4")
		log.Debug(bg, "debug5")

		log.Flush(bg)
		// Only the two newest survive after wrapping multiple times.
		assert.Len(t, "flushed", 2, s.entries)
		assert.Equal(t, "first", "debug4", s.entries[0].Message)
		assert.Equal(t, "second", "debug5", s.entries[1].Message)
	})

	t.Run("FlushIsIdempotent", func(t *testing.T) {
		t.Parallel()

		s := &fakeSink{}
		log := slog.Make(s).Leveled(slog.LevelInfo).FlightRecorder(8)

		log.Debug(bg, "debug")
		log.Flush(bg)
		log.Flush(bg)

		assert.Len(t, "flushed once", 1, s.entries)
	})

	t.Run("ZeroSizeDropsBelowLevel", func(t *testing.T) {
		t.Parallel()

		s := &fakeSink{}
		log := slog.Make(s).Leveled(slog.LevelInfo).FlightRecorder(0)

		log.Debug(bg, "debug")
		log.Info(bg, "info")
		log.Flush(bg)

		// Only the info entry is forwarded; the debug entry was dropped.
		assert.Len(t, "forwarded", 1, s.entries)
		assert.Equal(t, "only", "info", s.entries[0].Message)
	})

	t.Run("OrderIndependentWithLeveled", func(t *testing.T) {
		t.Parallel()

		s1 := &fakeSink{}
		s2 := &fakeSink{}
		leveledFirst := slog.Make(s1).Leveled(slog.LevelInfo).FlightRecorder(8)
		recorderFirst := slog.Make(s2).FlightRecorder(8).Leveled(slog.LevelInfo)

		for _, log := range []slog.Logger{leveledFirst, recorderFirst} {
			log.Debug(bg, "debug")
			log.Info(bg, "info")
		}

		// Both loggers hold the debug entry and forward the info entry.
		assert.Len(t, "leveled first immediate", 1, s1.entries)
		assert.Len(t, "recorder first immediate", 1, s2.entries)

		leveledFirst.Flush(bg)
		recorderFirst.Flush(bg)

		assert.Len(t, "leveled first flushed", 2, s1.entries)
		assert.Len(t, "recorder first flushed", 2, s2.entries)
		assert.Equal(t, "leveled first buffered", "debug", s1.entries[1].Message)
		assert.Equal(t, "recorder first buffered", "debug", s2.entries[1].Message)
	})

	t.Run("DerivedLoggersShareRecorder", func(t *testing.T) {
		t.Parallel()

		s := &fakeSink{}
		base := slog.Make(s).Leveled(slog.LevelInfo).FlightRecorder(8)
		derived := base.Named("sub").With(slog.F("key", "value"))

		derived.Debug(bg, "debug")
		// Flushing through the base logger emits the entry recorded via the
		// derived logger, and the derived fields and names are captured.
		base.Flush(bg)

		assert.Len(t, "flushed", 1, s.entries)
		assert.Equal(t, "message", "debug", s.entries[0].Message)
		assert.Equal(t, "names", []string{"sub"}, s.entries[0].LoggerNames)
		assert.Equal(t, "fields", slog.M(slog.F("key", "value")), s.entries[0].Fields)
	})

	t.Run("FlushForwardsToAllSinks", func(t *testing.T) {
		t.Parallel()

		s1 := &fakeSink{}
		s2 := &fakeSink{}
		log := slog.Make(s1, s2).Leveled(slog.LevelInfo).FlightRecorder(8)

		log.Debug(bg, "debug")
		log.Flush(bg)

		assert.Len(t, "sink1", 1, s1.entries)
		assert.Len(t, "sink2", 1, s2.entries)
	})

	t.Run("NoRecorderDropsBelowLevel", func(t *testing.T) {
		t.Parallel()

		s := &fakeSink{}
		log := slog.Make(s).Leveled(slog.LevelInfo)

		log.Debug(bg, "debug")
		log.Info(bg, "info")
		// Without a flight recorder, below-level entries are dropped and Flush
		// is a no-op.
		log.Flush(bg)

		assert.Len(t, "forwarded", 1, s.entries)
		assert.Equal(t, "only", "info", s.entries[0].Message)
	})

	t.Run("ConcurrentRecordAndFlush", func(t *testing.T) {
		t.Parallel()

		s := &lockedSink{}
		log := slog.Make(s).Leveled(slog.LevelInfo).FlightRecorder(64)

		var wg sync.WaitGroup
		for i := 0; i < 8; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 100; j++ {
					log.Debug(bg, "debug")
				}
				log.Flush(bg)
			}()
		}
		wg.Wait()

		// The exact count depends on interleaving; the test only asserts that
		// concurrent record and flush do not race or panic.
		log.Flush(bg)
	})
}

// lockedSink is a concurrency-safe sink for race detection tests.
type lockedSink struct {
	mu      sync.Mutex
	entries []slog.SinkEntry
}

func (s *lockedSink) LogEntry(_ context.Context, e slog.SinkEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = append(s.entries, e)
}

func (s *lockedSink) Sync() {}
