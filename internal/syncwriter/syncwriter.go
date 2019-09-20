// Package syncwriter implements a concurrency safe io.Writer wrapper.
package syncwriter

import (
	"io"
	"os"
	"sync"
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

	if f, ok := w.w.(*os.File); ok {
		// We do not want to sync if the writer is os.Stdout or os.Stderr
		// as that is unsupported on both Linux and MacOS and will return
		// an error about the fd being an invalid argument.
		// See https://github.com/uber-go/zap/issues/370
		if f == os.Stdout || f == os.Stderr {
			return nil
		}
		return f.Sync()
	}

	if s, ok := w.w.(syncer); ok {
		return s.Sync()
	}
	return nil
}
