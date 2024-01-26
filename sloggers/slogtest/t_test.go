package slogtest_test

import (
	"context"
	"testing"

	"golang.org/x/xerrors"

	"cdr.dev/slog"
	"cdr.dev/slog/internal/assert"
	"cdr.dev/slog/sloggers/slogtest"
)

func TestStateless(t *testing.T) {
	t.Parallel()

	tb := &fakeTB{}
	slogtest.Debug(tb, "hello")
	slogtest.Info(tb, "hello")

	slogtest.Error(tb, "canceled", slog.Error(xerrors.Errorf("test %w:", context.Canceled)))
	assert.Equal(t, "errors", 0, tb.errors)

	slogtest.Error(tb, "deadline", slog.Error(xerrors.Errorf("test %w:", context.DeadlineExceeded)))
	assert.Equal(t, "errors", 0, tb.errors)

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

func TestIgnoreErrorIs_Default(t *testing.T) {
	t.Parallel()

	tb := &fakeTB{}
	l := slogtest.Make(tb, nil)

	l.Error(bg, "canceled", slog.Error(xerrors.Errorf("test %w:", context.Canceled)))
	assert.Equal(t, "errors", 0, tb.errors)

	l.Error(bg, "deadline", slog.Error(xerrors.Errorf("test %w:", context.DeadlineExceeded)))
	assert.Equal(t, "errors", 0, tb.errors)

	l.Error(bg, "new", slog.Error(xerrors.New("test")))
	assert.Equal(t, "errors", 1, tb.errors)

	defer func() {
		recover()
		assert.Equal(t, "fatals", 1, tb.fatals)
	}()

	l.Fatal(bg, "hello", slog.Error(xerrors.Errorf("fatal %w:", context.Canceled)))
}

func TestIgnoreErrorIs_Explicit(t *testing.T) {
	t.Parallel()

	tb := &fakeTB{}
	ignored := xerrors.New("ignored")
	notIgnored := xerrors.New("not ignored")
	l := slogtest.Make(tb, &slogtest.Options{IgnoredErrorIs: []error{ignored}})

	l.Error(bg, "ignored", slog.Error(xerrors.Errorf("test %w:", ignored)))
	assert.Equal(t, "errors", 0, tb.errors)

	l.Error(bg, "not ignored", slog.Error(xerrors.Errorf("test %w:", notIgnored)))
	assert.Equal(t, "errors", 1, tb.errors)

	l.Error(bg, "canceled", slog.Error(xerrors.Errorf("test %w:", context.Canceled)))
	assert.Equal(t, "errors", 2, tb.errors)

	l.Error(bg, "deadline", slog.Error(xerrors.Errorf("test %w:", context.DeadlineExceeded)))
	assert.Equal(t, "errors", 3, tb.errors)

	l.Error(bg, "new", slog.Error(xerrors.New("test")))
	assert.Equal(t, "errors", 4, tb.errors)

	defer func() {
		recover()
		assert.Equal(t, "fatals", 1, tb.fatals)
	}()

	l.Fatal(bg, "hello", slog.Error(xerrors.Errorf("test %w:", ignored)))
}

func TestCleanup(t *testing.T) {
	t.Parallel()

	tb := &fakeTB{}
	l := slogtest.Make(tb, &slogtest.Options{})

	for _, fn := range tb.cleanups {
		fn()
	}

	// This should not log since the logger was cleaned up.
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
