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
	assert.Equal(tb, "meow", 3, 3)

	defer func() {
		recover()
		simpleassert.Equal(t, "fatals", 1, tb.fatals)
	}()
	assert.Equal(tb, "meow", 3, 4)
}

func TestEqual_error(t *testing.T) {
	t.Parallel()

	tb := &fakeTB{}
	assert.Equal(tb, "meow", io.EOF, fmt.Errorf("failed: %w", io.EOF))

	defer func() {
		recover()
		simpleassert.Equal(t, "fatals", 1, tb.fatals)
	}()
	assert.Equal(tb, "meow", io.ErrClosedPipe, fmt.Errorf("failed: %w", io.EOF))
}

func TestErrorContains(t *testing.T) {
	t.Parallel()

	tb := &fakeTB{}
	assert.ErrorContains(tb, "meow", io.EOF, "eof")

	defer func() {
		recover()
		simpleassert.Equal(t, "fatals", 1, tb.fatals)

	}()
	assert.ErrorContains(tb, "meow", io.ErrClosedPipe, "eof")

}
func TestSuccess(t *testing.T) {
	t.Parallel()

	tb := &fakeTB{}
	assert.Success(tb, "meow", nil)

	defer func() {
		recover()
		simpleassert.Equal(t, "fatals", 1, tb.fatals)
	}()
	assert.Success(tb, "meow", io.EOF)
}

func TestTrue(t *testing.T) {
	t.Parallel()

	tb := &fakeTB{}
	assert.True(tb, "meow", true)

	defer func() {
		recover()
		simpleassert.Equal(t, "fatals", 1, tb.fatals)
	}()
	assert.True(tb, "meow", false)
}

func TestFalse(t *testing.T) {
	t.Parallel()

	tb := &fakeTB{}
	assert.False(tb, "woof", false)

	defer func() {
		recover()
		simpleassert.Equal(t, "fatals", 1, tb.fatals)
	}()
	assert.False(tb, "woof", true)
}

func TestError(t *testing.T) {
	t.Parallel()

	tb := &fakeTB{}
	assert.Error(tb, "meow", io.EOF)

	defer func() {
		recover()
		simpleassert.Equal(t, "fatals", 1, tb.fatals)
	}()
	assert.Error(tb, "meow", nil)
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
