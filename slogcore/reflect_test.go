package slogcore

import (
	"fmt"
	"io"
	"reflect"
	"runtime"
	"testing"

	"golang.org/x/xerrors"

	"go.coder.com/slog/internal/diff"
)

func Test_reflectValue(t *testing.T) {
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
				Map{
					{"msg", String("wrap msg")},
					{"loc", String(testLocation(0, -6))},
					{"fun", String("go.coder.com/slog.Test_reflectValue")},
				},
				Map{
					{"msg", String("hi")},
					{"loc", String(testLocation(0, -10))},
					{"fun", String("go.coder.com/slog.Test_reflectValue")},
				},
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
			name: "LogValue",
			in:   myStruct{},
			out:  String("hi"),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actOut := reflectValue(reflect.ValueOf(tc.in))
			if diff := diff.Diff(tc.out, actOut); diff != "" {
				t.Fatalf("unexpected output: %v", diff)
			}
		})
	}
}

func testLocation(skip int, lineOffset int) string {
	_, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		panicf("failed to get caller information with skip %v", skip)
	}
	return fmt.Sprintf("%v:%v", file, line+lineOffset)
}

type myStruct struct{}

func (m myStruct) LogValue() interface{} {
	return "hi"
}
