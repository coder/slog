// Package humanfmt contains the code to format slog.Entry
// into a human readable format.
package humanfmt

import (
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"go.coder.com/slog"
	"go.coder.com/slog/slogval"
	"go.opencensus.io/trace"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"path/filepath"
)

// Entry returns a human readable format for ent.
func Entry(ent slog.Entry, enableColor bool) string {
	var ents string
	level := ent.Level.String()
	if enableColor {
		level = LevelColor(ent.Level)
	}
	ents += fmt.Sprintf("[%v] ", level)

	loc := fmt.Sprintf("%v:%v", filepath.Base(ent.File), ent.Line)
	if enableColor {
		loc = color.New(color.FgGreen).Sprint(loc)
	}
	ents += fmt.Sprintf("{%v} ", loc)

	if ent.LoggerName != "" {
		component := quoteKey(ent.LoggerName)
		if enableColor {
			component = color.New(color.FgMagenta).Sprint(component)
		}
		ents += fmt.Sprintf("(%v) ", component)
	}

	ts := ent.Time.Format(timestampMilli)
	ents += ts

	msg := quote(ent.Message)
	ents += fmt.Sprintf(": %v", msg)

	if ent.SpanContext != (trace.SpanContext{}) {
		ent.Fields = append(slog.Map(
			slog.F("trace", ent.SpanContext.TraceID),
			slog.F("span", ent.SpanContext.SpanID),
		), ent.Fields...)
	}
	m, ok := slogval.Encode(ent.Fields).(slogval.Map)
	if ok {
		if slogval.JSONTest {
			var errVal slogval.Field
			if len(m) >= 3 {
				errVal = m[2]
				m = append(m[:2], m[2+1:]...)
			}
			fields, err := json.Marshal(m)
			if err == nil {
				ents += "    " + string(fields)
			}
			if errVal != (slogval.Field{}) {
				ents += "\n" + fmtVal(slogval.Map{
					slogval.Field{"error", errVal.Value},
				})
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

func LevelColor(level slog.Level) string {
	var attr color.Attribute
	switch level {
	case slog.LevelDebug:
		return level.String()
	case slog.LevelInfo:
		attr = color.FgBlue
	case slog.LevelWarn:
		attr = color.FgYellow
	case slog.LevelError:
		attr = color.FgRed
	case slog.LevelCritical, slog.LevelFatal:
		attr = color.FgHiRed
	default:
		panic("humanfmt: unexpected level: " + string(level))
	}
	return color.New(attr).Sprint(level)
}

// IsTTY checks whether the given writer is a *os.File TTY.
func IsTTY(w io.Writer) bool {
	f, ok := w.(interface {
		Fd() uintptr
	})
	return ok && terminal.IsTerminal(int(f.Fd()))
}
