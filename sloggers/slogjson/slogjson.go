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
	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters"
	"go.coder.com/slog/internal/humanfmt"
	"io"
	"os"
	"time"

	jlexers "github.com/alecthomas/chroma/lexers/j"
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
		slog.F("level", ent.Level),
		slog.F("msg", ent.Message),
		slog.F("component", ent.Component),
		slog.F("caller", fmt.Sprintf("%v:%v", ent.File, ent.Line)),
		slog.F("func", ent.Func),
		slog.F("ts", jsonTimestamp(ent.Time)),
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
	buf, err := json.Marshal(v)
	if err != nil {
		os.Stderr.WriteString("slogjson: failed to encode entry to JSON: " + err.Error())
		return
	}
	buf = append(buf, '\n')
	if s.color {
		jsonLexer := chroma.Coalesce(jlexers.JSON)
		it, err := jsonLexer.Tokenise(nil, string(buf))
		if err != nil {
			os.Stderr.WriteString("slogjson: failed to tokenize JSON entry: " + err.Error())
			return
		}
		b := bytes.NewBuffer(buf)
		err = formatters.TTY8.Format(b, nhooyrJSON, it)
		if err != nil {
			os.Stderr.WriteString("slogjson: failed to format JSON entry: " + err.Error())
			return
		}
		buf = b.Bytes()
	}

	_, err = s.w.Write(buf)
	if err != nil {
		os.Stderr.WriteString("slogjson: failed to write JSON entry: " + err.Error())
	}
}

func jsonTimestamp(t time.Time) interface{} {
	ts, err := t.MarshalText()
	if err != nil {
		return xerrors.Errorf("failed to marshal timestamp to text: %w", err)
	}
	return string(ts)
}

// Adapted from https://github.com/alecthomas/chroma/blob/2f5349aa18927368dbec6f8c11608bf61c38b2dd/styles/bw.go#L7
// https://github.com/alecthomas/chroma/blob/2f5349aa18927368dbec6f8c11608bf61c38b2dd/formatters/tty_indexed.go
// https://github.com/alecthomas/chroma/blob/2f5349aa18927368dbec6f8c11608bf61c38b2dd/lexers/j/json.go
var nhooyrJSON = chroma.MustNewStyle("nhooyrJSON", chroma.StyleEntries{
	// Magenta.
	chroma.Keyword: "#7f007f",
	// Magenta.
	chroma.Number: "#7f007f",
	// Green.
	chroma.String: "#007f00",
})
