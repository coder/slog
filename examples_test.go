package slog_test

import (
	"context"
	"golang.org/x/xerrors"
	"io"
	"os"
	"testing"

	"go.coder.com/slog"
	"go.coder.com/slog/sloggers/slogjson"
	"go.coder.com/slog/sloggers/slogtest"
)

func Example_test() {
	// Nil here but would be provided by the testing framework.
	var t testing.TB

	slogtest.Info(t, "my message here",
		slog.F("field_name", "something or the other"),
		slog.F("some_map", slog.Map(
			slog.F("nested_fields", "wowow"),
		)),
		slog.Error(
			xerrors.Errorf("wrap1: %w",
				xerrors.Errorf("wrap2: %w",
					io.EOF),
			)),
	)

	//     t.go:55: 2019-09-13 19:27:15.336 [INFO]	<examples_test.go:47>	my message here	{"field_name": "something or the other", "some_map": {"nested_fields": "wowow"}} ...
	//        "error": wrap1:
	//            go.coder.com/slog_test.TestExample
	//                /Users/nhooyr/src/cdr/slog/examples_test.go:53
	//          - wrap2:
	//            go.coder.com/slog_test.TestExample
	//                /Users/nhooyr/src/cdr/slog/examples_test.go:54
	//          - EOF
}

func TestExample(t *testing.T) {
	slogtest.Info(t, "my message here",
		slog.F("field_name", "something or the other"),
		slog.F("some_map", slog.Map(
			slog.F("nested_fields", "wowow"),
		)),
		slog.Error(
			xerrors.Errorf("wrap1: %w",
				xerrors.Errorf("wrap2: %w",
					io.EOF),
			)),
	)
}

func TestJSON(t *testing.T) {
	l := slogjson.Make(os.Stderr)
	l.Info(context.Background(), "my message\r here",
		slog.F("field_name", "something or the other"),
		slog.F("some_map", slog.Map(
			slog.F("nested_fields", "wowow"),
		)),
		slog.Error(
			xerrors.Errorf("wrap1: %w",
				xerrors.Errorf("wrap2: %w",
					io.EOF),
			)),
	)

	slog.Stdlib(context.Background(), l).Println("hi\nmeow")
}
