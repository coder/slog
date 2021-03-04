// Package slogstackdriver contains the slogger for google cloud's stackdriver.
package slogstackdriver // import "cdr.dev/slog/sloggers/slogstackdriver"

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"cloud.google.com/go/compute/metadata"
	"go.opencensus.io/trace"
	logpbtype "google.golang.org/genproto/googleapis/logging/type"
	logpb "google.golang.org/genproto/googleapis/logging/v2"

	"cdr.dev/slog"
	"cdr.dev/slog/internal/syncwriter"
)

// Sink creates a slog.Sink configured to write JSON logs
// to stdout for stackdriver.
//
// See https://cloud.google.com/logging/docs/agent
func Sink(w io.Writer) slog.Sink {
	projectID, _ := metadata.ProjectID()

	return stackdriverSink{
		projectID: projectID,
		w:         syncwriter.New(w),
	}
}

type stackdriverSink struct {
	projectID string
	w         *syncwriter.Writer
}

func (s stackdriverSink) LogEntry(ctx context.Context, ent slog.SinkEntry) {
	// https://cloud.google.com/logging/docs/agent/configuration#special-fields
	e := slog.M(
		slog.F("severity", sev(ent.Level)),
		slog.F("message", ent.Message),
		slog.F("timestamp", ent.Time),
		slog.F("logging.googleapis.com/sourceLocation", &logpb.LogEntrySourceLocation{
			File:     ent.File,
			Line:     int64(ent.Line),
			Function: ent.Func,
		}),
	)

	if len(ent.LoggerNames) > 0 {
		e = append(e, slog.F("logging.googleapis.com/operation", &logpb.LogEntryOperation{
			Producer: strings.Join(ent.LoggerNames, "."),
		}))
	}

	if ent.SpanContext != (trace.SpanContext{}) {
		e = append(e,
			slog.F("logging.googleapis.com/trace", s.traceField(ent.SpanContext.TraceID)),
			slog.F("logging.googleapis.com/spanId", ent.SpanContext.SpanID.String()),
			slog.F("logging.googleapis.com/trace_sampled", ent.SpanContext.IsSampled()),
		)
	}

	e = append(e, ent.Fields...)

	buf, _ := json.Marshal(e)

	buf = append(buf, '\n')
	s.w.Write("slogstackdriver", buf)
}

func (s stackdriverSink) Sync() {
	s.w.Sync("stackdriverSink")
}

func sev(level slog.Level) logpbtype.LogSeverity {
	switch level {
	case slog.LevelDebug:
		return logpbtype.LogSeverity_DEBUG
	case slog.LevelInfo:
		return logpbtype.LogSeverity_INFO
	case slog.LevelWarn:
		return logpbtype.LogSeverity_WARNING
	case slog.LevelError:
		return logpbtype.LogSeverity_ERROR
	default:
		return logpbtype.LogSeverity_CRITICAL
	}
}

func (s stackdriverSink) traceField(tID trace.TraceID) string {
	return fmt.Sprintf("projects/%v/traces/%v", s.projectID, tID)
}
