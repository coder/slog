// Package skipctx contains helpers to put the frame skip level
// into the context. Used by stderrlog and testlog helpers to
// skip one more frame level.
package skipctx

import (
	"context"
)

type skipKey struct{}

func With(ctx context.Context, skip int) context.Context {
	skip += From(ctx)
	return context.WithValue(ctx, skipKey{}, skip)
}

func From(ctx context.Context) int {
	l, _ := ctx.Value(skipKey{}).(int)
	return l
}
