package slog

import "context"

type loggerCtxKey = struct{}

type sinkContext struct {
	context.Context
	Sink
}

func makeContext(ctx context.Context, l logger) SinkContext {
	ctx = context.WithValue(ctx, loggerCtxKey{}, l)
	return &sinkContext{
		Context: ctx,
		Sink:    l,
	}
}

func extractContext(ctx context.Context) (logger, bool) {
	v := ctx.Value(loggerCtxKey{})
	if v == nil {
		return logger{}, false
	}
	return v.(logger), true
}
