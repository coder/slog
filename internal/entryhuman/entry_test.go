package entryhuman_test

import (
	"io/ioutil"
	"testing"
	"time"

	"go.opencensus.io/trace"

	"cdr.dev/slog"
	"cdr.dev/slog/internal/assert"
	"cdr.dev/slog/internal/entryhuman"
)

var kt = time.Date(2000, time.February, 5, 4, 4, 4, 4, time.UTC)

func TestEntry(t *testing.T) {
	t.Parallel()

	test := func(t *testing.T, in slog.SinkEntry, exp string) {
		act := entryhuman.Fmt(ioutil.Discard, in)
		assert.Equal(t, exp, act, "entry")
	}

	t.Run("basic", func(t *testing.T) {
		t.Parallel()

		test(t, slog.SinkEntry{
			Message: "wowowow\tizi",
			Time:    kt,
			Level:   slog.LevelDebug,

			File: "myfile",
			Line: 100,
			Func: "ignored",
		}, `2000-02-05 04:04:04.000 [DEBUG]	<myfile:100>	"wowowow\tizi"`)
	})

	t.Run("multilineMessage", func(t *testing.T) {
		t.Parallel()

		test(t, slog.SinkEntry{
			Message: "line1\nline2",
			Level:   slog.LevelInfo,
		}, `0001-01-01 00:00:00.000 [INFO]	<.:0>	...
"msg": line1
line2`)
	})

	t.Run("named", func(t *testing.T) {
		t.Parallel()

		test(t, slog.SinkEntry{
			Level:      slog.LevelWarn,
			LoggerName: "named.meow",
		}, `0001-01-01 00:00:00.000 [WARN]	(named.meow)	<.:0>	""`)
	})

	t.Run("trace", func(t *testing.T) {
		t.Parallel()

		test(t, slog.SinkEntry{
			Level: slog.LevelError,
			SpanContext: trace.SpanContext{
				SpanID:  trace.SpanID{0, 1, 2, 3, 4, 5, 6, 7},
				TraceID: trace.TraceID{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
			},
		}, `0001-01-01 00:00:00.000 [ERROR]	<.:0>	""	{"trace": "000102030405060708090a0b0c0d0e0f", "span": "0001020304050607"}`)
	})

	t.Run("color", func(t *testing.T) {
		t.Parallel()

		act := entryhuman.Fmt(entryhuman.ForceColorWriter, slog.SinkEntry{
			Level: slog.LevelCritical,
			Fields: slog.M(
				slog.F("hey", "hi"),
			),
		})
		assert.Equal(t, "0001-01-01 00:00:00.000 \x1b[91m[CRITICAL]\x1b[0m\t\x1b[36m<.:0>\x1b[0m\t\"\"\t{\x1b[34m\"hey\"\x1b[0m: \x1b[32m\"hi\"\x1b[0m}", act, "entry")
	})
}
