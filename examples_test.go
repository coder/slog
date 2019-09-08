package slog_test

import (
	"golang.org/x/xerrors"
	"io"
	"testing"

	"go.coder.com/slog"
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
		slog.Component("test"),
	)

	// --- PASS: TestExample (0.00s)
	//    examples_test.go:46: Sep 08 13:54:34.532 [INFO] (test): my_message_here
	//        field_name: something or the other
	//        some_map:
	//          nested_fields: wowow
	//        error:
	//          - wrap1
	//            go.coder.com/slog_test.TestExample
	//              /Users/nhooyr/src/cdr/slog/examples_test.go:52
	//          - wrap2
	//            go.coder.com/slog_test.TestExample
	//              /Users/nhooyr/src/cdr/slog/examples_test.go:53
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
		slog.Component("test"),
	)
}
