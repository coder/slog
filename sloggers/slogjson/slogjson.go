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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"go.coder.com/slog"
	"go.coder.com/slog/internal/humanfmt"
	"go.coder.com/slog/internal/syncwriter"
	"go.coder.com/slog/slogval"
	"go.opencensus.io/trace"
	"golang.org/x/xerrors"
	"io"
	"os"
)

// Make creates a logger that writes JSON logs
// to the given writer. See package level docs
// for the format.
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

func (s jsonSink) LogEntry(ctx context.Context, ent slog.Entry) {
	m := slog.Map(
		slog.F("ts", ent.Time),
		slog.F("level", ent.Level),
		slog.F("component", ent.LoggerName),
		slog.F("msg", ent.Message),
		slog.F("caller", fmt.Sprintf("%v:%v", ent.File, ent.Line)),
		slog.F("func", ent.Func),
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

	v := slogval.Encode(m, func(v interface{}, visit slogval.VisitFunc) (_ slogval.Value, ok bool) {
		f, ok := v.(xerrors.Formatter)
		if !ok {
			return nil, false
		}
		return slogval.ExtractXErrorChain(f, visit), true
	})
	buf, err := json.MarshalIndent(v, "", "")
	if err != nil {
		os.Stderr.WriteString("slogjson: failed to encode entry to JSON: " + err.Error())
		return
	}
	lines := bytes.SplitN(buf, []byte("\n"), 6)
	if s.color {
		colorField(lines, 1, color.FgCyan)
		if ent.Level != slog.LevelDebug {
			colorField(lines, 2, humanfmt.LevelColor(ent.Level))
		}
		colorField(lines, 3, color.FgMagenta)
		colorField(lines, 4, color.FgGreen)
	}
	buf = bytes.Join(lines, []byte("\n"))
	buf = bytes.ReplaceAll(buf, []byte(",\n"), []byte(", "))
	buf = bytes.ReplaceAll(buf, []byte("\n"), []byte(""))
	buf = append(buf, '\n')

	buf = append([]byte("    "), buf...)

	_, err = s.w.Write(buf)
	if err != nil {
		os.Stderr.WriteString("slogjson: failed to write JSON entry: " + err.Error())
		return
	}
}

func colorField(lines [][]byte, n int, attr color.Attribute) {
	og := []byte(`: "`)
	repl := []byte(`: "` + color.New(attr).Sprint()[:5])

	lines[n] = bytes.Replace(lines[n], og, repl, 1)

	l := lines[n]
	lines[n] = append(l[:len(l)-2], []byte(color.New(color.Reset).Sprint()+`",`)...)
}
