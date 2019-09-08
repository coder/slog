package stdlibctx

import (
	"context"
)

type stdlibKey struct{}

// With appends to the skip offset in the context and returns
// a new context with the new skip offset.
func With(ctx context.Context) context.Context {
	return context.WithValue(ctx, stdlibKey{}, true)
}

// From returns the skip offset.
func From(ctx context.Context) bool {
	b, _ := ctx.Value(stdlibKey{}).(bool)
	return b
}
