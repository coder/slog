// Package sloghuman contains the slogger
// that writes logs in a human readable format.
package sloghuman // import "cdr.dev/slog/v3/sloggers/sloghuman"

import (
	"bytes"
	"context"
	"io"
	"sync"

	"cdr.dev/slog/v3"
	"cdr.dev/slog/v3/internal/entryhuman"
	"cdr.dev/slog/v3/internal/syncwriter"
)

// Sink creates a slog.Sink that writes logs in a human
// readable YAML like format to the given writer.
//
// If the writer implements Sync() error then
// it will be called when syncing.
func Sink(w io.Writer) slog.Sink {
	return &humanSink{
		w:  syncwriter.New(w),
		w2: w,
	}
}

type humanSink struct {
	w  *syncwriter.Writer
	w2 io.Writer
}

var bufPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 256))
	},
}

func (s humanSink) LogEntry(ctx context.Context, ent slog.SinkEntry) {
	buf1 := bufPool.Get().(*bytes.Buffer)
	buf1.Reset()
	defer bufPool.Put(buf1)

	entryhuman.Fmt(buf1, s.w2, ent)
	by := buf1.Bytes()

	// Prepare output buffer and indent lines after the first.
	buf2 := bufPool.Get().(*bytes.Buffer)
	buf2.Reset()
	defer bufPool.Put(buf2)

	// Pre-grow: worst-case add 4 spaces per non-empty line after the first.
	newlines := bytes.Count(by, []byte{'\n'})
	buf2.Grow(len(by) + newlines*4)

	start := 0
	lineIdx := 0
	for {
		idx := bytes.IndexByte(by[start:], '\n')
		var line []byte
		if idx >= 0 {
			line = by[start : start+idx]
		} else {
			line = by[start:]
		}

		if lineIdx > 0 && len(line) > 0 {
			buf2.WriteString("    ")
		}
		buf2.Write(line)
		buf2.WriteByte('\n')

		if idx < 0 {
			break
		}
		start += idx + 1
		lineIdx++
		if start >= len(by) {
			// The original logic always wrote a trailing newline
			// even for an empty last line; we already wrote it.
			break
		}
	}

	s.w.Write("sloghuman", buf2.Bytes())
}

func (s humanSink) Sync() {
	s.w.Sync("sloghuman")
}
