package slog_test

import (
	"context"
	"io"
	"net"
	"os"
	"testing"

	"golang.org/x/xerrors"

	"cdr.dev/slog"
	"cdr.dev/slog/sloggers/sloghuman"
	"cdr.dev/slog/sloggers/slogstackdriver"
	"cdr.dev/slog/sloggers/slogtest"
)

func Example() {
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

func ExampleWith() {
	ctx := slog.With(context.Background(), slog.F("field", 1))

	l := sloghuman.Make(os.Stdout)
	l.Info(ctx, "msg")

	// 2019-12-07 20:54:23.986 [INFO]	<example_test.go:20>	msg	{"field": 1}
}

func ExampleStdlib() {
	ctx := slog.With(context.Background(), slog.F("field", 1))
	l := slog.Stdlib(ctx, sloghuman.Make(os.Stdout))

	l.Print("msg")

	// 2019-12-07 20:54:23.986 [INFO]	(stdlib)	<example_test.go:29>	msg	{"field": 1}
}

func ExampleTee() {
	ctx := context.Background()
	l := sloghuman.Make(os.Stdout)

	f, err := os.OpenFile("stackdriver", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		l.Fatal(ctx, "failed to open stackdriver log file", slog.Error(err))
	}

	l = slog.Tee(l, slogstackdriver.Make(f, nil))

	l.Info(ctx, "log to stdout and stackdriver")

	// 2019-12-07 20:59:55.790 [INFO]	<example_test.go:46>	log to stdout and stackdriver
}

func ExampleLogger_Named() {
	ctx := context.Background()

	l := sloghuman.Make(os.Stdout)
	l = l.Named("http")
	l.Info(ctx, "received request", slog.F("remote address", net.IPv4(127, 0, 0, 1)))

	// 2019-12-07 21:20:56.974 [INFO]	(http)	<example_test.go:85>	received request	{"remote address": "127.0.0.1"}
}

func ExampleLogger_SetLevel() {
	ctx := context.Background()

	l := sloghuman.Make(os.Stdout)
	l.Debug(ctx, "testing1")
	l.Info(ctx, "received request")

	l.SetLevel(slog.LevelDebug)

	l.Debug(ctx, "testing2")

	// 2019-12-07 21:26:20.945 [INFO]	<example_test.go:95>	received request
	// 2019-12-07 21:26:20.945 [DEBUG]	<example_test.go:99>	testing2
}
