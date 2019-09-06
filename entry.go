package slog

import (
	"context"
	"fmt"
	"go.coder.com/slog/internal/skipctx"
	"go.opencensus.io/trace"
	"path/filepath"
	"runtime"
	"time"
)

type entry struct {
	time time.Time

	level level
	msg   string

	component string

	fn   string
	file string
	line int

	spanCtx trace.SpanContext

	fields fieldMap
}

func (ent entry) pinnedFields() string {
	pinned := fieldMap{}

	if ent.spanCtx != (trace.SpanContext{}) {
		pinned = pinned.append("trace", fieldString(ent.spanCtx.TraceID.String()))
		pinned = pinned.append("span", fieldString(ent.spanCtx.SpanID.String()))
	}

	return marshalFields(pinned)
}

func (ent entry) stringFields() string {
	pinned := ent.pinnedFields()
	fields := marshalFields(ent.fields)

	if pinned == "" {
		return fields
	}

	if fields == "" {
		return pinned
	}

	return pinned + "\n" + fields
}

// Same as time.StampMilli but the days in the month padded by zeros.
const timestampMilli = "Jan 02 15:04:05.000"

func (ent entry) String() string {
	var ents string
	if ent.file != "" {
		ents += fmt.Sprintf("%v:%v: ", filepath.Base(ent.file), ent.line)
	}
	ents += fmt.Sprintf("%v [%v]", ent.time.Format(timestampMilli), ent.level)

	if ent.component != "" {
		ents += fmt.Sprintf(" (%v)", quote(ent.component))
	}

	ents += fmt.Sprintf(": %v", quote(ent.msg))

	fields := ent.stringFields()
	if fields != "" {
		// We never return with a trailing newline because Go's testing framework adds one
		// automatically and if we include one, then we'll get two newlines.
		// We also do not indent the fields as go's test does that automatically
		// for extra lines in a log so if we did it here, the fields would be indented
		// twice in test logs. So the Stderr logger indents all the fields itself.
		ents += "\n" + fields
	}

	return ents
}

type entryConfig struct {
	level  level
	msg    string
	fields []interface{}
	skip   int
}

func (l parsedFields) entry(ctx context.Context, config entryConfig) entry {
	l = l.withContext(ctx)
	l = l.withFields(config.fields)

	ent := entry{
		time:      time.Now(),
		level:     config.level,
		component: l.component,
		msg:       config.msg,
		spanCtx:   trace.FromContext(ctx).SpanContext(),
		fields:    l.fields,
	}

	file, line, fn, ok := location(config.skip + 1 + skipctx.From(ctx))
	if ok {
		ent.file = file
		ent.line = line
		ent.fn = fn
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
