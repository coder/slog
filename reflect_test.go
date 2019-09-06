package slog

import (
	"fmt"
	"golang.org/x/xerrors"
	"io"
	"reflect"
	"runtime"
	"testing"
)

func Test_reflectValue(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   interface{}
		out  fieldValue
	}{
		{
			name: "xerror",
			in: xerrors.Errorf("wrap msg: %w",
				xerrors.Errorf("hi: %w", io.EOF),
			),
			out: fieldList{
				fieldMap{
					{"msg", fieldString("wrap msg")},
					{"loc", fieldString(location(0, -6))},
					{"fun", fieldString("go.coder.com/slog.Test_reflectValue")},
				},
				fieldMap{
					{"msg", fieldString("hi")},
					{"loc", fieldString(location(0, -10))},
					{"fun", fieldString("go.coder.com/slog.Test_reflectValue")},
				},
				fieldString("EOF"),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actOut := reflectFieldValue(reflect.ValueOf(tc.in))
			if diff := cmpDiff(tc.out, actOut); diff != "" {
				t.Fatalf("unexpected output: %v", diff)
			}
		})
	}
}

func location(skip int, lineOffset int) string {
	_, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		panicf("failed to get caller information with skip %v", skip)
	}
	return fmt.Sprintf("%v:%v", file, line+lineOffset)
}
