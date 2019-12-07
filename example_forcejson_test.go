package slog_test

import (
	"context"
	"fmt"
	"os"

	"cdr.dev/slog"
	"cdr.dev/slog/sloggers/sloghuman"
)

type stringer struct {
	X int `json:"x"`
}

func (s *stringer) String() string {
	return fmt.Sprintf("string method: %v", s.X)
}

func (s *stringer) SlogValue() interface{} {
	return slog.ForceJSON(s)
}

func ExampleForceJSON() {
	l := sloghuman.Make(os.Stdout)

	l.Info(context.Background(), "hello", slog.F("stringer", &stringer{X: 3}))

	// 2019-12-06 23:33:40.263 [INFO]	<example_force_json_test.go:27>	hello	{"stringer": {"x": 3}}
}
