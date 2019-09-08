// Package sloghuman contains the slogger
// that writes logs in a human readable format.
package sloghuman // import "go.coder.com/slog/sloggers/sloghuman"

import (
	"context"
	"io"
	"strings"
	"sync"

	"go.coder.com/slog"
	"go.coder.com/slog/internal/humanfmt"
)

type humanSink struct {
	mu    sync.Mutex
	w     io.Writer
	color bool
}

var _ slog.Sink = &humanSink{}

func (s *humanSink) LogEntry(ctx context.Context, ent slog.Entry) {
	str := humanfmt.Entry(ent, s.color)
	lines := strings.Split(str, "\n")

	fieldsLines := lines[1:]
	for i, line := range fieldsLines {
		if line == "" {
			continue
		}
		fieldsLines[i] = "\t" + line
	}

	str = strings.Join(lines, "\n")

	s.writeString(str + "\n")
}

func (s *humanSink) writeString(str string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	io.WriteString(s.w, str)
}

// Make creates a logger that writes logs in a human
// readable YAML like format to the given writer.
func Make(w io.Writer) slog.Logger {
	return slog.Make(&humanSink{
		w:     w,
		color: humanfmt.IsTTY(w),
	})
}
