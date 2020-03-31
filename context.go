package slog

import "context"

type loggerCtxKey = struct{}

func contextWithLogger(ctx context.Context, l logger) context.Context {
	return context.WithValue(ctx, loggerCtxKey{}, l)
}

func loggerFromContext(ctx context.Context) (logger, bool) {
	v := ctx.Value(loggerCtxKey{})
	if v == nil {
		return logger{}, false
	}
	return v.(logger), true
}
