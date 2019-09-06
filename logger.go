package slog

import (
	"context"
	"io"
	"os"
	"strings"
	"sync"
)

// Make creates a logger that writes logs to w.
func Make(w io.Writer) Logger {
	return makeLogger(w)
}

func makeLogger(w io.Writer) logger {
	return logger{
		mu: &sync.Mutex{},
		w:  w,
	}
}

type logger struct {
	// This is specifically a pointer because we use the logger as a value
	// to make the With functions simpler as the outer logger is shallow
	// cloned automatically.
	// (nhooyr): I wonder if this is really necessary.
	mu *sync.Mutex
	w  io.Writer
	l  parsedFields

	skip int
}

func (sl logger) Debug(ctx context.Context, msg string, fields ...interface{}) {
	sl.log(ctx, levelDebug, msg, fields)
}

func (sl logger) Info(ctx context.Context, msg string, fields ...interface{}) {
	sl.log(ctx, levelInfo, msg, fields)
}

func (sl logger) Warn(ctx context.Context, msg string, fields ...interface{}) {
	sl.log(ctx, levelWarn, msg, fields)
}

func (sl logger) Error(ctx context.Context, msg string, fields ...interface{}) {
	sl.log(ctx, levelError, msg, fields)
}

func (sl logger) Critical(ctx context.Context, msg string, fields ...interface{}) {
	sl.log(ctx, levelCritical, msg, fields)
}

func (sl logger) Fatal(ctx context.Context, msg string, fields ...interface{}) {
	sl.log(ctx, levelFatal, msg, fields)
	os.Exit(1)
}

func (sl logger) With(fields ...interface{}) Logger {
	sl.l = sl.l.withFields(fields)
	return sl
}

func (sl logger) log(ctx context.Context, sev level, msg string, fields []interface{}) {
	ent := sl.l.entry(ctx, entryConfig{
		level:  sev,
		msg:    msg,
		fields: fields,
		skip:   2,
	})

	sl.write(ent)

	if ent.level == levelFatal {
		os.Exit(1)
	}
}

func (sl logger) write(ent entry) {
	s := ent.String()
	lines := strings.Split(s, "\n")

	fieldsLines := lines[1:]
	for i, line := range fieldsLines {
		if line == "" {
			continue
		}
		fieldsLines[i] = "\t" + line
	}

	s = strings.Join(lines, "\n")

	sl.writeString(s + "\n")
}

func (sl logger) writeString(s string) {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	io.WriteString(sl.w, s)
}
