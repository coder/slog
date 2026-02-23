package slogjson_test

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"runtime"
	"strings"
	"testing"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"cdr.dev/slog/v3"
	"cdr.dev/slog/v3/internal/assert"
	"cdr.dev/slog/v3/internal/entryjson"
	"cdr.dev/slog/v3/sloggers/slogjson"
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
	exp := fmt.Sprintf(`{"level":"ERROR","msg":"line1\n\nline2","caller":"%v:35","func":"cdr.dev/slog/v3/sloggers/slogjson_test.TestMake","logger_names":["named"],"trace":"%v","span":"%v","fields":{"wowow":"me\nyou"}}
`, slogjsonTestFile, span.SpanContext().TraceID().String(), span.SpanContext().SpanID().String())
	assert.Equal(t, "entry", exp, j)
}

func TestNoDriverValue(t *testing.T) {
	t.Parallel()

	b := &bytes.Buffer{}
	l := slog.Make(slogjson.Sink(b))
	l = l.Named("named")
	validField := sql.NullString{
		String: "cat",
		Valid:  true,
	}
	invalidField := sql.NullString{
		String: "dog",
		Valid:  false,
	}
	validInt := sql.NullInt64{
		Int64: 42,
		Valid: true,
	}
	l.Error(bg, "error!", slog.F("inval", invalidField), slog.F("val", validField), slog.F("int", validInt))

	j := entryjson.Filter(b.String(), "ts")
	exp := fmt.Sprintf(`{"level":"ERROR","msg":"error!","caller":"%v:61","func":"cdr.dev/slog/v3/sloggers/slogjson_test.TestNoDriverValue","logger_names":["named"],"fields":{"inval":null,"val":"cat","int":42}}
`, slogjsonTestFile)
	assert.Equal(t, "entry", exp, j)
}

func TestContextErrors(t *testing.T) {
	t.Parallel()

	t.Run("DeadlineExceeded", func(t *testing.T) {
		t.Parallel()

		b := &bytes.Buffer{}
		l := slog.Make(slogjson.Sink(b))
		l.Error(bg, "request failed", slog.Error(context.DeadlineExceeded))

		j := entryjson.Filter(b.String(), "ts")
		assert.True(t, "error contains deadline exceeded", strings.Contains(j, `"error":"context deadline exceeded"`))
	})

	t.Run("Canceled", func(t *testing.T) {
		t.Parallel()

		b := &bytes.Buffer{}
		l := slog.Make(slogjson.Sink(b))
		l.Error(bg, "request failed", slog.Error(context.Canceled))

		j := entryjson.Filter(b.String(), "ts")
		assert.True(t, "error contains context canceled", strings.Contains(j, `"error":"context canceled"`))
	})

	t.Run("WrappedDeadlineExceeded", func(t *testing.T) {
		t.Parallel()

		b := &bytes.Buffer{}
		l := slog.Make(slogjson.Sink(b))
		l.Error(bg, "request failed", slog.Error(fmt.Errorf("dial: %w", context.DeadlineExceeded)))

		j := entryjson.Filter(b.String(), "ts")
		assert.True(t, "error contains dial: context deadline exceeded", strings.Contains(j, `"error":"dial: context deadline exceeded"`))
	})
}
