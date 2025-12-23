package slog_test

import (
	"context"
	"io"
	"net"
	"os"
	"testing"
	"time"

	"golang.org/x/xerrors"

	"cdr.dev/slog/v3"
	"cdr.dev/slog/v3/sloggers/sloghuman"
	"cdr.dev/slog/v3/sloggers/slogstackdriver"
	"cdr.dev/slog/v3/sloggers/slogtest"
)

func Example() {
	log := slog.Make(sloghuman.Sink(os.Stdout))

	log.Info(context.Background(), "my message here",
		slog.F("field_name", "something or the other"),
		slog.F("some_map", slog.M(
			slog.F("nested_fields", time.Date(2000, time.February, 5, 4, 4, 4, 0, time.UTC)),
		)),
		slog.Error(
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
	l := slog.Make(sloghuman.Sink(os.Stdout))

	type hello struct {
		Meow int       `json:"meow"`
		Bar  string    `json:"bar"`
		M    time.Time `json:"m"`
	}

	l.Info(context.Background(), "check out my structure",
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

	// t.go:55: 2019-12-05 21:20:31.218 [INFO]	my message here	field_name="something or the other"
}

func Example_multiple() {
	l := slog.Make(sloghuman.Sink(os.Stdout))

	f, err := os.OpenFile("stackdriver", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0o644)
	if err != nil {
		l.Fatal(context.Background(), "failed to open stackdriver log file", slog.Error(err))
	}

	l = l.AppendSinks(slogstackdriver.Sink(f))

	l.Info(context.Background(), "log to stdout and stackdriver")

	// 2019-12-07 20:59:55.790 [INFO]	log to stdout and stackdriver
}

func ExampleWith() {
	ctx := slog.With(context.Background(), slog.F("field", 1))

	l := slog.Make(sloghuman.Sink(os.Stdout))
	l.Info(ctx, "msg")

	// 2019-12-07 20:54:23.986 [INFO]	msg	field=1}
}

func ExampleStdlib() {
	ctx := slog.With(context.Background(), slog.F("field", 1))
	l := slog.Stdlib(ctx, slog.Make(sloghuman.Sink(os.Stdout)), slog.LevelInfo)

	l.Print("msg")

	// 2019-12-07 20:54:23.986 [INFO]	(stdlib)	msg	field=1
}

func ExampleLogger_Named() {
	ctx := context.Background()

	l := slog.Make(sloghuman.Sink(os.Stdout))
	l = l.Named("http")
	l.Info(ctx, "received request", slog.F("remote address", net.IPv4(127, 0, 0, 1)))

	// 2019-12-07 21:20:56.974 [INFO]	(http)	received request	remote_address=127.0.0.1}
}

func ExampleLogger_Leveled() {
	ctx := context.Background()

	l := slog.Make(sloghuman.Sink(os.Stdout))
	l.Debug(ctx, "testing1")
	l.Info(ctx, "received request")

	l = l.Leveled(slog.LevelDebug)

	l.Debug(ctx, "testing2")

	// 2019-12-07 21:26:20.945 [INFO]	received request
	// 2019-12-07 21:26:20.945 [DEBU]	testing2
}
