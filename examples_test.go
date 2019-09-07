package slog_test

import (
	"testing"

	"go.coder.com/slog"
	"go.coder.com/slog/slogtest"
)

func Example_test() {
	// Nil here but would be provided by the testing framework.
	var t *testing.T

	slogtest.Info(t, "my message here",
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

		slog.F("name", "hi"),
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
