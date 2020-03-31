package slog

import "context"

type loggerCtxKey = struct{}

type sinkContext struct {
	context.Context
	Sink
}

// SinkContext is a context that implements Sink.
// It may be returned by log creators to allow for composition.
type SinkContext interface {
	Sink
	context.Context
}

func contextWithLogger(ctx context.Context, l logger) SinkContext {
	ctx = context.WithValue(ctx, loggerCtxKey{}, l)
	return &sinkContext{
		Context: ctx,
		Sink:    l,
	}
}

func loggerFromContext(ctx context.Context) (logger, bool) {
	v := ctx.Value(loggerCtxKey{})
	if v == nil {
		return logger{}, false
	}
	return v.(logger), true
}
