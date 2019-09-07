package humanfmt

import (
	"testing"

	"go.coder.com/slog/internal/diff"
	"go.coder.com/slog/slogcore"
)

func Test_marshalFields(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   map[string]interface{}
		out  string
	}{
		{
			name: "stringWithNewlines",
			in: map[string]interface{}{
				"a": `hi
two
three`,
			},
			out: `a: hi
  two
  three`,
		},
		{
			name: "bool",
			in: map[string]interface{}{
				"a": false,
			},
			out: `a: false`,
		},
		{
			name: "float",
			in: map[string]interface{}{
				"a": 0.3,
			},
			out: `a: 0.3`,
		},
		{
			name: "int",
			in: map[string]interface{}{
				"a": -1,
			},
			out: `a: -1`,
		},
		{
			name: "uint",
			in: map[string]interface{}{
				"a": uint(3),
			},
			out: `a: 3`,
		},
		{
			name: "list",
			in: map[string]interface{}{
				"a": []interface{}{
					map[string]interface{}{
						"hi":  "hello",
						"hi3": "hello",
					},
					"3",
					[]string{"a", "b", "c"},
				},
			},
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
			in: map[string]interface{}{
				"a": struct{}{},
				"b": struct{}{},
				"c": struct{}{},
			},
			out: `a:
b:
c:`,
		},
		{
			name: "nestedMap",
			in: map[string]interface{}{
				"a": map[string]string{
					"0": "hi",
					"1": "hi",
				},
			},
			out: `a:
  0: hi
  1: hi`,
		},
		{
			name: "specialCharacterKey",
			in: map[string]interface{}{
				"nhooyr \tsoftware™️": "hi",
				"\rxeow\r": `mdsla
dsamkld`,
			},
			out: `"\rxeow\r": mdsla
  dsamkld
"nhooyr_\tsoftware™️": hi`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			v := slogcore.Reflect(tc.in)
			actOut := Fields(v.(slogcore.Map))
			t.Logf("yaml:\n%v", actOut)
			if diff := diff.Diff(tc.out, actOut); diff != "" {
				t.Fatalf("unexpected output: %v", diff)
			}
		})
	}
}
