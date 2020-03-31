package slog_test

import (
	"bytes"
	"context"
	"testing"

	"cdr.dev/slog"
	"cdr.dev/slog/internal/assert"
	"cdr.dev/slog/internal/entryhuman"
	"cdr.dev/slog/sloggers/sloghuman"
)

func TestStdlib(t *testing.T) {
	t.Parallel()

	b := &bytes.Buffer{}
	ctx := context.Background()
	ctx = slog.Make(sloghuman.Make(ctx, b))
	ctx = slog.With(ctx,
		slog.F("hi", "we"),
	)
	stdlibLog := slog.Stdlib(ctx)
	stdlibLog.Println("stdlib")

	et, rest, err := entryhuman.StripTimestamp(b.String())
	assert.Success(t, "strip timestamp", err)
	assert.False(t, "timestamp", et.IsZero())
	assert.Equal(t, "entry", " [INFO]\t(stdlib)\t<s_test.go:24>\tstdlib\t{\"hi\": \"we\"}\n", rest)
}
