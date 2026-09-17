package slog

import (
	"context"
	"sync"
)

// FlightRecorder is a Sink that holds entries below Level in a fixed-size,
// in-memory ring buffer instead of forwarding them. Entries at or above Level
// are forwarded to the wrapped sinks immediately. Flush forwards the currently
// buffered entries to the wrapped sinks, oldest first, and empties the buffer.
//
// FlightRecorder lets a program run its Logger at a low level
// (e.g. LevelDebug) so that verbose entries are captured, while only emitting
// those entries on demand, such as when an error or connection failure occurs.
// When the buffer is full, the oldest held entry is dropped to make room for
// the newest.
//
// FlightRecorder is safe for concurrent use.
type FlightRecorder struct {
	level Level
	next  []Sink

	mu    sync.Mutex
	ring  []SinkEntry
	start int
}

var _ Sink = (*FlightRecorder)(nil)

// Flusher is implemented by sinks that defer entries and can emit them on
// demand, such as FlightRecorder. It lets callers trigger a flush without
// depending on the concrete sink type.
type Flusher interface {
	Flush(ctx context.Context)
}

var _ Flusher = (*FlightRecorder)(nil)

// NewFlightRecorder returns a FlightRecorder that forwards entries at or above
// level to next immediately and holds up to size lower-level entries in memory
// until Flush is called. If size is <= 0, lower-level entries are dropped and
// FlightRecorder only forwards entries at or above level.
func NewFlightRecorder(level Level, size int, next ...Sink) *FlightRecorder {
	if size < 0 {
		size = 0
	}
	return &FlightRecorder{
		level: level,
		next:  next,
		ring:  make([]SinkEntry, 0, size),
	}
}

// LogEntry implements Sink. Entries at or above the flight recorder's level
// are forwarded to the wrapped sinks immediately; lower-level entries are held
// until Flush.
func (b *FlightRecorder) LogEntry(ctx context.Context, e SinkEntry) {
	if e.Level >= b.level {
		for _, s := range b.next {
			s.LogEntry(ctx, e)
		}
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	if cap(b.ring) == 0 {
		return
	}
	if len(b.ring) < cap(b.ring) {
		b.ring = append(b.ring, e)
		return
	}
	// The ring is full, so overwrite the oldest entry and advance start.
	b.ring[b.start] = e
	b.start = (b.start + 1) % cap(b.ring)
}

// Flush forwards the buffered entries to the wrapped sinks, oldest first, and
// empties the buffer. It is safe to call multiple times; a second call with no
// intervening entries forwards nothing.
func (b *FlightRecorder) Flush(ctx context.Context) {
	b.mu.Lock()
	entries := b.drain()
	b.mu.Unlock()

	for _, e := range entries {
		for _, s := range b.next {
			s.LogEntry(ctx, e)
		}
	}
}

// drain returns the buffered entries oldest first and resets the ring. It must
// be called with b.mu held.
func (b *FlightRecorder) drain() []SinkEntry {
	n := len(b.ring)
	if n == 0 {
		return nil
	}
	entries := make([]SinkEntry, 0, n)
	for i := 0; i < n; i++ {
		entries = append(entries, b.ring[(b.start+i)%cap(b.ring)])
	}
	b.ring = b.ring[:0]
	b.start = 0
	return entries
}

// Sync implements Sink by syncing the wrapped sinks. It does not flush buffered
// entries; call Flush for that.
func (b *FlightRecorder) Sync() {
	for _, s := range b.next {
		s.Sync()
	}
}
