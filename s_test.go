package slog_test

import (
	"bytes"
	"testing"

	"cdr.dev/slog"
	"cdr.dev/slog/internal/assert"
	"cdr.dev/slog/internal/entryhuman"
	"cdr.dev/slog/sloggers/sloghuman"
)

func TestStdlib(t *testing.T) {
	t.Parallel()

	b := &bytes.Buffer{}
	l := slog.Make(sloghuman.Sink(b)).With(
		slog.F("hi", "we"),
	)
	stdlibLog := slog.Stdlib(bg, l, slog.LevelInfo)
	stdlibLog.Println("stdlib")

	et, rest, err := entryhuman.StripTimestamp(b.String())
	assert.Success(t, "strip timestamp", err)
	assert.False(t, "timestamp", et.IsZero())
	assert.Equal(t, "entry", " [INFO]\t(stdlib)\t<cdr.dev/slog_test/s_test.go:21>\tTestStdlib\tstdlib\t{\"hi\": \"we\"}\n", rest)
}
