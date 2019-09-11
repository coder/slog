package slog

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

// ValueFunc is a convenient function wrapper around Value.
type ValueFunc func() interface{}

// LogValue implements Value.
func (v ValueFunc) LogValue() interface{} {
	return v()
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

// Map is used to create an ordered map of fields that can be
// logged.
func Map(fs ...Field) []Field {
	return fs
}

// Error is the standard key used for logging a Go error value.
func Error(err error) Field {
	return unparsedField{
		name: "error",
		v:    err,
	}
}

type fieldsKey struct{}

func fieldsWithContext(ctx context.Context, fields []Field) context.Context {
	return context.WithValue(ctx, fieldsKey{}, fields)
}

func fieldsFromContext(ctx context.Context) []Field {
	l, _ := ctx.Value(fieldsKey{}).([]Field)
	return l
}

// Context returns a context that contains the given fields.
// Any logs written with the provided context will have
// the given logs prepended.
// It will append to any fields already in ctx.
func Context(ctx context.Context, fields ...Field) context.Context {
	f1 := fieldsFromContext(ctx)
	f2 := combineFields(f1, fields)
	return fieldsWithContext(ctx, f2)
}

// Entry represents the structure of a log entry.
// It is the argument to the sink when logging.
type Entry struct {
	Time time.Time

	Level   Level
	Message string

	LoggerName string

	Func string
	File string
	Line int

	SpanContext trace.SpanContext

	Fields []Field
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
	LogEntry(ctx context.Context, e Entry)
}

// Make creates a logger that writes logs to sink.
func Make(s Sink) Logger {
	l := Logger{
		sinks: []sink{
			{
				sink:  s,
				level: new(int64),
			},
		},
		skip: 2,
	}
	l.SetLevel(LevelDebug)
	return l
}

type sink struct {
	name   string
	sink   Sink
	level  *int64
	fields []Field
}

func combineFields(f1, f2 []Field) []Field {
	f3 := make([]Field, 0, len(f1)+len(f2))
	f3 = append(f3, f1...)
	f3 = append(f3, f2...)
	return f3
}
func (s sink) withFields(fields []Field) sink {
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
func (l Logger) Error(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, LevelError, msg, fields)
}

// Critical logs the msg and fields at LevelCritical.
func (l Logger) Critical(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, LevelCritical, msg, fields)
}

// Fatal logs the msg and fields at LevelFatal.
func (l Logger) Fatal(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, LevelFatal, msg, fields)
}

var helpersMu = &sync.Mutex{}
var logHelpers = map[string]struct{}{}

// Helper marks the calling function as a helper
// and skips it for source location information.
func (l Logger) Helper() {
	helpersMu.Lock()
	defer helpersMu.Unlock()
	_, _, fn := location(1)
	logHelpers[fn] = struct{}{}
}

// With returns a Logger that prepends the given fields on every
// logged entry.
// It will append to any fields already in the Logger.
func (l Logger) With(fields ...Field) Logger {
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

func (l Logger) log(ctx context.Context, level Level, msg string, fields []Field) {
	params := entryParams{
		time:        time.Now(),
		level:       level,
		msg:         msg,
		fields:      fields,
		spanContext: trace.FromContext(ctx).SpanContext(),
	}
	params = params.fillLoc(l.skip + 1)

	for _, s := range l.sinks {
		slevel := Level(atomic.LoadInt64(s.level))
		if level < slevel {
			// We will not log levels below the current log level.
			continue
		}
		ent := s.entry(ctx, params)

		s.sink.LogEntry(ctx, ent)
	}

	if level == LevelFatal {
		os.Exit(1)
	}
}

type entryParams struct {
	time   time.Time
	level  Level
	msg    string
	fields []Field

	file string
	line int
	fn   string

	spanContext trace.SpanContext
}

func (p entryParams) fillFromFrame(f runtime.Frame) entryParams {
	p.fn = f.Function
	p.file = f.File
	p.line = f.Line
	return p
}

func (p entryParams) fillLoc(skip int) entryParams {
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
		return p.fillFromFrame(first)
	}
	for {
		frame, more := frames.Next()
		if !more {
			return p.fillFromFrame(first)
		}
		if _, ok := logHelpers[frame.Function]; !ok {
			// Found a frame that wasn't inside a helper function.
			return p.fillFromFrame(frame)
		}
	}
}

func (s sink) entry(ctx context.Context, params entryParams) Entry {
	s = s.withContext(ctx)
	s = s.withFields(params.fields)

	ent := Entry{
		Time:        params.time,
		Level:       params.level,
		LoggerName:  s.name,
		Message:     params.msg,
		SpanContext: params.spanContext,
		Fields:      s.fields,

		File: params.file,
		Line: params.line,
		Func: params.fn,
	}

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

type jsonValue struct {
	v interface{}
}

func (v jsonValue) LogValueJSON() interface{} {
	return v.v
}

func (v jsonValue) LogValue() interface{} {
	return v.v
}

// JSON tells the sink that it is valid
// to log the value as JSON.
// TODO this is jank.
func JSON(v interface{}) Value {
	return jsonValue{v: v}
}
