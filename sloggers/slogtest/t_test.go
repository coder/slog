package slogtest_test

import (
	"context"
	"testing"

	"cdr.dev/slog/v2"
	"cdr.dev/slog/v2/internal/assert"
	"cdr.dev/slog/v2/sloggers/slogtest"
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
	ctx = slog.Tee(ctx, slogtest.Make(tb, &slogtest.Options{
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
