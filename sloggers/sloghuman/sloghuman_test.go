package sloghuman_test

import (
	"bytes"
	"context"
	"testing"

	"cdr.dev/slog"
	"cdr.dev/slog/internal/assert"
	"cdr.dev/slog/internal/slogfmt"
	"cdr.dev/slog/sloggers/sloghuman"
)

var bg = context.Background()

func TestMake(t *testing.T) {
	t.Parallel()

	b := &bytes.Buffer{}
	l := sloghuman.Make(b)
	l.Info(bg, "line1\n\nline2", slog.F("wowow", "me\nyou"))
	l.Sync()

	et, rest, err := slogfmt.StripTimestamp(b.String())
	assert.Success(t, err, "strip timestamp")
	assert.False(t, et.IsZero(), "timestamp")
	assert.Equal(t, " [INFO]\t<sloghuman_test.go:21>\t...\t{\"wowow\": \"me\\nyou\"}\n  \"msg\": line1\n\n  line2\n", rest, "entry")
}
