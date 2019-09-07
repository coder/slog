package slog

import (
	"context"
	"os"

	"go.coder.com/slog/internal/skipctx"
	"go.coder.com/slog/slogcore"
)

// Tee enables logging to multiple loggers.
// Does not support the logger returned by Test.
func Tee(loggers ...Logger) Logger {
	return multiLogger{
		loggers: loggers,
	}
}

type multiLogger struct {
	loggers []Logger
}

func (l multiLogger) Debug(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, slogcore.Debug, msg, fields)
}

func (l multiLogger) Info(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, slogcore.Info, msg, fields)
}

func (l multiLogger) Warn(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, slogcore.Warn, msg, fields)
}

func (l multiLogger) Error(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, slogcore.Error, msg, fields)
}

func (l multiLogger) Critical(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, slogcore.Critical, msg, fields)
}

func (l multiLogger) Fatal(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, slogcore.Fatal, msg, fields)
}

func (l multiLogger) log(ctx context.Context, level slogcore.Level, msg string, fields []Field) {
	ctx = skipctx.With(ctx, 2)
	for _, l := range l.loggers {
		switch level {
		case slogcore.Debug:
			l.Debug(ctx, msg, fields...)
		case slogcore.Info:
			l.Info(ctx, msg, fields...)
		case slogcore.Warn:
			l.Warn(ctx, msg, fields...)
		case slogcore.Error:
			l.Error(ctx, msg, fields...)
		case slogcore.Critical, slogcore.Fatal:
			l.Critical(ctx, msg, fields...)
		}
	}

	if level == slogcore.Fatal {
		os.Exit(1)
	}
}

func (l multiLogger) With(fields ...Field) Logger {
	loggers2 := make([]Logger, len(l.loggers))

	for i, l := range l.loggers {
		loggers2[i] = l.With(fields...)
	}

	return multiLogger{
		loggers: loggers2,
	}
}
