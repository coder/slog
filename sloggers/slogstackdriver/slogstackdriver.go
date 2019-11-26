// Package slogstackdriver contains the slogger for GCP.
package slogstackdriver // import "go.coder.com/slog/sloggers/slogstackdriver"

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"runtime/debug"
	"strconv"

	"cloud.google.com/go/compute/metadata"
	"cloud.google.com/go/errorreporting"
	"cloud.google.com/go/logging"
	"go.opencensus.io/trace"
	"golang.org/x/xerrors"
	logpb "google.golang.org/genproto/googleapis/logging/v2"

	"go.coder.com/slog"
)

// Config for the stackdriver logger.
type Config struct {
	// Name of the service that logs will be written for.
	Service string

	// Version of the service.
	Version int
}

// Make creates a slog.Logger configured to write logs to
// stackdriver.
func Make(config Config) (slog.Logger, error) {
	ctx := context.Background()

	projectID, err := metadata.ProjectID()
	if err != nil {
		return slog.Logger{}, xerrors.Errorf("failed to get project ID: %w", err)
	}
	instanceName, err := metadata.InstanceName()
	if err != nil {
		return slog.Logger{}, xerrors.Errorf("failed to get instance name: %w", err)
	}

	cl, err := logging.NewClient(ctx, "projects/"+projectID)
	if err != nil {
		return slog.Logger{}, xerrors.Errorf("failed to create new logging client: %w", err)
	}

	cl.OnError = onError

	logID := fmt.Sprintf("%v-v%v", config.Service, config.Version)

	ec, err := errorreporting.NewClient(ctx, projectID, errorreporting.Config{
		ServiceName:    config.Service,
		ServiceVersion: strconv.Itoa(config.Version),
		OnError:        errorreportingOnError,
	})
	if err != nil {
		return slog.Logger{}, xerrors.Errorf("failed to create new error reporting client: %v", err)
	}

	l := cl.Logger(logID, logging.CommonLabels(map[string]string{
		"instance_name": instanceName,
		"internal_ip":   internalIP(),
	}))

	return slog.Make(stackdriverSink{
		l:  l,
		ec: ec,
	}), nil
}

type stackdriverSink struct {
	projectID string
	l         *logging.Logger
	ec        *errorreporting.Client
}

func (s stackdriverSink) LogEntry(ctx context.Context, ent slog.SinkEntry) error {
	if ent.Message != "" {
		ent.Fields = append(slog.Map{}, ent.Fields...)
		// https://cloud.google.com/logging/docs/view/overview#expanding
		ent.Fields = append(ent.Fields, slog.F{"message", ent.Message})
	}
	e := logging.Entry{
		Timestamp: ent.Time,
		Severity:  sev(ent.Level),
		Payload:   ent.Fields,
		SourceLocation: &logpb.LogEntrySourceLocation{
			File:     ent.File,
			Line:     int64(ent.Line),
			Function: ent.Func,
		},
	}

	if ent.LoggerName != "" {
		e.Operation = &logpb.LogEntryOperation{
			Producer: ent.LoggerName,
		}
	}

	if ent.SpanContext != (trace.SpanContext{}) {
		e.Trace = s.traceField(ent.SpanContext.TraceID)
		e.SpanID = ent.SpanContext.SpanID.String()
		e.TraceSampled = ent.SpanContext.IsSampled()
	}

	if ent.Level >= slog.LevelError {
		s.ec.Report(errorReportingEntry(ent))
	}

	s.l.Log(e)
	return nil
}

func errorReportingEntry(ent slog.SinkEntry) errorreporting.Entry {
	errEnt := errorreporting.Entry{
		Error: xerrors.New(ent.Message),
		Stack: debug.Stack(),
	}

	lines := bytes.Split(errEnt.Stack, []byte("\n"))

	goroutine := lines[:1]
	// 1 skips the goroutine N line, then we skip the next 4 frames.
	start := lines[1+2*5:]

	lines = append(goroutine, start...)

	errEnt.Stack = bytes.Join(lines, []byte("\n"))

	return errEnt
}

func (s stackdriverSink) Sync() error {
	s.ec.Flush()

	err := s.l.Flush()
	if err != nil {
		return xerrors.Errorf("failed to flush stackdriver sink: %w", err)
	}
	return nil
}

func sev(level slog.Level) logging.Severity {
	switch level {
	case slog.LevelDebug:
		return logging.Debug
	case slog.LevelInfo:
		return logging.Info
	case slog.LevelWarn:
		return logging.Warning
	case slog.LevelError:
		return logging.Error
	case slog.LevelCritical, slog.LevelFatal:
		return logging.Critical
	default:
		panic(fmt.Sprintf("slogstackdriver: unexpected level %v", level))
	}
}

func (s stackdriverSink) traceField(tID trace.TraceID) string {
	return fmt.Sprintf("projects/%v/traces/%v", s.projectID, tID)
}

func internalIP() string {
	ip, err := metadata.InternalIP()
	if err != nil {
		return err.Error()
	}
	return ip
}

func onError(err error) {
	fmt.Fprintf(os.Stderr, "failed to write logs to stackdriver: %+v\n", err)
}

func errorreportingOnError(err error) {
	fmt.Fprintf(os.Stderr, "failed to write error report to stackdriver: %+v\n", err)
}
