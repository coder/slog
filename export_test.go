package slog

func (l *Logger) SetErrorf(fn func(string, ...interface{})) {
	l.errorf = fn
}

func (l *Logger) SetExit(fn func(int)) {
	l.exit = fn
}
