// Package slogjson contains the slogger that writes logs in a JSON format.
//
// Format
//
//  {
//    "level": "INFO",
//    "msg": "hi",
//    "ts": "",
//    "caller": "slog/examples_test.go:62",
//    "func": "go.coder.com/slog/sloggers/slogtest_test.TestExampleTest",
//    "component": "comp.subcomp",
//    "trace": "<traceid>",
//    "span": "<spanid>",
//    "fields": {
//      "myField": "fieldValue"
//    }
//  }
package slogjson // import "go.coder.com/slog/sloggers/slogjson"

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"go.opencensus.io/trace"
	"golang.org/x/xerrors"

	"go.coder.com/slog"
	"go.coder.com/slog/internal/syncwriter"
	"go.coder.com/slog/slogval"
)

// Make creates a logger that writes JSON logs
// to the given writer. See package level docs
// for the format.
func Make(w io.Writer) slog.Logger {
	return slog.Make(jsonSink{
		w: syncwriter.New(w),
	})
}

type jsonSink struct {
	w *syncwriter.Writer
}

func (s jsonSink) LogEntry(ctx context.Context, ent slog.Entry) {
	m := slog.Map(
		slog.F("level", ent.Level),
		slog.F("msg", ent.Message),
		slog.F("ts", jsonTimestamp(ent.Time)),
		slog.F("caller", fmt.Sprintf("%v:%v", ent.File, ent.Line)),
		slog.F("func", ent.Func),
		slog.F("component", ent.Component),
	)

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

	v := slogval.Reflect(m)
	// We use NewEncoder because it reuses buffers behind the scenes which we cannot
	// do with json.Marshal.
	e := json.NewEncoder(s.w)
	e.Encode(v)
}

func jsonTimestamp(t time.Time) interface{} {
	ts, err := t.MarshalText()
	if err != nil {
		return xerrors.Errorf("failed to marshal timestamp to text: %w", err)
	}
	return string(ts)
}
