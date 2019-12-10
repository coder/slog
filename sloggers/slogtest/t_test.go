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
	assert.Equal(t, 1, tb.errors, "errors")

	defer func() {
		recover()
		assert.Equal(t, 1, tb.fatals, "fatals")
	}()

	slogtest.Fatal(tb, "hello")
}

func TestIgnoreErrors(t *testing.T) {
	t.Parallel()

	tb := &fakeTB{}
	l := slog.Make(slogtest.Make(tb, &slogtest.Options{
		IgnoreErrors: true,
	}))

	l.Error(bg, "hello")
	assert.Equal(t, 0, tb.errors, "errors")

	defer func() {
		recover()
		assert.Equal(t, 0, tb.fatals, "fatals")
	}()

	l.Fatal(bg, "hello")
}

var bg = context.Background()

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
