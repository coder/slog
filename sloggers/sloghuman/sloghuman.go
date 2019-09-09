// Package sloghuman contains the slogger
// that writes logs in a human readable format.
package sloghuman // import "go.coder.com/slog/sloggers/sloghuman"

import (
	"context"
	"io"
	"strings"

	"go.coder.com/slog"
	"go.coder.com/slog/internal/humanfmt"
	"go.coder.com/slog/internal/syncwriter"
)

type humanSink struct {
	w     *syncwriter.Writer
	color bool
}

func (s humanSink) LogEntry(ctx context.Context, ent slog.Entry) {
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
		fieldsLines[i] = strings.Repeat("    ", 4) + line
	}

	str = strings.Join(lines, "\n")

	io.WriteString(s.w, str+"\n")
}

// Make creates a logger that writes logs in a human
// readable YAML like format to the given writer.
func Make(w io.Writer) slog.Logger {
	return slog.Make(&humanSink{
		w:     syncwriter.New(w),
		color: humanfmt.IsTTY(w),
	})
}
