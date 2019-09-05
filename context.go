package core

import (
	"context"

	"go.coder.com/m/lib/log"
	"go.coder.com/m/lib/log/internal/logctx"
)

type (
	loggerKey struct{}
	skipKey   struct{}
	stdlibKey struct{}
)

func withContext(ctx context.Context, l Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, l)
}

func fromContext(ctx context.Context) Logger {
	l, _ := ctx.Value(loggerKey{}).(Logger)
	return l
}

func With(ctx context.Context, fields log.F) context.Context {
	l := fromContext(ctx)
	l = l.With(fields)
	return withContext(ctx, l)
}

func init() {
	logctx.With = func(ctx context.Context, fields map[string]interface{}) context.Context {
		return With(ctx, fields)
	}
	logctx.WithSkip = WithSkip
	logctx.WithStdlib = WithStdlib
}

func WithSkip(ctx context.Context, skip int) context.Context {
	skip += skipFrom(ctx)
	return context.WithValue(ctx, skipKey{}, skip)
}

func skipFrom(ctx context.Context) int {
	l, _ := ctx.Value(skipKey{}).(int)
	return l
}

func WithStdlib(ctx context.Context) context.Context {
	return context.WithValue(ctx, stdlibKey{}, true)
}

func IsStdlib(ctx context.Context) bool {
	ok, _ := ctx.Value(stdlibKey{}).(bool)
	return ok
}
