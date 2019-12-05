// Package slogstackdriver contains the slogger for GCP.
package slogstackdriver // import "cdr.dev/slog/sloggers/slogstackdriver"

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"cloud.google.com/go/compute/metadata"
	"go.opencensus.io/trace"
	"golang.org/x/xerrors"
	logpbtype "google.golang.org/genproto/googleapis/logging/type"
	logpb "google.golang.org/genproto/googleapis/logging/v2"

	"cdr.dev/slog"
	"cdr.dev/slog/internal/syncwriter"
)

// Config for the stackdriver logger.
type Config struct {
	Labels slog.Map
}

// Make creates a slog.Logger configured to write JSON logs
// to stdout for stackdriver.
//
// See https://cloud.google.com/logging/docs/agent
func Make(config Config) slog.Logger {
	projectID, _ := metadata.ProjectID()

	return slog.Make(stackdriverSink{
		projectID: projectID,
		w:         syncwriter.New(os.Stdout),
	})
}

type stackdriverSink struct {
	projectID string
	w         *syncwriter.Writer
}

func (s stackdriverSink) LogEntry(ctx context.Context, ent slog.SinkEntry) error {
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

	if ent.LoggerName != "" {
		e = append(e, slog.F("logging.googleapis.com/operation", &logpb.LogEntryOperation{
			Producer: ent.LoggerName,
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

	buf, err := json.Marshal(e)
	if err != nil {
		return xerrors.Errorf("slogstackdriver: failed to encode entry to JSON: %w", err)
	}

	buf = append(buf, '\n')
	_, err = s.w.Write(buf)
	if err != nil {
		return xerrors.Errorf("slogstackdriver: failed to write JSON entry: %w", err)
	}
	return nil
}

func (s stackdriverSink) Sync() error {
	return s.w.Sync()
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
	case slog.LevelCritical, slog.LevelFatal:
		return logpbtype.LogSeverity_CRITICAL
	default:
		panic(fmt.Sprintf("slogstackdriver: unexpected level %v", level))
	}
}

func (s stackdriverSink) traceField(tID trace.TraceID) string {
	return fmt.Sprintf("projects/%v/traces/%v", s.projectID, tID)
}
