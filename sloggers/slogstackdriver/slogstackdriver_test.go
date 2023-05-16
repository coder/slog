package slogstackdriver_test

import (
	"bytes"
	"context"
	"fmt"
	"runtime"
	"testing"

	"go.uber.org/goleak"

	"cloud.google.com/go/compute/metadata"
	"go.opencensus.io/trace"
	logpbtype "google.golang.org/genproto/googleapis/logging/type"

	"cdr.dev/slog"
	"cdr.dev/slog/internal/assert"
	"cdr.dev/slog/internal/entryjson"
	"cdr.dev/slog/sloggers/slogstackdriver"
)

var (
	bg                               = context.Background()
	_, slogstackdriverTestFile, _, _ = runtime.Caller(0)
)

func TestStackdriver(t *testing.T) {
	t.Parallel()

	ctx, s := trace.StartSpan(bg, "meow")
	b := &bytes.Buffer{}
	l := slog.Make(slogstackdriver.Sink(b))
	l = l.Named("meow")
	l.Error(ctx, "line1\n\nline2", slog.F("wowow", "me\nyou"))

	projectID, _ := metadata.ProjectID()

	j := entryjson.Filter(b.String(), "timestampSeconds")
	j = entryjson.Filter(j, "timestampNanos")
	exp := fmt.Sprintf(`{"logging.googleapis.com/severity":"ERROR","message":"line1\n\nline2","logging.googleapis.com/sourceLocation":{"file":"%v","line":34,"function":"cdr.dev/slog/sloggers/slogstackdriver_test.TestStackdriver"},"logging.googleapis.com/operation":{"producer":"meow"},"logging.googleapis.com/trace":"projects/%v/traces/%v","logging.googleapis.com/spanId":"%v","logging.googleapis.com/trace_sampled":false,"wowow":"me\nyou"}
`, slogstackdriverTestFile, projectID, s.SpanContext().TraceID, s.SpanContext().SpanID)
	assert.Equal(t, "entry", exp, j)
}

func TestSevMapping(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "level", logpbtype.LogSeverity_DEBUG, slogstackdriver.Sev(slog.LevelDebug))
	assert.Equal(t, "level", logpbtype.LogSeverity_INFO, slogstackdriver.Sev(slog.LevelInfo))
	assert.Equal(t, "level", logpbtype.LogSeverity_WARNING, slogstackdriver.Sev(slog.LevelWarn))
	assert.Equal(t, "level", logpbtype.LogSeverity_ERROR, slogstackdriver.Sev(slog.LevelError))
	assert.Equal(t, "level", logpbtype.LogSeverity_CRITICAL, slogstackdriver.Sev(slog.LevelCritical))
}

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}
