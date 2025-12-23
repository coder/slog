package slogtest_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"golang.org/x/xerrors"

	"cdr.dev/slog/v3"
	"cdr.dev/slog/v3/internal/assert"
	"cdr.dev/slog/v3/sloggers/slogtest"
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

func TestIgnoreErrorFn(t *testing.T) {
	t.Parallel()

	tb := &fakeTB{}
	ignored := testCodedError{code: 777}
	notIgnored := testCodedError{code: 911}
	l := slogtest.Make(tb, &slogtest.Options{IgnoreErrorFn: func(ent slog.SinkEntry) bool {
		err, ok := slogtest.FindFirstError(ent)
		if !ok {
			t.Error("did not contain an error")
			return false
		}
		ce := testCodedError{}
		if !xerrors.As(err, &ce) {
			return false
		}
		return ce.code != 911
	}})

	l.Error(bg, "ignored", slog.Error(xerrors.Errorf("test %w:", ignored)))
	assert.Equal(t, "errors", 0, tb.errors)

	l.Error(bg, "not ignored", slog.Error(xerrors.Errorf("test %w:", notIgnored)))
	assert.Equal(t, "errors", 1, tb.errors)

	// still ignored by default for IgnoredErrorIs
	l.Error(bg, "canceled", slog.Error(xerrors.Errorf("test %w:", context.Canceled)))
	assert.Equal(t, "errors", 1, tb.errors)

	l.Error(bg, "new", slog.Error(xerrors.New("test")))
	assert.Equal(t, "errors", 2, tb.errors)

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

func TestUnmarshalable(t *testing.T) {
	t.Parallel()
	tb := &fakeTB{}
	l := slogtest.Make(tb, &slogtest.Options{})
	s := &selfRef{}
	s.Ref = s
	s2 := selfRef{Ref: s} // unmarshalable because it contains a cyclic ref
	l.Info(bg, "hello", slog.F("self", s2))
	assert.Equal(t, "errors", 1, tb.errors)
	assert.Len(t, "len errorfs", 1, tb.errorfs)
	assert.True(t, "errorfs", strings.Contains(tb.errorfs[0], "failed to log field \"self\":"))
}

type selfRef struct {
	Ref *selfRef
}

var bg = context.Background()

type fakeTB struct {
	testing.TB

	logs     int
	errors   int
	errorfs  []string
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

func (tb *fakeTB) Errorf(msg string, v ...interface{}) {
	tb.errors++
	tb.errorfs = append(tb.errorfs, fmt.Sprintf(msg, v...))
}

func (tb *fakeTB) Fatal(v ...interface{}) {
	tb.fatals++
	panic("")
}

func (tb *fakeTB) Cleanup(fn func()) {
	tb.cleanups = append(tb.cleanups, fn)
}

type testCodedError struct {
	code int
}

func (e testCodedError) Error() string {
	return fmt.Sprintf("code: %d", e.code)
}
