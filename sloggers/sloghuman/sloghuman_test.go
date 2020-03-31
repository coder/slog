package sloghuman_test

import (
	"bytes"
	"context"
	"testing"

	"cdr.dev/slog/v2"
	"cdr.dev/slog/v2/internal/assert"
	"cdr.dev/slog/v2/internal/entryhuman"
	"cdr.dev/slog/v2/sloggers/sloghuman"
)

func TestMake(t *testing.T) {
	t.Parallel()

	b := &bytes.Buffer{}
	ctx := context.Background()
	ctx = sloghuman.Make(ctx, b)
	slog.Info(ctx, "line1\n\nline2", slog.F("wowow", "me\nyou"))
	slog.Sync(ctx)

	et, rest, err := entryhuman.StripTimestamp(b.String())
	assert.Success(t, "strip timestamp", err)
	assert.False(t, "timestamp", et.IsZero())
	assert.Equal(t, "entry", " [INFO]\t<sloghuman_test.go:20>\t...\t{\"wowow\": \"me\\nyou\"}\n  \"msg\": line1\n\n         line2\n", rest)
}
