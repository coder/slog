// Package sloghuman contains the slogger
// that writes logs in a human readable format.
package sloghuman // import "cdr.dev/slog/sloggers/sloghuman"

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"sync"

	"cdr.dev/slog"
	"cdr.dev/slog/internal/entryhuman"
	"cdr.dev/slog/internal/syncwriter"
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

	buf2 := bufPool.Get().(*bytes.Buffer)
	buf2.Reset()
	defer bufPool.Put(buf2)

	entryhuman.Fmt(buf1, s.w2, ent)

	var (
		i  int
		sc = bufio.NewScanner(buf1)
	)

	// We need to add 4 spaces before every field line for readability.
	// humanfmt doesn't do it for us because the testSink doesn't want
	// it as *testing.T automatically does it.
	for ; sc.Scan(); i++ {
		if i > 0 && len(sc.Bytes()) > 0 {
			buf2.Write([]byte("    "))
		}
		buf2.Write(sc.Bytes())
		buf2.WriteByte('\n')
	}

	s.w.Write("sloghuman", buf2.Bytes())
}

func (s humanSink) Sync() {
	s.w.Sync("sloghuman")
}
