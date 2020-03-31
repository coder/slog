package slogtest_test

import (
	"context"
	"testing"

	"cdr.dev/slog"
	"cdr.dev/slog/internal/assert"
	"cdr.dev/slog/sloggers/slogtest"
)

func TestStateless(t *testing.T) {
	t.Parallel()

	tb := &fakeTB{}
	slogtest.Debug(tb, "hello")
	slogtest.Info(tb, "hello")

	slogtest.Error(tb, "hello")
	assert.Equal(t, "errors", 1, tb.errors)

	defer func() {
		recover()
		assert.Equal(t, "fatals", 1, tb.fatals)
	}()

	slogtest.Fatal(tb, "hello")
}

func TestIgnoreErrors(t *testing.T) {
	t.Parallel()

	tb := &fakeTB{}
	ctx := context.Background()
	ctx = slog.Make(ctx, slogtest.Make(tb, &slogtest.Options{
		IgnoreErrors: true,
	}))

	slog.Error(ctx, "hello")
	assert.Equal(t, "errors", 0, tb.errors)

	defer func() {
		recover()
		assert.Equal(t, "fatals", 0, tb.fatals)
	}()

	slog.Fatal(ctx, "hello")
}

type fakeTB struct {
	testing.TB

	errors int
	fatals int
}

func (tb *fakeTB) Helper() {}

func (tb *fakeTB) Log(v ...interface{}) {}

func (tb *fakeTB) Error(v ...interface{}) {
	tb.errors++
}

func (tb *fakeTB) Fatal(v ...interface{}) {
	tb.fatals++
	panic("")
}
