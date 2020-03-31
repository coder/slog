// Package sloghuman contains the slogger
// that writes logs in a human readable format.
package sloghuman // import "cdr.dev/slog/v2/sloggers/sloghuman"

import (
	"context"
	"io"
	"strings"

	"cdr.dev/slog/v2"
	"cdr.dev/slog/v2/internal/entryhuman"
	"cdr.dev/slog/v2/internal/syncwriter"
)

// Make creates a logger that writes logs in a human
// readable YAML like format to the given writer.
//
// If the writer implements Sync() error then
// it will be called when syncing.
func Make(ctx context.Context, w io.Writer) slog.SinkContext {
	return slog.Make(ctx, &humanSink{
		w:  syncwriter.New(w),
		w2: w,
	})
}

type humanSink struct {
	w  *syncwriter.Writer
	w2 io.Writer
}

func (s humanSink) LogEntry(_ context.Context, ent slog.SinkEntry) {
	str := entryhuman.Fmt(s.w2, ent)
	lines := strings.Split(str, "\n")

	// We need to add 4 spaces before every field line for readability.
	// humanfmt doesn't do it for us because the testSink doesn't want
	// it as *testing.T automatically does it.
	fieldsLines := lines[1:]
	for i, line := range fieldsLines {
		if line == "" {
			continue
		}
		fieldsLines[i] = strings.Repeat(" ", 2) + line
	}

	str = strings.Join(lines, "\n")

	s.w.Write("sloghuman", []byte(str+"\n"))
}

func (s humanSink) Sync() {
	s.w.Sync("sloghuman")
}
