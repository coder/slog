package slogval_test

import (
	"fmt"
	"io"
	"runtime"
	"testing"

	"golang.org/x/xerrors"

	"go.coder.com/slog/internal/diff"
	"go.coder.com/slog/slogval"
)

func TestReflect(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   interface{}
		out  slogval.Value
	}{
		{
			name: "xerror",
			in: xerrors.Errorf("wrap msg: %w",
				xerrors.Errorf("hi: %w", io.EOF),
			),
			out: slogval.List{
				slogval.String(`wrap msg
go.coder.com/slog/slogval_test.TestReflect
  ` + testLocation(0, -6),
				),
				slogval.String(`hi
go.coder.com/slog/slogval_test.TestReflect
  ` + testLocation(0, -9),
				),
				slogval.String("EOF"),
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
			out: slogval.Map{
				{"hi", slogval.String("b")},
				{"f", slogval.String("c")},
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
			out: slogval.Map{
				{"hi", slogval.String("b")},
				{"f", slogval.String("c")},
			},
		},
		{
			name: "LogValue",
			in:   myStruct{},
			out:  slogval.String("hi"),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actOut := slogval.ReflectUnsafe(tc.in)
			if diff := diff.Diff(tc.out, actOut); diff != "" {
				t.Fatalf("unexpected output: %v", diff)
			}
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
