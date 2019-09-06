package slog

import (
	"context"
	"go.opencensus.io/trace"
	"golang.org/x/xerrors"
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

func parseFields(fields []interface{}) parsedFields {
	var l parsedFields

	for i := 0; i < len(fields); i++ {
		f := fields[i]
		switch f := f.(type) {
		case componentField:
			l = l.appendComponent(string(f))
			continue
		case Field:
			l = l.appendField(f.LogKey(), f)
			continue
		case string:
			i++
			if i < len(fields) {
				v := fields[i]
				l = l.appendField(f, v)
				continue
			}

			// Missing value for key.
			err := xerrors.Errorf("missing log value for key %v", f)
			l = l.appendField("missing_value_error", err)
		default:
			// Unexpected key type.
			err := xerrors.Errorf("unexpected log key of type %T: %#v", f, f)
			l = l.appendField("bad_key_error", err)
			// Skip the next value under the assumption that it is a value.
			i++
		}
	}

	return l
}

func (l parsedFields) appendField(k string, v interface{}) parsedFields {
	l.fields = l.fields.append(k, reflectFieldValue(reflect.ValueOf(v)))
	return l
}

func (l parsedFields) withFields(f []interface{}) parsedFields {
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
