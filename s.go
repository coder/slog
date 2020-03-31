package slog

import (
	"context"
	"log"
	"os"
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
func Stdlib(ctx context.Context) *log.Logger {
	ctx = Named(ctx, "stdlib")

	l, ok := extractContext(ctx)
	if !ok {
		// Give stderr logger if no slog.
		return log.New(os.Stderr, "", 0)
	}
	l.skip += 3
	ctx = makeContext(ctx, l)

	w := &stdlogWriter{
		ctx: ctx,
	}

	return log.New(w, "", 0)
}

type stdlogWriter struct {
	ctx context.Context
}

func (w stdlogWriter) Write(p []byte) (n int, err error) {
	msg := string(p)
	// stdlib includes a trailing newline on the msg that
	// we do not want.
	msg = strings.TrimSuffix(msg, "\n")

	Info(w.ctx, msg)

	return len(p), nil
}
