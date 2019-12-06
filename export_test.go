package slog

import (
	"io"
)

var Exits = 0
var Errors = 0

func init() {
	exit = func(code int) {
		Exits++
	}
	ferrorf = func(io.Writer, string, ...interface{}) (int, error) {
		Errors++
		return 0, nil
	}
}
