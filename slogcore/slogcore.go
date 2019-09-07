// slogcore contains the necessary code
// to implement a custom slog.Sink.
package slogcore // import "go.coder.com/slog/slogcore"

import (
	"time"

	"go.opencensus.io/trace"
)

// Sink is the destination of a Logger.
type Sink interface {
	WriteLogEntry(e Entry)
}

type Level string

const (
	Debug    Level = "DEBUG"
	Info     Level = "INFO"
	Warn     Level = "WARN"
	Error    Level = "ERROR"
	Critical Level = "CRITICAL"
	Fatal    Level = "FATAL"
)

type Entry struct {
	Time time.Time

	Level   Level
	Message string

	Component string

	Func string
	File string
	Line int

	SpanContext trace.SpanContext

	Fields Map
}
