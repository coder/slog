package slog

import (
	"context"
	"os"
)

// DefaultLogger is a default logger and is used by package level logging functions Debug, Info, Warn, Err, Critical and Fatal
// DefaultLogger does not have any sinks by default
var DefaultLogger Logger = Logger{
	sinks: make([]Sink, 0),
	level: LevelDebug,
	exit:  os.Exit,
}

// Debug logs the msg and fields at LevelDebug with DefaultLogger.
func Debug(ctx context.Context, msg string, fields ...Field) {
	DefaultLogger.log(ctx, LevelDebug, msg, fields)
}

// Info logs the msg and fields at LevelInfo with DefaultLogger.
func Info(ctx context.Context, msg string, fields ...Field) {
	DefaultLogger.log(ctx, LevelInfo, msg, fields)
}

// Warn logs the msg and fields at LevelWarn with DefaultLogger.
func Warn(ctx context.Context, msg string, fields ...Field) {
	DefaultLogger.log(ctx, LevelWarn, msg, fields)
}

// Err logs the msg and fields at LevelError with DefaultLogger.
//
// It will then Sync().
func Err(ctx context.Context, msg string, fields ...Field) {
	DefaultLogger.log(ctx, LevelError, msg, fields)
	DefaultLogger.Sync()
}

// Critical logs the msg and fields at LevelCritical with DefaultLogger.
//
// It will then Sync().
func Critical(ctx context.Context, msg string, fields ...Field) {
	DefaultLogger.log(ctx, LevelCritical, msg, fields)
	DefaultLogger.Sync()
}

// Fatal logs the msg and fields at LevelFatal with DefaultLogger.
//
// It will then Sync() and os.Exit(1).
func Fatal(ctx context.Context, msg string, fields ...Field) {
	DefaultLogger.log(ctx, LevelFatal, msg, fields)
	DefaultLogger.Sync()

	if DefaultLogger.exit == nil {
		DefaultLogger.exit = defaultExitFn
	}

	DefaultLogger.exit(1)
}
