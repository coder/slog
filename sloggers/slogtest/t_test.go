package slogtest_test

import (
	"context"
	"testing"

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
	l := slogtest.Make(tb, &slogtest.Options{
		IgnoreErrors: true,
	})

	l.Error(bg, "hello")
	assert.Equal(t, "errors", 0, tb.errors)

	defer func() {
		recover()
		assert.Equal(t, "fatals", 1, tb.fatals)
	}()

	l.Fatal(bg, "hello")
}

func TestCleanup(t *testing.T) {
	t.Parallel()

	tb := &fakeTB{}
	l := slogtest.Make(tb, &slogtest.Options{})

	for _, fn := range tb.cleanups {
		fn()
	}

	// This shoud not log since the logger was cleaned up.
	l.Info(bg, "hello")
	assert.Equal(t, "no logs", 0, tb.logs)
}

func TestSkipCleanup(t *testing.T) {
	t.Parallel()

	tb := &fakeTB{}
	slogtest.Make(tb, &slogtest.Options{
		SkipCleanup: true,
	})

	assert.Len(t, "no cleanups", 0, tb.cleanups)
}

var bg = context.Background()

type fakeTB struct {
	testing.TB

	logs     int
	errors   int
	fatals   int
	cleanups []func()
}

func (tb *fakeTB) Helper() {}

func (tb *fakeTB) Log(v ...interface{}) {
	tb.logs++
}

func (tb *fakeTB) Error(v ...interface{}) {
	tb.errors++
}

func (tb *fakeTB) Fatal(v ...interface{}) {
	tb.fatals++
	panic("")
}

func (tb *fakeTB) Cleanup(fn func()) {
	tb.cleanups = append(tb.cleanups, fn)
}
