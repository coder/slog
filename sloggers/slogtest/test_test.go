package slogtest_test

import (
	"io"
	"strconv"
	"testing"

	"golang.org/x/xerrors"

	"go.coder.com/slog"
	"go.coder.com/slog/sloggers/slogtest"
)

type meow struct {
	a int
}

func (m meow) LogKey() string {
	return "hi" + strconv.Itoa(m.a)
}

func (m meow) LogValue() interface{} {
	return "xdxd"
}

func TestExampleTest(t *testing.T) {
	t.Parallel()

	slogtest.Info(t, "my message here",
		slog.F("field_name", "something or the other"),
		slog.F("some_map", slog.Map(
			slog.F("nested_fields", "wowow"),
		)),
		slog.Error(xerrors.Errorf("wrap2: %w",
			xerrors.Errorf("wrap1: %w",
				io.EOF,
			),
		)),
		slog.Component("test"),
		slog.F("hi3", slog.Map(
			meow{1},
			meow{2},
			meow{3},
		)),
	)
}
