package sloghuman_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
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
	assert.Equal(t, "entry", " [info]  ...  wowow=\"me\\nyou\"\n    msg= line1\n\n         line2\n", rest)
}

func TestVisual(t *testing.T) {
	t.Setenv("FORCE_COLOR", "true")
	if os.Getenv("TEST_VISUAL") == "" {
		t.Skip("TEST_VISUAL not set")
	}

	l := slog.Make(sloghuman.Sink(os.Stdout)).Leveled(slog.LevelDebug)
	l.Debug(bg, "small potatos", slog.F("aaa", "mmm"), slog.F("bbb", "nnn"), slog.F("age", 24))
	l.Info(bg, "line1\n\nline2", slog.F("wowow", "me\nyou"))
	l.Warn(bg, "oops", slog.F("aaa", "mmm"))
	l = l.Named("sublogger")
	l.Error(bg, "big oops", slog.F("aaa", "mmm"), slog.Error(fmt.Errorf("this happened\nand this")))
	l.Sync()
}
