package slogjson // import "go.coder.com/slog/sloggers/slogjson"

import (
	"io"

	"go.coder.com/slog"
)

// Make creates a logger that writes JSON logs
// to the given writer. The format is as follows:
func Make(w io.Writer, opts *slog.Options) slog.Logger {
	panic("TODO")
}
