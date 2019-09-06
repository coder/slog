package slog_test

import (
	"io"
	"testing"

	"golang.org/x/xerrors"

	"go.coder.com/slog"
	"go.coder.com/slog/testlog"
)

func TestExampleTest(t *testing.T) {
	t.Parallel()

	testlog.Info(t, "my message here",
		slog.F("field_name", "something or the other"),
		slog.F("some_map", map[string]interface{}{
			"nested_fields": "wowow",
		}),
		slog.Error(xerrors.Errorf("wrap2: %w",
			xerrors.Errorf("wrap1: %w",
				io.EOF,
			),
		)),
		slog.Component("test"),
	)
}
