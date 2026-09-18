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
	capacity int

	mu   sync.Mutex
	ring []SinkEntry
	// pos is the index the next entry is written to; length is the number of
	// entries currently held. Both stay bounded by capacity, so neither grows
	// without limit.
	pos    int
	length int
}

func newFlightRecorder(size int) *flightRecorder {
	if size < 0 {
		size = 0
	}
	return &flightRecorder{
		capacity: size,
		ring:     make([]SinkEntry, size),
	}
}

// record stores e in the ring, overwriting the oldest entry once full.
func (f *flightRecorder) record(e SinkEntry) {
	if f.capacity == 0 {
		return
	}
	f.mu.Lock()
	f.ring[f.pos] = e
	f.pos = (f.pos + 1) % f.capacity
	if f.length < f.capacity {
		f.length++
	}
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
	if f.length == 0 {
		return nil
	}
	// Once the ring is full, pos points at the oldest entry; before that the
	// oldest entry is at index 0.
	oldest := 0
	if f.length == f.capacity {
		oldest = f.pos
	}
	entries := make([]SinkEntry, 0, f.length)
	for i := 0; i < f.length; i++ {
		entries = append(entries, f.ring[(oldest+i)%f.capacity])
	}
	f.length = 0
	f.pos = 0
	return entries
}
