package syncwriter

import (
	"io"
	"testing"

	"cdr.dev/slog/internal/assert"
)

func TestWriter_Sync(t *testing.T) {
	t.Parallel()

	t.Run("nonSyncWriter", func(t *testing.T) {
		w := &Writer{}
		assert.Nil(t, w.Sync(), "syncErr")
	})

	t.Run("syncWriter", func(t *testing.T) {
		w := &Writer{
			w: syncWriter{
				sw: func() error {
					return io.EOF
				},
			},
		}
		assert.Equal(t, io.EOF, w.Sync(), "syncErr")
	})
}

func Test_errorsIsAny(t *testing.T) {
	t.Parallel()

	assert.True(t, errorsIsAny(io.EOF, io.ErrUnexpectedEOF, io.EOF), "err")
	assert.False(t, errorsIsAny(io.EOF, io.ErrUnexpectedEOF, io.ErrClosedPipe), "err")
}

type writerFunc func(p []byte) (int, error)

func (f writerFunc) Write(p []byte) (int, error) {
	return f(p)
}

type syncWriter struct {
	writerFunc
	sw func() error
}

func (sw syncWriter) Sync() error {
	return sw.sw()
}
