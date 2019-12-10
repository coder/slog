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
	l := slog.Make(sloghuman.Make(b)).With(
		slog.F("hi", "we"),
	)
	stdlibLog := slog.Stdlib(bg, l)
	stdlibLog.Println("stdlib")

	et, rest, err := entryhuman.StripTimestamp(b.String())
	assert.Success(t, err, "strip timestamp")
	assert.False(t, et.IsZero(), "timestamp")
	assert.Equal(t, " [INFO]\t(stdlib)\t<s_test.go:21>\tstdlib\t{\"hi\": \"we\"}\n", rest, "entry")
}
