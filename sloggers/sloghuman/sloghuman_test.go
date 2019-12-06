package sloghuman_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"cdr.dev/slog"
	"cdr.dev/slog/internal/assert"
	"cdr.dev/slog/internal/humanfmt"
	"cdr.dev/slog/sloggers/sloghuman"
)

var bg = context.Background()

func TestMake(t *testing.T) {
	t.Parallel()

	b := &bytes.Buffer{}
	l := sloghuman.Make(b)
	l.Info(bg, "line1\n\nline2", slog.F("wowow", "me\nyou"))
	l.Sync()

	s := b.String()
	ts := s[:len(humanfmt.TimeFormat)]
	rest := s[len(humanfmt.TimeFormat):]

	et, err := time.Parse(humanfmt.TimeFormat, ts)
	assert.Success(t, err, "time.Parse")
	assert.False(t, et.IsZero(), "timestamp")
	assert.Equal(t, " [INFO]\t<sloghuman_test.go:22>\t...\t{\"wowow\": \"me\\nyou\"}\n  \"msg\": line1\n\n  line2\n", rest, "entry")
}
