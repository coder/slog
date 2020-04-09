package slog

func (l *Logger) SetExit(fn func(int)) {
	l.exit = fn
}
