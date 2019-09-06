package slog_test

import (
	"context"
	"testing"

	"go.coder.com/slog"
	"go.coder.com/slog/stderrlog"
	"go.coder.com/slog/testlog"
)

func Example_stderr() {
	ctx := context.Background()
	stderrlog.Info(ctx, "my message here",
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

	// test_test.go:17: Sep 06 14:33:34.677 [INFO] (test): my_message_here
	//	field_name: something or the other
	//	some_map:
	//	  nested_fields: wowow
	//	error:
	//	  - msg: wrap2
	//	    loc: /Users/nhooyr/src/cdr/slog/test_test.go:22
	//	    fun: go.coder.com/slog_test.TestExampleStderr
	//	  - msg: wrap1
	//	    loc: /Users/nhooyr/src/cdr/slog/test_test.go:23
	//	    fun: go.coder.com/slog_test.TestExampleStderr
	//	  - EOF
	//	name: wow
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

	// --- PASS: TestExampleTest (0.00s)
	//    test_test.go:38: Sep 06 14:33:52.628 [INFO] (test): my_message_here
	//        field_name: something or the other
	//        some_map:
	//          nested_fields: wowow
	//        error:
	//          - msg: wrap2
	//            loc: /Users/nhooyr/src/cdr/slog/test_test.go:43
	//            fun: go.coder.com/slog_test.TestExampleTest
	//          - msg: wrap1
	//            loc: /Users/nhooyr/src/cdr/slog/test_test.go:44
	//            fun: go.coder.com/slog_test.TestExampleTest
	//          - EOF
	//        name: wow
}
