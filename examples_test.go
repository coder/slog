package slog_test

import (
	"context"
	"go.coder.com/slog"
	"go.coder.com/slog/stderrlog"
)

func Example_stderr() {
	ctx := context.Background()
	stderrlog.Info(ctx, "my message here",
		"field_name", "something or the other",
		"some_map", map[string]string{
			"nested_fields": "wowow",
		},
		"some slice", []interface{}{
			1,
			"foof",
			"bar",
			true,
		},
		slog.Component("test"),
	)
}

// func Example_test() {
// 	// Nil here but would be provided by the testing framework.
// 	var t *testing.T
//
// 	testlog.Info(t, "my message here",
// 		"field_name", "something or the other",
// 		"some_map", map[string]interface{}{
// 			"nested_fields": "wowow",
// 		},
// 		"some slice", []interface{}{
// 			1,
// 			"foof",
// 			"bar",
// 			true,
// 		},
// 		log.Component, "test",
// 	)
// }
