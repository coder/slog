// Package humanfmt contains the code to format slog.Entry
// into a human readable format.
package humanfmt

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/fatih/color"
	"go.opencensus.io/trace"
	"golang.org/x/crypto/ssh/terminal"

	"go.coder.com/slog"
	"go.coder.com/slog/slogval"
)

// Entry returns a human readable format for ent.
func Entry(ent slog.Entry, enableColor bool) string {
	var ents string
	level := ent.Level.String()
	if enableColor {
		level = color.New(levelColor(ent.Level)).Sprint(level)
	}
	ents += fmt.Sprintf("[%v]\t", level)

	loc := fmt.Sprintf("%v:%v", filepath.Base(ent.File), ent.Line)
	if enableColor {
		loc = color.New(color.FgGreen).Sprint(loc)
	}
	ents += fmt.Sprintf("{%v}\t", loc)

	if ent.Component != "" {
		component := quoteKey(ent.Component)
		if enableColor {
			component = color.New(color.FgMagenta).Sprint(component)
		}
		// ents += fmt.Sprintf("(%v)\t", component)
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
	fields := stringFields(ent.Fields)
	if fields != "" {
		// We never return with a trailing newline because Go's testing framework adds one
		// automatically and if we include one, then we'll get two newlines.
		// We also do not indent the fields as go's test does that automatically
		// for extra lines in a log so if we did it here, the fields would be indented
		// twice in test logs. So the Stderr logger indents all the fields itself.
		ents += "\n" + fields
	}

	return ents
}

func stringFields(fields []slog.Field) string {
	m, ok := slogval.Encode(fields).(slogval.Map)
	if ok {
		return humanFields(m)
	}
	return ""
}

// Same as time.StampMilli but the days in the month padded by zeros.
const timestampMilli = "Jan 02 15:04:05.000"

func levelColor(level slog.Level) color.Attribute {
	switch level {
	case slog.LevelDebug, slog.LevelInfo:
		return color.FgBlue
	case slog.LevelWarn:
		return color.FgYellow
	case slog.LevelError:
		return color.FgRed
	case slog.LevelCritical, slog.LevelFatal:
		return color.FgHiRed
	}
	panic("humanfmt: unexpected level: " + string(level))
}

// IsTTY checks whether the given writer is a *os.File TTY.
func IsTTY(w io.Writer) bool {
	f, ok := w.(interface {
		Fd() uintptr
	})
	return ok && terminal.IsTerminal(int(f.Fd()))
}
