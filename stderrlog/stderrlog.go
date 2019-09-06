// Package stderrlog is a helper for logging to stderr.
// It will also set the stdlib log package to log to stderr.
package stderrlog

import (
	"context"
	"go.coder.com/slog"
	"go.coder.com/slog/internal/skipctx"
	"log"
	"os"
)

var stderrLogger = slog.Make(os.Stderr)

// Redirect all stdlib logs to our default logger.
func init() {
	l := slog.Stdlib(context.Background(), stderrLogger)
	log.SetOutput(l.Writer())
}

// Debug logs the given msg and fields on the debug level to stderr.
func Debug(ctx context.Context, msg string, fields ...slog.Field) {
	stderrLogger.Debug(skipctx.With(ctx, 1), msg, fields...)
}

// Info logs the given msg and fields on the info level to stderr.
func Info(ctx context.Context, msg string, fields ...slog.Field) {
	stderrLogger.Info(skipctx.With(ctx, 1), msg, fields...)
}

// Warn logs the given msg and fields on the warn level to stderr.
func Warn(ctx context.Context, msg string, fields ...slog.Field) {
	stderrLogger.Warn(skipctx.With(ctx, 1), msg, fields...)
}

// Error logs the given msg and fields on the error level to stderr.
func Error(ctx context.Context, msg string, fields ...slog.Field) {
	stderrLogger.Error(skipctx.With(ctx, 1), msg, fields...)
}

// Critical logs the given msg and fields on the critical level to stderr.
func Critical(ctx context.Context, msg string, fields ...slog.Field) {
	stderrLogger.Critical(skipctx.With(ctx, 1), msg, fields...)
}

// Fatal logs the given msg and fields on the fatal level to stderr.
func Fatal(ctx context.Context, msg string, fields ...slog.Field) {
	stderrLogger.Fatal(skipctx.With(ctx, 1), msg, fields...)
}
