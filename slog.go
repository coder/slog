package slog

import (
	"context"
	"fmt"
	"testing"
)

// field represents a log field.
type field struct {
	// name represents the name of the object.
	name string

	// Value can be a string, bool, int, float64, []interface{} or []field.
	value fieldValue
}

// TODO implement.
type Field interface {
	LogKey() string
	Value
}

type Value interface {
	LogValue() (value interface{}, err error)
}

type Logger interface {
	// Debug means a potentially noisy log.
	Debug(ctx context.Context, msg string, fields ...interface{})
	// Info means an informational log.
	Info(ctx context.Context, msg string, fields ...interface{})
	// Warn means something may be going wrong.
	Warn(ctx context.Context, msg string, fields ...interface{})
	// Error means the an error occured but does not require immediate attention.
	Error(ctx context.Context, msg string, fields ...interface{})
	// Critical means an error occured and requires immediate attention.
	Critical(ctx context.Context, msg string, fields ...interface{})
	// Fatal is the same as critical but calls os.Exit(1) afterwards.
	Fatal(ctx context.Context, msg string, fields ...interface{})

	// With returns a logger that will merge the given fields with all fields logged.
	// Fields logged with one of the above methods or from the context will always take priority.
	// Use the global with function when the fields being stored belong in the context and this
	// when they do not.
	With(fields ...interface{}) Logger
}

// Special keys in the logger's fields.
const (
	// Component represents the component a log is being logged for.
	// If there is already a component set, it will be joined by ".".
	// E.g. if the component is currently "my_component" and then later
	// the component "my_pkg" is set, then the final component will be
	// "my_component.my_pkg".
	// The component must be of type string.
	Component = "component"

	// Error is the standard key used for logging a Go error value.
	Error = "error"
)

// With returns a context that contains the given fields.
// Any logs written with the provided context will contain
// the given fields.
func With(ctx context.Context, fields ...interface{}) context.Context {
	panic("TODO")
}

func Stderr() Logger {
	panic("TODO")
}

func Test(t *testing.T) Logger {
	panic("TODO")
}

func panicf(f string, v ...interface{}) {
	f = "slog: " + f
	s := fmt.Sprintf(f, v...)
	panic(s)
}
