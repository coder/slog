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

func (w *Writer) Sync() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if s, ok := w.w.(syncer); ok {
		return s.Sync()
	}
	return nil
}
