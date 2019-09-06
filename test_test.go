package slog_test

import (
	"go.coder.com/slog"
	"go.coder.com/slog/testlog"
	"golang.org/x/xerrors"
	"testing"
)

func TestTest(t *testing.T) {
	testlog.Info(t, "wow",
		slog.Error(xerrors.Errorf("wow")),
	)
}
