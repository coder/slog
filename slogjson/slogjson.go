// Package slogjson contains the slogger that writes logs in a JSON format.
//
// Format
//
//  {
//    "ts": "2019-09-10T20:19:07.159852-05:00",
//    "level": "INFO",
//    "component": "comp.subcomp",
//    "msg": "hi",
//    "caller": "slog/examples_test.go:62",
//    "func": "cdr.dev/slog/sloggers/slogtest_test.TestExampleTest",
//    "trace": "<traceid>",
//    "span": "<spanid>",
//    "fields": {
//      "myField": "fieldValue"
//    }
//  }
package slogjson // import "cdr.dev/slog/sloggers/slogjson"

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"go.opencensus.io/trace"
	"golang.org/x/xerrors"

	"cdr.dev/slog"
	"cdr.dev/slog/internal/humanfmt"
	"cdr.dev/slog/internal/syncwriter"
)

// Make creates a logger that writes JSON logs
// to the given writer. See package level docs
// for the format.
// If the writer implements Sync() error then
// it will be called when syncing.
func Make(w io.Writer) slog.Logger {
	return slog.Make(jsonSink{
		w:     syncwriter.New(w),
		color: humanfmt.IsTTY(w),
	})
}

type jsonSink struct {
	w     *syncwriter.Writer
	color bool
}

func (s jsonSink) LogEntry(ctx context.Context, ent slog.SinkEntry) error {
	m := slog.Map{
		{"ts", ent.Time},
		{"level", ent.Level},
		{"component", ent.LoggerName},
		{"msg", ent.Message},
		{"caller", fmt.Sprintf("%v:%v", ent.File, ent.Line)},
		{"func", ent.Func},
	}

	if ent.SpanContext != (trace.SpanContext{}) {
		m = append(m,
			slog.F{"trace", ent.SpanContext.TraceID},
			slog.F{"span", ent.SpanContext.SpanID},
		)
	}

	if len(ent.Fields) > 0 {
		m = append(m,
			slog.F{"fields", ent.Fields},
		)
	}

	buf, err := json.Marshal(m)
	if err != nil {
		return xerrors.Errorf("slogjson: failed to encode entry to JSON: %w", err)
	}

	buf = append(buf, '\n')
	_, err = s.w.Write(buf)
	if err != nil {
		return xerrors.Errorf("slogjson: failed to write JSON entry: %w", err)
	}
	return nil
}

func (s jsonSink) Sync() error {
	return s.w.Sync()
}
