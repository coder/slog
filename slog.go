package slog

import (
	"context"
)

// Logger is the core interface for logging.
type Logger interface {
	// Debug means a potentially noisy log.
	Debug(ctx context.Context, msg string, fields ...Field)
	// Info means an informational log.
	Info(ctx context.Context, msg string, fields ...Field)
	// Warn means something may be going wrong.
	Warn(ctx context.Context, msg string, fields ...Field)
	// Error means the an error occured but does not require immediate attention.
	Error(ctx context.Context, msg string, fields ...Field)
	// Critical means an error occured and requires immediate attention.
	Critical(ctx context.Context, msg string, fields ...Field)
	// Fatal is the same as critical but calls os.Exit(1) afterwards.
	Fatal(ctx context.Context, msg string, fields ...Field)

	// With returns a logger that will merge the given fields with all fields logged.
	// Fields logged with one of the above methods or from the context will always take priority.
	// Use the global with function when the fields being stored belong in the context and this
	// when they do not.
	With(fields ...Field) Logger
}

// field represents a log field.
type field struct {
	name  string
	value fieldValue
}

// Field represents a log field.
type Field interface {
	LogKey() string
	Value
}

// Value represents a log value.
// The value returned will be logged.
// Your own types can implement this interface to
// override their logging appearance.
type Value interface {
	LogValue() interface{}
}

// ValueFunc is a function that computes its logging
// representation. Use it to override the logging
// representation of a structure inline.
type ValueFunc func() interface{}

// LogValue implements Value.
func (fn ValueFunc) LogValue() interface{} {
	return fn()
}

type componentField string

func (c componentField) LogKey() string        { panic("never called") }
func (c componentField) LogValue() interface{} { panic("never called") }

// Component represents the component a log is being logged for.
// If there is already a component set, it will be joined by ".".
// E.g. if the component is currently "my_component" and then later
// the component "my_pkg" is set, then the final component will be
// "my_component.my_pkg".
func Component(name string) Field {
	return componentField(name)
}

type unparsedField struct {
	name string
	v    interface{}
}

func (f unparsedField) LogKey() string {
	return f.name
}

func (f unparsedField) LogValue() interface{} {
	return f.v
}

// F is used to log arbitrary fields with the logger.
func F(name string, v interface{}) Field {
	return unparsedField{
		name: name,
		v:    v,
	}
}

// Error is the standard key used for logging a Go error value.
func Error(err error) Field {
	return unparsedField{
		name: "error",
		v:    err,
	}
}

type loggerKey struct{}

func withContext(ctx context.Context, l parsedFields) context.Context {
	return context.WithValue(ctx, loggerKey{}, l)
}

func fromContext(ctx context.Context) parsedFields {
	l, _ := ctx.Value(loggerKey{}).(parsedFields)
	return l
}

// With returns a context that contains the given fields.
// Any logs written with the provided context will contain
// the given fields.
func With(ctx context.Context, fields ...Field) context.Context {
	l := fromContext(ctx)
	l = l.withFields(fields)
	return withContext(ctx, l)
}
