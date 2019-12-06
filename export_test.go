package slog

var Exits = 0
var Errors = 0

func init() {
	exit = func(code int) {
		Exits++
	}
	errorf = func(string, ...interface{}) {
		Errors++
	}
}
