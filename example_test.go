package slog_test

import (
	"context"
	"io"
	"net"
	"os"
	"testing"
	"time"

	"go.opencensus.io/trace"
	"golang.org/x/xerrors"

	"cdr.dev/slog/v2"
	"cdr.dev/slog/v2/sloggers/sloghuman"
	"cdr.dev/slog/v2/sloggers/slogstackdriver"
	"cdr.dev/slog/v2/sloggers/slogtest"
)

func Example() {
	ctx := sloghuman.Make(context.Background(), os.Stdout)

	slog.Info(ctx, "my message here",
		slog.F("field_name", "something or the other"),
		slog.F("some_map", slog.M(
			slog.F("nested_fields", time.Date(2000, time.February, 5, 4, 4, 4, 0, time.UTC)),
		)),
		slog.Err(
			xerrors.Errorf("wrap1: %w",
				xerrors.Errorf("wrap2: %w",
					io.EOF,
				),
			),
		),
	)

	// 2019-12-09 05:04:53.398 [INFO]	<example.go:16>	my message here	{"field_name": "something or the other", "some_map": {"nested_fields": "2000-02-05T04:04:04Z"}} ...
	//  "error": wrap1:
	//      main.main
	//          /Users/nhooyr/src/cdr/scratch/example.go:22
	//    - wrap2:
	//      main.main
	//          /Users/nhooyr/src/cdr/scratch/example.go:23
	//    - EOF
}

func Example_struct() {
	ctx := sloghuman.Make(context.Background(), os.Stdout)

	type hello struct {
		Meow int       `json:"meow"`
		Bar  string    `json:"bar"`
		M    time.Time `json:"m"`
	}

	slog.Info(ctx, "check out my structure",
		slog.F("hello", hello{
			Meow: 1,
			Bar:  "barbar",
			M:    time.Date(2000, time.February, 5, 4, 4, 4, 0, time.UTC),
		}),
	)

	// 2019-12-16 17:31:51.769 [INFO]	<example_test.go:56>	check out my structure	{"hello": {"meow": 1, "bar": "barbar", "m": "2000-02-05T04:04:04Z"}}
}

func Example_testing() {
	// Provided by the testing package in tests.
	var t testing.TB

	slogtest.Info(t, "my message here",
		slog.F("field_name", "something or the other"),
	)

	// t.go:55: 2019-12-05 21:20:31.218 [INFO]	<examples_test.go:42>	my message here	{"field_name": "something or the other"}
}

func Example_tracing() {
	var ctx context.Context
	ctx = sloghuman.Make(context.Background(), os.Stdout)

	ctx, _ = trace.StartSpan(ctx, "spanName")

	slog.Info(ctx, "my msg", slog.F("hello", "hi"))

	// 2019-12-09 21:59:48.110 [INFO]	<example_test.go:62>	my msg	{"trace": "f143d018d00de835688453d8dc55c9fd", "span": "f214167bf550afc3", "hello": "hi"}
}

func Example_multiple() {
	ctx := sloghuman.Make(context.Background(), os.Stdout)

	f, err := os.OpenFile("stackdriver", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		slog.Fatal(ctx, "failed to open stackdriver log file", slog.Err(err))
	}

	ctx = slog.Make(l, slogstackdriver.Make(ctx, f))

	slog.Info(ctx, "log to stdout and stackdriver")

	// 2019-12-07 20:59:55.790 [INFO]	<example_test.go:46>	log to stdout and stackdriver
}

func ExampleWith() {
	ctx := slog.With(context.Background(), slog.F("field", 1))

	ctx = sloghuman.Make(ctx, os.Stdout)
	slog.Info(ctx, "msg")

	// 2019-12-07 20:54:23.986 [INFO]	<example_test.go:20>	msg	{"field": 1}
}

func ExampleStdlib() {
	ctx := slog.With(context.Background(), slog.F("field", 1))
	l := slog.Stdlib(sloghuman.Make(ctx, os.Stdout))

	l.Print("msg")

	// 2019-12-07 20:54:23.986 [INFO]	(stdlib)	<example_test.go:29>	msg	{"field": 1}
}

func ExampleNamed() {
	ctx := context.Background()

	ctx = sloghuman.Make(ctx, os.Stdout)
	ctx = slog.Named(ctx, "http")
	slog.Info(ctx, "received request", slog.F("remote address", net.IPv4(127, 0, 0, 1)))

	// 2019-12-07 21:20:56.974 [INFO]	(http)	<example_test.go:85>	received request	{"remote address": "127.0.0.1"}
}

func ExampleLeveled() {
	ctx := context.Background()

	ctx = sloghuman.Make(ctx, os.Stdout)
	slog.Debug(ctx, "testing1")
	slog.Info(ctx, "received request")

	ctx = slog.Leveled(ctx, slog.LevelDebug)

	slog.Debug(ctx, "testing2")

	// 2019-12-07 21:26:20.945 [INFO]	<example_test.go:95>	received request
	// 2019-12-07 21:26:20.945 [DEBUG]	<example_test.go:99>	testing2
}
