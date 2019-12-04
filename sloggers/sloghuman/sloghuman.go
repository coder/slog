// Package sloghuman contains the slogger
// that writes logs in a human readable format.
package sloghuman // import "cdr.dev/slog/sloggers/sloghuman"

import (
	"context"
	"io"
	"os"
	"strings"

	"cdr.dev/slog"
	"cdr.dev/slog/internal/humanfmt"
	"cdr.dev/slog/internal/syncwriter"
)

// Make creates a logger that writes logs in a human
// readable YAML like format to the given writer.
//
// If the writer implements Sync() error then
// it will be called when syncing.
func Make(w io.Writer) slog.Logger {
	return slog.Make(&humanSink{
		w:     syncwriter.New(w),
		color: humanfmt.IsTTY(w) || os.Getenv("FORCE_COLOR") != "",
	})
}

type humanSink struct {
	w     *syncwriter.Writer
	color bool
}

func (s humanSink) LogEntry(ctx context.Context, ent slog.SinkEntry) error {
	str := humanfmt.Entry(ent, s.color)
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

	io.WriteString(s.w, str+"\n")
	return nil
}

func (s humanSink) Sync() error {
	return s.w.Sync()
}
