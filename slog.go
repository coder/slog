package slog

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sync/atomic"
	"time"

	"go.opencensus.io/trace"

	"go.coder.com/slog/internal/skipctx"
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

type loggerKey struct{}

func withContext(ctx context.Context, l parsedFields) context.Context {
	return context.WithValue(ctx, loggerKey{}, l)
}

func fromContext(ctx context.Context) parsedFields {
	l, _ := ctx.Value(loggerKey{}).(parsedFields)
	return l
}

// Context returns a context that contains the given fields.
// Any logs written with the provided context will contain
// the given fields.
// It will append to any fields already in ctx.
func Context(ctx context.Context, fields ...Field) context.Context {
	l := fromContext(ctx)
	l = l.withFields(fields)
	return withContext(ctx, l)
}

type Entry struct {
	Time time.Time

	Level   Level
	Message string

	Component string

	Func string
	File string
	Line int

	SpanContext trace.SpanContext

	Fields []Field
}

type Level int

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
		testingHelper: func() {},
	}
	l.SetLevel(LevelDebug)

	if sink, ok := s.(interface {
		XXX_slogTestingHelper() func()
	}); ok {
		l.testingHelper = sink.XXX_slogTestingHelper()
	}
	return l
}

type sink struct {
	sink  Sink
	level *int64
	pl    parsedFields
}

type Logger struct {
	testingHelper func()

	sinks []sink
}

func (l Logger) Debug(ctx context.Context, msg string, fields ...Field) {
	l.testingHelper()
	l.log(ctx, LevelDebug, msg, fields)
}

func (l Logger) Info(ctx context.Context, msg string, fields ...Field) {
	l.testingHelper()
	l.log(ctx, LevelInfo, msg, fields)
}

func (l Logger) Warn(ctx context.Context, msg string, fields ...Field) {
	l.testingHelper()
	l.log(ctx, LevelWarn, msg, fields)
}

func (l Logger) Error(ctx context.Context, msg string, fields ...Field) {
	l.testingHelper()
	l.log(ctx, LevelError, msg, fields)
}

func (l Logger) Critical(ctx context.Context, msg string, fields ...Field) {
	l.testingHelper()
	l.log(ctx, LevelCritical, msg, fields)
}

func (l Logger) Fatal(ctx context.Context, msg string, fields ...Field) {
	l.testingHelper()
	l.log(ctx, LevelFatal, msg, fields)
}

func (l Logger) With(fields ...Field) Logger {
	sinks := make([]sink, len(l.sinks))
	for i, s := range l.sinks {
		s.pl = s.pl.withFields(fields)
		sinks[i] = s
	}
	l.sinks = sinks
	return l
}

func (l Logger) SetLevel(level Level) {
	for _, s := range l.sinks {
		atomic.StoreInt64(s.level, int64(level))
	}
}

func (l Logger) log(ctx context.Context, level Level, msg string, fields []Field) {
	l.testingHelper()

	for _, s := range l.sinks {
		slevel := Level(atomic.LoadInt64(s.level))
		if level < slevel {
			// We will not log levels below the current log level.
			continue
		}
		ent := s.pl.entry(ctx, entryParams{
			level:  level,
			msg:    msg,
			fields: fields,
			skip:   2,
		})

		s.sink.LogEntry(ctx, ent)
	}

	if level == LevelFatal {
		os.Exit(1)
	}
}

type parsedFields struct {
	component string
	spanCtx   trace.SpanContext

	fields []Field
}

func parseFields(fields []Field) parsedFields {
	var l parsedFields
	l.fields = make([]Field, 0, len(fields))

	for _, f := range fields {
		if s, ok := f.(componentField); ok {
			l = l.appendComponent(string(s))
			continue
		}
		l.fields = append(l.fields, f)
	}

	return l
}

func (l parsedFields) withFields(f []Field) parsedFields {
	return l.with(parseFields(f))
}

func (l parsedFields) with(l2 parsedFields) parsedFields {
	l = l.appendComponent(l2.component)
	if l2.spanCtx != (trace.SpanContext{}) {
		l.spanCtx = l2.spanCtx
	}

	l.fields = append(l.fields, l2.fields...)
	return l
}

func (l parsedFields) appendComponent(name string) parsedFields {
	if l.component == "" {
		l.component = name
	} else if name != "" {
		l.component += "." + name
	}
	return l
}

func (l parsedFields) withContext(ctx context.Context) parsedFields {
	l2 := fromContext(ctx)
	if len(l2.fields) == 0 {
		return l
	}

	return l.with(l2)
}

type entryParams struct {
	level  Level
	msg    string
	fields []Field
	skip   int
}

func (l parsedFields) entry(ctx context.Context, params entryParams) Entry {
	l = l.withContext(ctx)
	l = l.withFields(params.fields)

	ent := Entry{
		Time:        time.Now(),
		Level:       params.level,
		Component:   l.component,
		Message:     params.msg,
		SpanContext: trace.FromContext(ctx).SpanContext(),
		Fields:      l.fields,
	}

	file, line, fn, ok := location(params.skip + 1 + skipctx.From(ctx))
	if ok {
		ent.File = file
		ent.Line = line
		ent.Func = fn
	}
	return ent
}

func location(skip int) (file string, line int, fn string, ok bool) {
	pc, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		return "", 0, "", false
	}
	f := runtime.FuncForPC(pc)
	return file, line, f.Name(), true
}

// Tee enables logging to multiple loggers.
func Tee(ls ...Logger) Logger {
	var l Logger

	for _, l2 := range ls {
		if l2.testingHelper != nil {
			if l.testingHelper == nil {
				panic("slog.Tee: cannot Tee multiple slogtest Loggers")
			}
			l.testingHelper = l2.testingHelper
		}
		l.sinks = append(l.sinks, l2.sinks...)
	}

	return l
}
