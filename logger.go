package slog

import (
	"context"
	"io"
	"os"
	"strings"
	"sync"

	"go.coder.com/slog/internal/humanfmt"
	"go.coder.com/slog/slogcore"
)

type humanSink struct {
	mu sync.Mutex
	w  io.Writer
}

func (w *humanSink) WriteLogEntry(ent slogcore.Entry) {
	s := humanfmt.Entry(ent)
	lines := strings.Split(s, "\n")

	fieldsLines := lines[1:]
	for i, line := range fieldsLines {
		if line == "" {
			continue
		}
		fieldsLines[i] = "\t" + line
	}

	s = strings.Join(lines, "\n")

	w.writeString(s + "\n")
}

func (w *humanSink) writeString(s string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	io.WriteString(w.w, s)
}

func Human(w io.Writer) Logger {
	return Make(&humanSink{
		w: w,
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
