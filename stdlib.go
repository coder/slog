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
	l.skip += 4

	l = l.clone()
	for i, s := range l.sinks {
		if ts, ok := s.sink.(testSink); ok {
			l.sinks[i].sink = ts.Stdlib()
		}
	}

	w := &stdlogWriter{
		Log: func(msg string) {
			l.Info(ctx, msg)
		},
	}

	return log.New(w, "", 0)
}

type stdlogWriter struct {
	Log func(msg string)
}

func (w stdlogWriter) Write(p []byte) (n int, err error) {
	msg := string(p)
	// stdlib includes a trailing newline on the msg but we will
	// insert it later in the string method of the entry so
	// we do not want it here.
	msg = strings.TrimSuffix(msg, "\n")

	w.Log(msg)

	return len(p), nil
}
