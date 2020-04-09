package slog

import "context"

type loggerCtxKey = struct{}

// SinkContext is used by slog.Make to compose many loggers together.
type SinkContext struct {
	Sink
	context.Context
}

func contextWithLogger(ctx context.Context, l logger) SinkContext {
	ctx = context.WithValue(ctx, loggerCtxKey{}, l)
	return SinkContext{
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
