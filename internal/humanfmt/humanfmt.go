// Package humanfmt contains the code to format slog.Entry
// into a human readable format.
package humanfmt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"go.coder.com/slog"
	"go.coder.com/slog/slogval"
	"go.opencensus.io/trace"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/xerrors"
	"io"
	"path/filepath"
)

// Entry returns a human readable format for ent.
func Entry(ent slog.Entry, enableColor bool) string {
	var ents string
	ts := ent.Time.Format(timestampMilli)
	ents += ts + " "

	level := ent.Level.String()
	if enableColor {
		level = color.New(LevelColor(ent.Level)).Sprint(ent.Level)
	}
	ents += fmt.Sprintf("[%v]\t", level)

	if ent.LoggerName != "" {
		component := quoteKey(ent.LoggerName)
		if enableColor {
			component = color.New(color.FgMagenta).Sprint(component)
		}
		ents += fmt.Sprintf("(%v)\t", component)
	}

	msg := quote(ent.Message)
	if enableColor {
		msg = color.New(color.FgGreen).Sprint(msg)
	}
	ents += fmt.Sprintf("%v\t", msg)

	loc := fmt.Sprintf("%v:%v", filepath.Base(ent.File), ent.Line)
	if enableColor {
		loc = color.New(color.FgCyan).Sprint(loc)
	}
	ents += fmt.Sprintf("<%v> ", loc)

	if ent.SpanContext != (trace.SpanContext{}) {
		ent.Fields = append(slog.Map(
			slog.F("trace", ent.SpanContext.TraceID),
			slog.F("span", ent.SpanContext.SpanID),
		), ent.Fields...)
	}
	slogJSON := true
	visitFn := func(v interface{}, visit slogval.VisitFunc) (_ slogval.Value, ok bool) (nil)
	if slogJSON {
		visitFn = func(v interface{}, visit slogval.VisitFunc) (_ slogval.Value, ok bool) {
			f, ok := v.(xerrors.Formatter)
			if !ok {
				return nil, false
			}
			return slogval.ExtractXErrorChain(f, visit), true
		}
	}
	m, ok := slogval.Encode(ent.Fields, visitFn).(slogval.Map)
	if ok {
		if slogJSON {
			fields, err := json.MarshalIndent(m, "", "")
			if err == nil {
				fields = bytes.ReplaceAll(fields, []byte(",\n"), []byte(", "))
				fields = bytes.ReplaceAll(fields, []byte("\n"), []byte(""))
				fields = JSON(fields)
				ents += "\t" + string(fields)
			}
		} else {
			// We never return with a trailing newline because Go's testing framework adds one
			// automatically and if we include one, then we'll get two newlines.
			// We also do not indent the fields as go's test does that automatically
			// for extra lines in a log so if we did it here, the fields would be indented
			// twice in test logs. So the Stderr logger indents all the fields itself.
			ents += "\n" + fmtVal(m)
		}
	}

	return ents
}

// Same as time.StampMilli but the days in the month padded by zeros.
const timestampMilli = "Jan 02 15:04:05.000"

func LevelColor(level slog.Level) color.Attribute {
	switch level {
	case slog.LevelDebug:
		return color.Reset
	case slog.LevelInfo:
		return color.FgBlue
	case slog.LevelWarn:
		return color.FgYellow
	case slog.LevelError:
		return color.FgRed
	case slog.LevelCritical, slog.LevelFatal:
		return color.FgHiRed
	default:
		panic("humanfmt: unexpected level: " + string(level))
	}
}

// IsTTY checks whether the given writer is a *os.File TTY.
func IsTTY(w io.Writer) bool {
	f, ok := w.(interface {
		Fd() uintptr
	})
	return ok && terminal.IsTerminal(int(f.Fd()))
}
