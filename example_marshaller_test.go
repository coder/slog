package slog_test

import (
	"context"
	"os"

	"cdr.dev/slog"
	"cdr.dev/slog/sloggers/sloghuman"
)

type myStruct struct {
	foo int
	bar int
}

func (s myStruct) MarshalJSON() ([]byte, error) {
	return slog.M(
		slog.F("foo", s.foo),
		slog.F("bar", s.foo),
	).MarshalJSON()
}

func Example_marshaller() {
	l := slog.Make(sloghuman.Sink(os.Stdout))

	l.Info(context.Background(), "wow",
		slog.F("myStruct", myStruct{
			foo: 1,
			bar: 2,
		}),
	)

	// 2019-12-16 17:31:37.120 [INFO]	<example_marshaller_test.go:26>	wow	{"myStruct": {"foo": 1, "bar": 1}}
}
