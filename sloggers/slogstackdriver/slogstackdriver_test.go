package slogstackdriver_test

import (
	"context"
	"testing"

	"cdr.dev/slog/sloggers/slogstackdriver"
)

func TestStackdriver(t *testing.T) {
	l := slogstackdriver.Make(slogstackdriver.Config{})
	l.Info(context.Background(), "meow")
	l.Sync()
}
