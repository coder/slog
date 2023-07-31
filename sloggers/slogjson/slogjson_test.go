package slogjson_test

import (
	"bytes"
	"context"
	"fmt"
	"runtime"
	"testing"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"cdr.dev/slog"
	"cdr.dev/slog/internal/assert"
	"cdr.dev/slog/internal/entryjson"
	"cdr.dev/slog/sloggers/slogjson"
)

var _, slogjsonTestFile, _, _ = runtime.Caller(0)

var bg = context.Background()

func TestMake(t *testing.T) {
	t.Parallel()

	tp := sdktrace.NewTracerProvider()
	tracer := tp.Tracer("tracer")
	ctx, span := tracer.Start(bg, "trace")
	span.End()
	_ = tp.Shutdown(bg)
	b := &bytes.Buffer{}
	l := slog.Make(slogjson.Sink(b))
	l = l.Named("named")
	l.Error(ctx, "line1\n\nline2", slog.F("wowow", "me\nyou"))

	j := entryjson.Filter(b.String(), "ts")
	exp := fmt.Sprintf(`{"level":"ERROR","msg":"line1\n\nline2","caller":"%v:33","func":"cdr.dev/slog/sloggers/slogjson_test.TestMake","logger_names":["named"],"trace":"%v","span":"%v","fields":{"wowow":"me\nyou"}}
`, slogjsonTestFile, span.SpanContext().TraceID().String(), span.SpanContext().SpanID().String())
	assert.Equal(t, "entry", exp, j)
}
