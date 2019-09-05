package log_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"go.coder.com/slog"
	"go.coder.com/slog/cloudlog"
	"go.coder.com/slog/stderrlog"
	"go.coder.com/slog/testlog"
)

func Example_stderr() {
	ctx := context.Background()
	stderrlog.Info(ctx, "my message here", log.F{
		"field_name": "something or the other",
		"some_map": log.F{
			"nested_fields": "wowow",
		},
		"some slice": []interface{}{
			1,
			"foof",
			"bar",
			true,
		},
		log.Component: "test",
	})
}

func Example_stackdriver() {
	ctx := context.Background()
	l, closefn, err := cloudlog.Make(ctx, cloudlog.Config{
		Service: "my_service",
		Version: 1,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize logger: %v", err)
		os.Exit(1)
	}
	defer func() {
		err := closefn()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to close logger: %v", err)
		}
	}()

	l.Info(ctx, "my message here", log.F{
		"field_name": "something or the other",
		"some_map": map[string]interface{}{
			"nested_fields": "wowow",
		},
		"some slice": []interface{}{
			1,
			"foof",
			"bar",
			true,
		},
		log.Component: "test",
	})
}

func Example_test() {
	// Nil here but would be provided by the testing framework.
	var t *testing.T

	testlog.Info(t, "my message here", log.F{
		"field_name": "something or the other",
		"some_map": map[string]interface{}{
			"nested_fields": "wowow",
		},
		"some slice": []interface{}{
			1,
			"foof",
			"bar",
			true,
		},
		log.Component: "test",
	})
}
