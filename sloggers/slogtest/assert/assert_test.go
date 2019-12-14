package assert_test

import (
	"fmt"
	"io"
	"testing"

	simpleassert "cdr.dev/slog/internal/assert"
	"cdr.dev/slog/sloggers/slogtest/assert"
)

func TestEqual(t *testing.T) {
	t.Parallel()

	tb := &fakeTB{}
	assert.Equal(tb, 3, 3, "meow")

	defer func() {
		recover()
		simpleassert.Equal(t, 1, tb.fatals, "fatals")
	}()
	assert.Equal(tb, 3, 4, "meow")
}

func TestEqual_error(t *testing.T) {
	t.Parallel()

	tb := &fakeTB{}
	assert.Equal(tb, io.EOF, fmt.Errorf("failed: %w", io.EOF), "meow")

	defer func() {
		recover()
		simpleassert.Equal(t, 1, tb.fatals, "fatals")
	}()
	assert.Equal(tb, io.ErrClosedPipe, fmt.Errorf("failed: %w", io.EOF), "meow")
}

func TestSuccess(t *testing.T) {
	t.Parallel()

	tb := &fakeTB{}
	assert.Success(tb, nil, "meow")

	defer func() {
		recover()
		simpleassert.Equal(t, 1, tb.fatals, "fatals")
	}()
	assert.Success(tb, io.EOF, "meow")
}

func TestTrue(t *testing.T) {
	t.Parallel()

	tb := &fakeTB{}
	assert.True(tb, true, "meow")

	defer func() {
		recover()
		simpleassert.Equal(t, 1, tb.fatals, "fatals")
	}()
	assert.True(tb, false, "meow")
}

func TestError(t *testing.T) {
	t.Parallel()

	tb := &fakeTB{}
	assert.Error(tb, io.EOF, "meow")

	defer func() {
		recover()
		simpleassert.Equal(t, 1, tb.fatals, "fatals")
	}()
	assert.Error(tb, nil, "meow")
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
	panic(fmt.Sprint(v...))
}
