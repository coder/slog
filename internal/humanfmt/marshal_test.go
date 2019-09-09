package humanfmt

import (
	"testing"

	"go.coder.com/slog"
	"go.coder.com/slog/internal/assert"
	"go.coder.com/slog/slogval"
)

func Test_marshalFields(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   []slog.Field
		out  string
	}{
		{
			name: "stringWithNewlines",
			in: slog.Map(
				slog.F("a", `hi
two
three`),
			),
			out: `a: hi
  two
  three`,
		},
		{
			name: "bool",
			in: slog.Map(
				slog.F("a", false),
			),
			out: `a: false`,
		},
		{
			name: "float",
			in: slog.Map(
				slog.F("a", 0.3),
			),
			out: `a: 0.3`,
		},
		{
			name: "int",
			in: slog.Map(
				slog.F("a", -1),
			),
			out: `a: -1`,
		},
		{
			name: "uint",
			in: slog.Map(
				slog.F("a", uint(3)),
			),
			out: `a: 3`,
		},
		{
			name: "list",
			in: slog.Map(
				slog.F("a", []interface{}{
					slog.Map(
						slog.F("hi", "hello"),
						slog.F("hi3", "hello"),
					),
					"3",
					[]string{"a", "b", "c"},
				},
				)),
			out: `a:
  - hi: hello
    hi3: hello
  - 3
  -
    - a
    - b
    - c`,
		},
		{
			name: "emptyStruct",
			in: slog.Map(
				slog.F("a", struct{}{}),
				slog.F("b", struct{}{}),
				slog.F("c", struct{}{}),
			),
			out: `a:
b:
c:`,
		},
		{
			name: "nestedMap",
			in: slog.Map(
				slog.F("a", map[string]string{
					"1": "hi",
					"0": "hi",
				}),
			),
			out: `a:
  0: hi
  1: hi`,
		},
		{
			name: "specialCharacterKey",
			in: slog.Map(
				slog.F("nhooyr \tsoftware™️", "hi"),
				slog.F("\rxeow\r", `mdsla
dsamkld`),
			),
			out: `"nhooyr_\tsoftware™️": hi
"\rxeow\r": mdsla
  dsamkld`,
		},
		{
			name: "specialCharacterKey",
			in: slog.Map(
				slog.F("nhooyr \tsoftware™️", "hi"),
				slog.F("\rxeow\r", `mdsla
dsamkld`),
			),
			out: `"nhooyr_\tsoftware™️": hi
"\rxeow\r": mdsla
  dsamkld`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			v := slogval.Encode(tc.in).(slogval.Map)
			actOut := fmtVal(v)
			t.Logf("yaml:\n%v", actOut)
			assert.Equalf(t, tc.out, actOut, "unexpected output")
		})
	}
}
