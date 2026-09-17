package slog

import (
	"context"
	"sync"
)

// FlightRecorder returns a Logger that records entries below its level in a
// fixed-size, in-memory ring buffer instead of dropping them, while still
// forwarding entries at or above its level to the sinks as usual. The recorded
// entries are forwarded to the sinks ("flushed") automatically whenever an entry
// at LevelError or above is logged, and can be flushed manually with
// Logger.Flush.
//
// This keeps verbose (e.g. debug) logs available for diagnosing a failure
// without emitting them during normal operation: run the Logger at LevelInfo
// with a flight recorder, and the debug entries leading up to an error are
// emitted only when the error occurs. When the buffer is full, the oldest
// recorded entry is dropped to make room for the newest.
//
// Enabling flight recording is independent of the Logger's level, so
// Make(sink).Leveled(LevelInfo).FlightRecorder(n) and
// Make(sink).FlightRecorder(n).Leveled(LevelInfo) are equivalent. If size is
// <= 0, no entries are recorded.
//
// The returned Logger and any Loggers derived from it share the same underlying
// buffer.
func (l Logger) FlightRecorder(size int) Logger {
	l.flightRecorder = newFlightRecorder(size)
	return l
}

// Flush forwards the entries currently held by the Logger's flight recorder to
// its sinks, oldest first, and empties the buffer. It is a no-op when flight
// recording is not enabled. Flushing also happens automatically when an entry at
// LevelError or above is logged.
func (l Logger) Flush(ctx context.Context) {
	if l.flightRecorder != nil {
		l.flightRecorder.flush(ctx, l.sinks)
	}
}

// flightRecorder keeps a rolling history of entries in a fixed-size ring buffer
// until they are flushed. It is safe for concurrent use.
type flightRecorder struct {
	size int

	mu      sync.Mutex
	ring    []SinkEntry
	written int
}

func newFlightRecorder(size int) *flightRecorder {
	if size < 0 {
		size = 0
	}
	return &flightRecorder{
		size: size,
		ring: make([]SinkEntry, size),
	}
}

// record stores e in the ring, overwriting the oldest entry once full.
func (f *flightRecorder) record(e SinkEntry) {
	if f.size == 0 {
		return
	}
	f.mu.Lock()
	// A running write index means the write path never has to branch on whether
	// the ring is full: writes always land at written%size, and drain derives
	// the entry count and oldest position from written.
	f.ring[f.written%f.size] = e
	f.written++
	f.mu.Unlock()
}

// flush forwards the recorded entries to sinks, oldest first, and empties the
// buffer.
func (f *flightRecorder) flush(ctx context.Context, sinks []Sink) {
	f.mu.Lock()
	entries := f.drain()
	f.mu.Unlock()

	for _, e := range entries {
		for _, s := range sinks {
			s.LogEntry(ctx, e)
		}
	}
}

// drain returns the recorded entries oldest first and resets the buffer. It must
// be called with f.mu held.
func (f *flightRecorder) drain() []SinkEntry {
	n := f.written
	if n > f.size {
		n = f.size
	}
	if n == 0 {
		return nil
	}
	oldest := 0
	if f.written > f.size {
		oldest = f.written % f.size
	}
	entries := make([]SinkEntry, 0, n)
	for i := 0; i < n; i++ {
		entries = append(entries, f.ring[(oldest+i)%f.size])
	}
	f.written = 0
	return entries
}
