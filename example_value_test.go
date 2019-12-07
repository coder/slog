package slog_test

import (
	"context"
	"os"

	"cdr.dev/slog"
	"cdr.dev/slog/sloggers/sloghuman"
)

type vals struct {
	first  int
	second int
}

func (s *vals) SlogValue() interface{} {
	return slog.M(
		slog.F("total", s.first+s.second),
		slog.F("first", s.first),
		slog.F("second", s.second),
	)
}

func ExampleValue() {
	l := sloghuman.Make(os.Stdout)
	l.Info(context.Background(), "hello", slog.F("val", &vals{
		first:  3,
		second: 6,
	}))

	// 2019-12-07 21:06:14.636 [INFO]	<example_value_test.go:26>	hello	{"val": {"total": 9, "first": 3, "second": 6}}
}
