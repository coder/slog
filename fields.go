package slog

import (
	"context"
	"runtime"
	"time"

	"go.opencensus.io/trace"

	"go.coder.com/slog/internal/skipctx"
	"go.coder.com/slog/slogcore"
)

type parsedFields struct {
	component string
	spanCtx   trace.SpanContext

	fields slogcore.Map
}

func parseFields(fields []Field) parsedFields {
	var l parsedFields
	l.fields = make(slogcore.Map, 0, len(fields))

	for _, f := range fields {
		if s, ok := f.(componentField); ok {
			l = l.appendComponent(string(s))
			continue
		}
		l = l.appendField(f.LogKey(), f.LogValue())
	}

	return l
}

func (l parsedFields) appendField(k string, v interface{}) parsedFields {
	l.fields = l.fields.Append(k, slogcore.Reflect(v))
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

	l.fields = l.fields.AppendFields(l2.fields)
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
	level  slogcore.Level
	msg    string
	fields []Field
	skip   int
}

func (l parsedFields) entry(ctx context.Context, params entryParams) slogcore.Entry {
	l = l.withContext(ctx)
	l = l.withFields(params.fields)

	ent := slogcore.Entry{
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
