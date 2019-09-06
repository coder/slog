package slog_test

import (
	"context"
	"io"
	"testing"

	"golang.org/x/xerrors"

	"go.coder.com/slog"
	"go.coder.com/slog/stderrlog"
	"go.coder.com/slog/testlog"
)

func TestExampleStderr(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	stderrlog.Info(ctx, "my message here",
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

		slog.F("name", slog.ValueFunc(func() interface{} {
			return "wow"
		})),
	)
}

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

		slog.F("name", slog.ValueFunc(func() interface{} {
			return "wow"
		})),
	)
}
