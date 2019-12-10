// Package slog implements minimal structured logging.
//
// See https://cdr.dev/slog for overview docs and a comparison with existing libraries.
//
// The examples are the best way to understand how to use this library effectively.
//
// This package provides a high level API around the Sink interface.
// The implementations are in the sloggers subdirectory.
package slog // import "cdr.dev/slog"

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"go.opencensus.io/trace"
)

// Sink is the destination of a Logger.
//
// All sinks must be safe for concurrent use.
type Sink interface {
	LogEntry(ctx context.Context, e SinkEntry)
	Sync()
}

// LogEntry logs the given entry with the context to the
// underlying sinks.
//
// It extends the entry with the set fields and names.
func (l Logger) LogEntry(ctx context.Context, e SinkEntry) {
	if e.Level < l.level {
		return
	}

	e.Fields = l.fields.append(e.Fields)
	e.Names = appendName(e.Names, l.names...)

	for _, s := range l.sinks {
		s.LogEntry(ctx, e)
	}
}

// Sync calls Sync on all the underlying sinks.
func (l Logger) Sync() {
	for _, s := range l.sinks {
		s.Sync()
	}
}

// Logger wraps Sink with a easy to use API.
//
// Logger is safe for concurrent use.
type Logger struct {
	names  []string
	sinks  []Sink
	skip   int
	fields Map
	level  Level

	exit func(int)
}

// Make creates a logger that writes logs to the passed sinks at LevelInfo.
func Make(sinks ...Sink) Logger {
	return Logger{
		sinks: sinks,
		level: LevelInfo,

		exit: os.Exit,
	}
}

// Debug logs the msg and fields at LevelDebug.
func (l Logger) Debug(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, LevelDebug, msg, fields)
}

// Info logs the msg and fields at LevelInfo.
func (l Logger) Info(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, LevelInfo, msg, fields)
}

// Warn logs the msg and fields at LevelWarn.
func (l Logger) Warn(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, LevelWarn, msg, fields)
}

// Error logs the msg and fields at LevelError.
//
// It will also Sync() before returning.
func (l Logger) Error(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, LevelError, msg, fields)
	l.Sync()
}

// Critical logs the msg and fields at LevelCritical.
//
// It will also Sync() before returning.
func (l Logger) Critical(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, LevelCritical, msg, fields)
	l.Sync()
}

// Fatal logs the msg and fields at LevelFatal.
//
// It will also Sync() before returning.
func (l Logger) Fatal(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, LevelFatal, msg, fields)
	l.Sync()
	l.exit(1)
}

var helpers sync.Map

// Helper marks the calling function as a helper
// and skips it for source location information.
// It's the slog equivalent of testing.TB.Helper().
func Helper() {
	_, _, fn := location(1)
	helpers.LoadOrStore(fn, struct{}{})
}

// With returns a Logger that prepends the given fields on every
// logged entry.
// It will append to any fields already in the Logger.
func (l Logger) With(fields ...Field) Logger {
	l.fields = l.fields.append(fields)
	return l
}

// Named names the logger.
// If there is already a name set, it will be joined by ".".
// E.g. if the name is currently "my_component" and then later
// the name "my_pkg" is set, then the final component will be
// "my_component.my_pkg".
func (l Logger) Named(name string) Logger {
	l.names = appendName(l.names, name)
	return l
}

// Leveled returns a Logger that only logs entries
// equal to or above the given level.
func (l Logger) Leveled(level Level) Logger {
	l.level = level
	return l
}

func (ent SinkEntry) fillFromFrame(f runtime.Frame) SinkEntry {
	ent.Func = f.Function
	ent.File = f.File
	ent.Line = f.Line
	return ent
}

func (ent SinkEntry) fillLoc(skip int) SinkEntry {
	// Copied from testing.T
	const maxStackLen = 50
	var pc [maxStackLen]uintptr

	// Skip two extra frames to account for this function
	// and runtime.Callers itself.
	n := runtime.Callers(skip+2, pc[:])
	frames := runtime.CallersFrames(pc[:n])
	for {
		frame, more := frames.Next()
		_, helper := helpers.Load(frame.Function)
		if !helper || !more {
			// Found a frame that wasn't a helper function.
			// Or we ran out of frames to check.
			return ent.fillFromFrame(frame)
		}
	}
}

func location(skip int) (file string, line int, fn string) {
	pc, file, line, _ := runtime.Caller(skip + 1)
	f := runtime.FuncForPC(pc)
	return file, line, f.Name()
}

func appendName(names []string, names2 ...string) []string {
	names = append([]string(nil), names...)
	names = append(names, names2...)
	return names
}

func (l Logger) log(ctx context.Context, level Level, msg string, fields Map) {
	ent := l.entry(ctx, level, msg, fields)
	l.LogEntry(ctx, ent)
}

func (l Logger) entry(ctx context.Context, level Level, msg string, fields Map) SinkEntry {
	ent := SinkEntry{
		Time:        time.Now().UTC(),
		Level:       level,
		Message:     msg,
		Fields:      fieldsFromContext(ctx).append(fields),
		SpanContext: trace.FromContext(ctx).SpanContext(),
	}
	ent = ent.fillLoc(l.skip + 3)
	return ent
}

// Field represents a log field.
type Field struct {
	Name  string
	Value interface{}
}

// F is a convenience constructor for Field.
func F(name string, value interface{}) Field {
	return Field{Name: name, Value: value}
}

// M is a convenience constructor for Map
func M(fs ...Field) Map {
	return fs
}

// Value represents a log value.
// Implement SlogValue in your own types to override
// the value encoded when logging.
type Value interface {
	SlogValue() interface{}
}

// Error is the standard key used for logging a Go error value.
func Error(err error) Field {
	return F("error", err)
}

type fieldsKey struct{}

func fieldsWithContext(ctx context.Context, fields Map) context.Context {
	return context.WithValue(ctx, fieldsKey{}, fields)
}

func fieldsFromContext(ctx context.Context) Map {
	l, _ := ctx.Value(fieldsKey{}).(Map)
	return l
}

// With returns a context that contains the given fields.
// Any logs written with the provided context will have
// the given logs prepended.
// It will append to any fields already in ctx.
func With(ctx context.Context, fields ...Field) context.Context {
	f1 := fieldsFromContext(ctx)
	f2 := f1.append(fields)
	return fieldsWithContext(ctx, f2)
}

// SinkEntry represents the structure of a log entry.
// It is the argument to the sink when logging.
type SinkEntry struct {
	Time time.Time

	Level   Level
	Message string

	// Names represents the chain of names on the
	// logger constructed with Named.
	Names []string

	Func string
	File string
	Line int

	SpanContext trace.SpanContext

	Fields Map
}

// Level represents a log level.
type Level int

// The supported log levels.
const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelCritical
	LevelFatal
)

var levelStrings = map[Level]string{
	LevelDebug:    "DEBUG",
	LevelInfo:     "INFO",
	LevelWarn:     "WARN",
	LevelError:    "ERROR",
	LevelCritical: "CRITICAL",
	LevelFatal:    "FATAL",
}

func (l Level) String() string {
	s, ok := levelStrings[l]
	if !ok {
		return fmt.Sprintf("slog.Level(%v)", int(l))
	}
	return s
}

func (f1 Map) append(f2 Map) Map {
	f3 := make(Map, 0, len(f1)+len(f2))
	f3 = append(f3, f1...)
	f3 = append(f3, f2...)
	return f3
}
