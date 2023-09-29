package slogstackdriver_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"runtime"
	"testing"
	"time"

	"go.uber.org/goleak"

	"cloud.google.com/go/compute/metadata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
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

	tp := sdktrace.NewTracerProvider()
	tracer := tp.Tracer("tracer")
	ctx, span := tracer.Start(bg, "trace")
	span.End()
	_ = tp.Shutdown(bg)
	b := &bytes.Buffer{}
	l := slog.Make(slogstackdriver.Sink(b))
	l = l.Named("meow")
	l.Error(ctx, "line1\n\nline2", slog.F("wowow", "me\nyou"))

	projectID, _ := metadataClient(t).ProjectID()

	j := entryjson.Filter(b.String(), "timestampSeconds")
	j = entryjson.Filter(j, "timestampNanos")
	exp := fmt.Sprintf(`{"logging.googleapis.com/severity":"ERROR","severity":"ERROR","message":"line1\n\nline2","logging.googleapis.com/sourceLocation":{"file":"%v","line":40,"function":"cdr.dev/slog/sloggers/slogstackdriver_test.TestStackdriver"},"logging.googleapis.com/operation":{"producer":"meow"},"logging.googleapis.com/trace":"projects/%v/traces/%v","logging.googleapis.com/spanId":"%v","logging.googleapis.com/trace_sampled":%v,"wowow":"me\nyou"}
`, slogstackdriverTestFile, projectID, span.SpanContext().TraceID(), span.SpanContext().SpanID(), span.SpanContext().IsSampled())
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

func metadataClient(t testing.TB) *metadata.Client {
	// When not running in Google Cloud, the default metadata client will
	// leak a goroutine.
	//
	// We use a very short timeout because the metadata server should be
	// within the same datacenter as the cloud instance.
	tp := http.DefaultTransport.(*http.Transport).Clone()
	httpClient := &http.Client{
		Timeout:   time.Second * 3,
		Transport: tp,
	}
	client := metadata.NewClient(httpClient)
	t.Cleanup(httpClient.CloseIdleConnections)
	return client
}
