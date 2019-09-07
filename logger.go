package slog

import (
	"context"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"os"
	"strings"
	"sync"

	"go.coder.com/slog/internal/humanfmt"
	"go.coder.com/slog/slogcore"
)

type humanSink struct {
	mu    sync.Mutex
	w     io.Writer
	color bool
}

func (s *humanSink) WriteLogEntry(ent slogcore.Entry) {
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

func isTTY(w io.Writer) bool {
	f, ok := w.(interface {
		Fd() uintptr
	})
	return ok && terminal.IsTerminal(int(f.Fd()))
}

func Human(w io.Writer) Logger {
	return Make(&humanSink{
		w:     w,
		color: isTTY(w),
	})
}

// Make creates a logger that writes logs to sink.
func Make(sink slogcore.Sink) Logger {
	return logger{
		sink: sink,
	}
}

type logger struct {
	sink slogcore.Sink
	l    parsedFields
}

func (sl logger) Debug(ctx context.Context, msg string, fields ...Field) {
	sl.log(ctx, slogcore.Debug, msg, fields)
}

func (sl logger) Info(ctx context.Context, msg string, fields ...Field) {
	sl.log(ctx, slogcore.Info, msg, fields)
}

func (sl logger) Warn(ctx context.Context, msg string, fields ...Field) {
	sl.log(ctx, slogcore.Warn, msg, fields)
}

func (sl logger) Error(ctx context.Context, msg string, fields ...Field) {
	sl.log(ctx, slogcore.Error, msg, fields)
}

func (sl logger) Critical(ctx context.Context, msg string, fields ...Field) {
	sl.log(ctx, slogcore.Critical, msg, fields)
}

func (sl logger) Fatal(ctx context.Context, msg string, fields ...Field) {
	sl.log(ctx, slogcore.Fatal, msg, fields)
	os.Exit(1)
}

func (sl logger) With(fields ...Field) Logger {
	sl.l = sl.l.withFields(fields)
	return sl
}

func (sl logger) log(ctx context.Context, level slogcore.Level, msg string, fields []Field) {
	ent := sl.l.entry(ctx, entryParams{
		level:  level,
		msg:    msg,
		fields: fields,
		skip:   2,
	})

	sl.sink.WriteLogEntry(ent)

	if ent.Level == slogcore.Fatal {
		os.Exit(1)
	}
}
