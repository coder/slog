package slog

import "context"

func SetExit(ctx context.Context, fn func(int)) context.Context {
	l, ok := extractContext(ctx)
	if !ok {
		return ctx
	}
	l.exit = fn
	return makeContext(ctx, l)
}
