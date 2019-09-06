package slog

import (
	"context"
	"go.opencensus.io/trace"
	"reflect"
)

type level string

const (
	levelDebug    level = "DEBUG"
	levelInfo     level = "INFO"
	levelWarn     level = "WARN"
	levelError    level = "ERROR"
	levelCritical level = "CRITICAL"
	levelFatal    level = "FATAL"
)

type parsedFields struct {
	component string
	spanCtx   trace.SpanContext

	fields fieldMap
}

func parseFields(fields []Field) parsedFields {
	var l parsedFields
	l.fields = make(fieldMap, 0, len(fields))

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
	l.fields = l.fields.append(k, reflectFieldValue(reflect.ValueOf(v)))
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

	l.fields = l.fields.appendFields(l2.fields)
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
