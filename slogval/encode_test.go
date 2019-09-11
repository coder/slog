package slogval

import (
	"fmt"
	"io"
	"runtime"
	"testing"

	"golang.org/x/xerrors"

	"go.coder.com/slog/internal/assert"
)

func TestEncode(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   interface{}
		out  Value
	}{
		{
			name: "xerror",
			in: xerrors.Errorf("wrap msg: %w",
				xerrors.Errorf("hi: %w", io.EOF),
			),
			out: List{
				String(`wrap msg
go.coder.com/slog/slogval.TestEncode
  ` + testLocation(0, -6),
				),
				String(`hi
go.coder.com/slog/slogval.TestEncode
  ` + testLocation(0, -9),
				),
				String("EOF"),
			},
		},
		{
			name: "logTag",
			in: struct {
				a string `log:"-"`
				b string `log:"hi"`
				c string `log:"f"`
			}{
				"a",
				"b",
				"c",
			},
			out: Map{
				{"hi", String("b")},
				{"f", String("c")},
			},
		},
		{
			name: "logValue",
			in:   myStruct{},
			out:  String("hi"),
		},
		{
			name: "embeddedStruct",
			in:   outerStruct{},
			out: Map{
				{"field_3", Int(0)},
				{"field_5", Int(0)},
				{"field_1", String("")},
				{"field_2", Int(0)},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actOut := Encode(tc.in, nil)
			assert.Equalf(t, tc.out, actOut, "unexpected output")
		})
	}
}

func testLocation(skip int, lineOffset int) string {
	_, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		panic("failed to get caller information")
	}
	return fmt.Sprintf("%v:%v", file, line+lineOffset)
}

type myStruct struct{}

func (m myStruct) LogValue() interface{} {
	return "hi"
}

type outerStruct struct {
	innerStruct

	field1 string
	field2 int
}

type innerStruct struct {
	field3 int
	field5 int
}
