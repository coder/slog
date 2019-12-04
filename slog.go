// Package slog implements minimal structured logging.
//
// See https://cdr.dev/slog for more overview docs and a comparison with existing libraries.
//
// Sink implementations available in sloghuman, slogjson, slogstackdriver and slogtest.
package slog // import "cdr.dev/slog"

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"go.opencensus.io/trace"
)

// F represents a log field.
type F struct {
	Name  string
	Value interface{}
}

// Value represents a log value.
// Implement LogValue in your own types to override
// the value encoded when logging.
type Value interface {
	LogValue() interface{}
}

// Error is the standard key used for logging a Go error value.
func Error(err error) F {
	return F{
		Name:  "error",
		Value: err,
	}
}

type fieldsKey struct{}

func fieldsWithContext(ctx context.Context, fields Map) context.Context {
	return context.WithValue(ctx, fieldsKey{}, fields)
}

func fieldsFromContext(ctx context.Context) Map {
	l, _ := ctx.Value(fieldsKey{}).(Map)
	return l
}

// Context returns a context that contains the given fields.
// Any logs written with the provided context will have
// the given logs prepended.
// It will append to any fields already in ctx.
func Context(ctx context.Context, fields ...F) context.Context {
	f1 := fieldsFromContext(ctx)
	f2 := combineFields(f1, fields)
	return fieldsWithContext(ctx, f2)
}

// SinkEntry represents the structure of a log entry.
// It is the argument to the sink when logging.
type SinkEntry struct {
	Time time.Time

	Level   Level
	Message string

	LoggerName string

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
		return fmt.Sprintf(`"unknown_level: %v"`, int(l))
	}
	return s
}

// Sink is the destination of a Logger.
type Sink interface {
	LogEntry(ctx context.Context, e SinkEntry) error
	Sync() error
}

// Make creates a logger that writes logs to sink.
func Make(s Sink) Logger {
	var l Logger
	l.sinks = []sink{
		{
			sink:  s,
			level: new(int64),
		},
	}
	l.SetLevel(LevelDebug)
	return l
}

type sink struct {
	name   string
	sink   Sink
	level  *int64
	fields Map
}

func combineFields(f1, f2 Map) Map {
	f3 := make(Map, 0, len(f1)+len(f2))
	f3 = append(f3, f1...)
	f3 = append(f3, f2...)
	return f3
}
func (s sink) withFields(fields Map) sink {
	s.fields = combineFields(s.fields, fields)
	return s
}

func (s sink) named(name string) sink {
	if s.name == "" {
		s.name = name
	} else if name != "" {
		s.name += "." + name
	}
	return s
}

func (s sink) withContext(ctx context.Context) sink {
	f := fieldsFromContext(ctx)
	return s.withFields(f)
}

var helpersMu sync.Mutex
var helpers = make(map[string]struct{})

// Logger allows logging a ordered slice of fields
// to an underlying set of sinks.
type Logger struct {
	sinks []sink
	skip  int
}

func (l Logger) clone() Logger {
	l.sinks = append([]sink(nil), l.sinks...)
	return l
}

// Debug logs the msg and fields at LevelDebug.
func (l Logger) Debug(ctx context.Context, msg string, fields ...F) {
	l.log(ctx, LevelDebug, msg, fields)
}

// Info logs the msg and fields at LevelInfo.
func (l Logger) Info(ctx context.Context, msg string, fields ...F) {
	l.log(ctx, LevelInfo, msg, fields)
}

// Warn logs the msg and fields at LevelWarn.
func (l Logger) Warn(ctx context.Context, msg string, fields ...F) {
	l.log(ctx, LevelWarn, msg, fields)
}

// Error logs the msg and fields at LevelError.
func (l Logger) Error(ctx context.Context, msg string, fields ...F) {
	l.log(ctx, LevelError, msg, fields)
}

// Critical logs the msg and fields at LevelCritical.
func (l Logger) Critical(ctx context.Context, msg string, fields ...F) {
	l.log(ctx, LevelCritical, msg, fields)
}

// Fatal logs the msg and fields at LevelFatal.
func (l Logger) Fatal(ctx context.Context, msg string, fields ...F) {
	l.log(ctx, LevelFatal, msg, fields)
}

// Helper marks the calling function as a helper
// and skips it for source location information.
// It's the slog equivalent of *testing.T.Helper().
func Helper() {
	_, _, fn := location(1)
	addHelper(fn)
}

