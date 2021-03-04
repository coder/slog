// Package slogjson contains the slogger that writes logs in JSON.
//
// Format
//
//  {
//    "ts": "2019-09-10T20:19:07.159852-05:00",
//    "level": "INFO",
//    "logger_names": ["comp", "subcomp"],
//    "msg": "hi",
//    "caller": "slog/examples_test.go:62",
//    "func": "cdr.dev/slog/sloggers/slogtest_test.TestExampleTest",
//    "trace": "<traceid>",
//    "span": "<spanid>",
//    "fields": {
//      "my_field": "field value"
//    }
//  }
package slogjson // import "cdr.dev/slog/sloggers/slogjson"

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"go.opencensus.io/trace"

	"cdr.dev/slog"
	"cdr.dev/slog/internal/syncwriter"
)

// Sink creates a slog.Sink that writes JSON logs
// to the given writer. See package level docs
// for the format.
// If the writer implements Sync() error then
// it will be called when syncing.
func Sink(w io.Writer) slog.Sink {
	return jsonSink{
		w: syncwriter.New(w),
	}
}

type jsonSink struct {
	w *syncwriter.Writer
}

func (s jsonSink) LogEntry(ctx context.Context, ent slog.SinkEntry) {
	m := slog.M(
		slog.F("ts", ent.Time),
		slog.F("level", ent.Level),
		slog.F("msg", ent.Message),
		slog.F("caller", fmt.Sprintf("%v:%v", ent.File, ent.Line)),
		slog.F("func", ent.Func),
	)

	if len(ent.LoggerNames) > 0 {
		m = append(m, slog.F("logger_names", ent.LoggerNames))
	}

	if ent.SpanContext != (trace.SpanContext{}) {
		m = append(m,
			slog.F("trace", ent.SpanContext.TraceID),
			slog.F("span", ent.SpanContext.SpanID),
		)
	}

	if len(ent.Fields) > 0 {
		m = append(m,
			slog.F("fields", ent.Fields),
		)
	}

	buf, _ := json.Marshal(m)

	buf = append(buf, '\n')
	s.w.Write("slogjson", buf)
}

func (s jsonSink) Sync() {
	s.w.Sync("slogjson")
}
