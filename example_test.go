package slog_test

import (
	"io"
	"testing"

	"golang.org/x/xerrors"

	"cdr.dev/slog"
	"cdr.dev/slog/sloggers/slogtest"
)

func Example_slogtest() {
	// Would be provided by the testing framework.
	var t testing.TB

	slogtest.Info(t, "my message here",
		slog.F("field_name", "something or the other"),
		slog.F("some_map", slog.M(
			slog.F("nested_fields", "wowow"),
		)),
		slog.Error(
			xerrors.Errorf("wrap1: %w",
				xerrors.Errorf("wrap2: %w",
					io.EOF,
				),
			),
		),
	)

	// t.go:55: 2019-12-05 21:20:31.218 [INFO]	<examples_test.go:42>	my message here	{"field_name": "something or the other", "some_map": {"nested_fields": "wowow"}} ...
	//    "error": wrap1:
	//        cdr.dev/slog_test.TestExample
	//            /Users/nhooyr/src/cdr/slog/examples_test.go:48
	//      - wrap2:
	//        cdr.dev/slog_test.TestExample
	//            /Users/nhooyr/src/cdr/slog/examples_test.go:49
	//      - EOF
}

func TestExample(t *testing.T) {
	t.Parallel()

	slogtest.Info(t, "my message here",
		slog.F("field_name", "something or the other"),
		slog.F("some_map", slog.M(
			slog.F("nested_fields", "wowow"),
		)),
		slog.Error(
			xerrors.Errorf("wrap1: %w",
				xerrors.Errorf("wrap2: %w",
					io.EOF,
				),
			),
		),
	)
}
