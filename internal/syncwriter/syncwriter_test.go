package syncwriter

import (
	"bytes"
	"io"
	"os"
	"testing"

	"cdr.dev/slog/internal/assert"
)

type testWriter struct {
	w      *Writer
	errors int
}

func TestWriter_Sync(t *testing.T) {
	t.Parallel()

	newWriter := func(w io.Writer) *testWriter {
		tw := &testWriter{
			w: New(w),
		}
		tw.w.errorf = func(f string, v ...interface{}) {
			tw.errors++
		}
		return tw
	}

	t.Run("nonSyncWriter", func(t *testing.T) {
		t.Parallel()

		tw := newWriter(nil)
		tw.w.Sync("test")
		assert.Equal(t, 0, tw.errors, "errors")
	})

	t.Run("syncWriter", func(t *testing.T) {
		t.Parallel()

		tw := newWriter(syncWriter{
			sw: func() error {
				return io.EOF
			},
		})
		tw.w.Sync("test")
		assert.Equal(t, 1, tw.errors, "errors")
	})

	t.Run("stdout", func(t *testing.T) {
		t.Parallel()

		tw := newWriter(os.Stdout)
		tw.w.Sync("test")
		assert.Equal(t, 0, tw.errors, "errors")
	})
}

func Test_errorsIsAny(t *testing.T) {
	t.Parallel()

	assert.True(t, errorsIsAny(io.EOF, io.ErrUnexpectedEOF, io.EOF), "err")
	assert.False(t, errorsIsAny(io.EOF, io.ErrUnexpectedEOF, io.ErrClosedPipe), "err")
}

type syncWriter struct {
	*bytes.Buffer
	sw func() error
}

var _ syncer = &syncWriter{}

func (sw syncWriter) Sync() error {
	return sw.sw()
}
