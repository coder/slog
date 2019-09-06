package slog

import (
	"context"
	"go.coder.com/slog/internal/skipctx"
	"os"
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
	l.log(ctx, levelDebug, msg, fields)
}

func (l multiLogger) Info(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, levelInfo, msg, fields)
}

func (l multiLogger) Warn(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, levelWarn, msg, fields)
}

func (l multiLogger) Error(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, levelError, msg, fields)
}

func (l multiLogger) Critical(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, levelCritical, msg, fields)
}

func (l multiLogger) Fatal(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, levelDebug, msg, fields)
}

func (l multiLogger) log(ctx context.Context, level level, msg string, fields []Field) {
	ctx = skipctx.With(ctx, 2)
	for _, l := range l.loggers {
		switch level {
		case levelDebug:
			l.Debug(ctx, msg, fields...)
		case levelInfo:
			l.Info(ctx, msg, fields...)
		case levelWarn:
			l.Warn(ctx, msg, fields...)
		case levelError:
			l.Error(ctx, msg, fields...)
		case levelCritical, levelFatal:
			l.Critical(ctx, msg, fields...)
		}
	}

	if level == levelFatal {
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
