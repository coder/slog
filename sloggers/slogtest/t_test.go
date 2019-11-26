package slogtest_test

import (
	"context"
	"io"
	"testing"

	"golang.org/x/xerrors"

	"go.coder.com/slog"
	"go.coder.com/slog/sloggers/slogtest"
)

type meow struct {
	a int
}

func (m meow) LogValue() interface{} {
	return "xdxd"
}

func TestExampleTest(t *testing.T) {
	t.Parallel()

	slogtest.Info(t, "my message here",
		slog.F{"field_name", "something or the other"},
		slog.F{"some_map", slog.Map{
			{"nested_fields", "wowow"},
		}},
		slog.Error(xerrors.Errorf("wrap2: %w",
			xerrors.Errorf("wrap1: %w",
				io.EOF,
			),
		)),
		slog.F{"hi3", slog.Map{
			{"meow", meow{1}},
			{"meow", meow{2}},
			{"meow", meow{3}},
		}},
	)

	l := slogtest.Make(t, nil).With(
		slog.F{"hi", "anmol"},
	)
	stdlibLog := slog.Stdlib(context.Background(), l)
	stdlibLog.Println("stdlib")
}
