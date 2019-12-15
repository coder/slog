package syncwriter

import (
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
		assert.Equal(t, "errors", 0, tw.errors)
	})

	t.Run("syncWriter", func(t *testing.T) {
		t.Parallel()

		tw := newWriter(syncWriter{
			wf: func([]byte) (int, error) {
				return 0, io.EOF
			},
			sf: func() error {
				return io.EOF
			},
		})
		tw.w.Write("hello", nil)
		assert.Equal(t, "errors", 1, tw.errors)
		tw.w.Sync("test")
		assert.Equal(t, "errors", 2, tw.errors)
	})

	t.Run("stdout", func(t *testing.T) {
		t.Parallel()

		tw := newWriter(os.Stdout)
		tw.w.Sync("test")
		assert.Equal(t, "errors", 0, tw.errors)
	})

	t.Run("errorf", func(t *testing.T) {
		t.Parallel()

		sw := New(syncWriter{
			wf: func([]byte) (int, error) {
				return 0, io.EOF
			},
			sf: func() error {
				return io.EOF
			},
		})
		sw.Write("hello", nil)
	})
}

func Test_errorsIsAny(t *testing.T) {
	t.Parallel()

	assert.True(t, "err", errorsIsAny(io.EOF, io.ErrUnexpectedEOF, io.EOF))
	assert.False(t, "err", errorsIsAny(io.EOF, io.ErrUnexpectedEOF, io.ErrClosedPipe))
}

type syncWriter struct {
	wf func([]byte) (int, error)
	sf func() error
}

var _ syncer = &syncWriter{}

func (sw syncWriter) Write(p []byte) (int, error) {
	return sw.wf(p)
}

func (sw syncWriter) Sync() error {
	return sw.sf()
}
