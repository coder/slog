package slog_test

import (
	"testing"

	"golang.org/x/xerrors"

	"go.coder.com/slog"
	"go.coder.com/slog/slogval"
)

func TestEncode(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   interface{}
	}{
		{
			name: "unexportedErrorField",
			in: struct {
				err error
			}{
				err: xerrors.Errorf("wow"),
			},
		},
		{
			name: "unexported[]slog.Field",
			in: struct {
				fields []slog.Field
			}{
				fields: slog.Map(slog.F("wow", "two")),
			},
		},
		{
			name: "unexported_slogval.Map",
			in: struct {
				fields slogval.Map
			}{
				fields: slogval.Map{
					slogval.Field{
						Name:  "wow",
						Value: slogval.String("meow"),
					},
				},
			},
		}, {
			name: "embeddedFields",
			in: struct {
				string
				int
				float32
				*Meow
			}{"meow", 3, 4, &Meow{"wow"}},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Log(slog.Encode(tc.in))
		})
	}
}

type Meow struct {
	meow string
}
