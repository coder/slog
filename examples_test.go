package slog_test

import (
	"context"
	"go.coder.com/slog"
	"go.coder.com/slog/stderrlog"
	"go.coder.com/slog/testlog"
	"testing"
)

func Example_stderr() {
	ctx := context.Background()
	stderrlog.Info(ctx, "my message here",
		slog.F("field_name", "something or the other"),
		slog.F("some_map", map[string]string{
			"nested_fields": "wowow",
		}),
		slog.F("some slice", []interface{}{
			1,
			"foof",
			"bar",
			true,
		}),
		slog.Component("test"),
	)
}

func Example_test() {
	// Nil here but would be provided by the testing framework.
	var t *testing.T

	testlog.Info(t, "my message here",
		slog.F("field_name", "something or the other"),
		slog.F("some_map", map[string]interface{}{
			"nested_fields": "wowow",
		}),
		slog.F("some slice", []interface{}{
			1,
			"foof",
			"bar",
			true,
		}),
		slog.Component("test"),

		slog.F("name", slog.ValueFunc(func() interface{} {
			return "wow"
		})),
	)
}
