package slog

import (
	"context"
	"log"
	"strings"
)

// Stdlib creates a standard library logger from the given logger.
//
// All logs will be logged at the Info level and the given ctx
// will be passed to the logger's Info method, thereby logging
// all fields and tracing info in the context.
//
// You can redirect the stdlib default logger with log.SetOutput
// to the Writer on the logger returned by this function.
// See the example.
func Stdlib(ctx context.Context, l Logger) *log.Logger {
	l.skip += 3

	l = l.Named("stdlib")

	w := &stdlogWriter{
		ctx: ctx,
		l:   l,
	}

	return log.New(w, "", 0)
}

type stdlogWriter struct {
	ctx context.Context
	l   Logger
}

func (w stdlogWriter) Write(p []byte) (n int, err error) {
	msg := string(p)
	// stdlib includes a trailing newline on the msg that
	// we do not want.
	msg = strings.TrimSuffix(msg, "\n")

	w.l.Info(w.ctx, msg)

	return len(p), nil
}
