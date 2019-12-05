// Package syncwriter implements a concurrency safe io.Writer wrapper.
package syncwriter

import (
	"errors"
	"io"
	"os"
	"sync"
	"syscall"
)

// Writer implements a concurrency safe io.Writer wrapper.
type Writer struct {
	mu sync.Mutex
	w  io.Writer
}

// New returns a new Writer that writes to w.
func New(w io.Writer) *Writer {
	return &Writer{
		w: w,
	}
}

// Write implements io.Writer.
func (w *Writer) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.w.Write(p)
}

type syncer interface {
	Sync() error
}

var _ syncer = &os.File{}

// Sync calls Sync on the underlying writer
// if possible.
func (w *Writer) Sync() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	s, ok := w.w.(syncer)
	if !ok {
		return nil
	}
	err := s.Sync()
	if _, ok := w.w.(*os.File); ok {
		// Opened files do not necessarily support syncing.
		// E.g. stdout and stderr both do not so we need
		// to ignore these errors.
		// See https://github.com/uber-go/zap/issues/370
		// See https://github.com/cdr/slog/pull/43
		if errors.Is(err, syscall.EINVAL) || errors.Is(err, syscall.ENOTTY) {
			return nil
		}
	}
	return err
}
