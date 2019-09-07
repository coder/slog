// Package skipctx contains helpers to put the frame skip level
// into the context. Used by stderrlog and testslog helpers to
// skip one more frame level.
package skipctx

import (
	"context"
)

type skipKey struct{}

// With appends to the skip offset in the context and returns
// a new context with the new skip offset.
func With(ctx context.Context, skip int) context.Context {
	skip += From(ctx)
	return context.WithValue(ctx, skipKey{}, skip)
}

// From returns the skip offset.
func From(ctx context.Context) int {
	l, _ := ctx.Value(skipKey{}).(int)
	return l
}
