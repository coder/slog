package sloghuman_test

import (
	"bytes"
	"context"
	"testing"

	"cdr.dev/slog"
	"cdr.dev/slog/internal/assert"
	"cdr.dev/slog/internal/entryhuman"
	"cdr.dev/slog/sloggers/sloghuman"
)

var bg = context.Background()

func TestMake(t *testing.T) {
	t.Parallel()

	b := &bytes.Buffer{}
	l := slog.Make(sloghuman.Sink(b))
	l.Info(bg, "line1\n\nline2", slog.F("wowow", "me\nyou"))
	l.Sync()

	et, rest, err := entryhuman.StripTimestamp(b.String())
	assert.Success(t, "strip timestamp", err)
	assert.False(t, "timestamp", et.IsZero())
	assert.Equal(t, "entry", " [INFO]\t<cdr.dev/slog/sloggers/sloghuman_test/sloghuman_test.go:21>\tTestMake\t...\t{\"wowow\": \"me\\nyou\"}\n  \"msg\": line1\n\n         line2\n", rest)
}