func addHelper(fn string) {
	helpersMu.Lock()
	helpers[fn] = struct{}{}
	helpersMu.Unlock()
}

// With returns a Logger that prepends the given fields on every
// logged entry.
// It will append to any fields already in the Logger.
func (l Logger) With(fields ...F) Logger {
	l = l.clone()
	for i, s := range l.sinks {
		l.sinks[i] = s.withFields(fields)
	}
	return l
}

// Named names the logger.
// If there is already a name set, it will be joined by ".".
// E.g. if the name is currently "my_component" and then later
// the name "my_pkg" is set, then the final component will be
// "my_component.my_pkg".
func (l Logger) Named(name string) Logger {
	l = l.clone()
	for i, s := range l.sinks {
		l.sinks[i] = s.named(name)
	}
	return l
}

// SetLevel changes the Logger's level.
func (l Logger) SetLevel(level Level) {
	for _, s := range l.sinks {
		atomic.StoreInt64(s.level, int64(level))
	}
}

func (l Logger) log(ctx context.Context, level Level, msg string, fields Map) {
	ent := SinkEntry{
		Time:        time.Now().UTC(),
		Level:       level,
		Message:     msg,
		Fields:      fields,
		SpanContext: trace.FromContext(ctx).SpanContext(),
	}
	helpersMu.Lock()
	ent = ent.fillLoc(helpers, l.skip+2)
	helpersMu.Unlock()

	for _, s := range l.sinks {
		slevel := Level(atomic.LoadInt64(s.level))
		if level < slevel {
			// We will not log levels below the current log level.
			continue
		}
		err := s.sink.LogEntry(ctx, s.entry(ctx, ent))
		if err != nil {
			fmt.Fprintf(os.Stderr, "slog: sink with name %v and type %T failed to log entry: %+v", s.name, s.sink, err)
			continue
		}
	}

	switch level {
	case LevelCritical, LevelError, LevelFatal:
		l.Sync()
		if level == LevelFatal {
			os.Exit(1)
		}
	}
}

// Sync calls Sync on the sinks underlying the logger.
// Used it to ensure all logs are flushed during exit.
func (l Logger) Sync() {
	for _, s := range l.sinks {
		err := s.sink.Sync()
		if err != nil {
			fmt.Fprintf(os.Stderr, "slog: sink with name %v and type %T failed to sync: %+v\n", s.name, s.sink, err)
			continue
		}
	}
}

func (ent SinkEntry) fillFromFrame(f runtime.Frame) SinkEntry {
	ent.Func = f.Function
	ent.File = f.File
	ent.Line = f.Line
	return ent
}

func (ent SinkEntry) fillLoc(helpers map[string]struct{}, skip int) SinkEntry {
	// Copied from testing.T
	const maxStackLen = 50
	var pc [maxStackLen]uintptr

	// Skip two extra frames to account for this function
	// and runtime.Callers itself.
	n := runtime.Callers(skip+2, pc[:])
	if n == 0 {
		panic("slog: zero callers found")
	}

	frames := runtime.CallersFrames(pc[:n])
	first, more := frames.Next()
	if !more {
		return ent.fillFromFrame(first)
	}

	frame := first
	for {
		if _, ok := helpers[frame.Function]; !ok {
			// Found a frame that wasn't inside a helper function.
			return ent.fillFromFrame(frame)
		}
		frame, more = frames.Next()
		if !more {
			return ent.fillFromFrame(first)
		}
	}
}

func (s sink) entry(ctx context.Context, ent SinkEntry) SinkEntry {
	s = s.withContext(ctx)
	s = s.withFields(ent.Fields)
	s = s.named(ent.LoggerName)

	ent.LoggerName = s.name
	ent.Fields = s.fields

	return ent
}

func location(skip int) (file string, line int, fn string) {
	pc, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		panic("slog: zero callers found")
	}
	f := runtime.FuncForPC(pc)
	return file, line, f.Name()
}

// Tee enables logging to multiple loggers.
func Tee(ls ...Logger) Logger {
	var l Logger
	for _, l2 := range ls {
		l.sinks = append(l.sinks, l2.sinks...)
	}
	return l
}

// JSON wraps around another type to indicate that it should be
// encoded via json.Marshal instead of via the default
// reflection based encoder.
type JSON struct {
	V interface{}
}
