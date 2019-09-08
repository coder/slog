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
	if ent.File != "" {
		ents += fmt.Sprintf("%v:%v: ", filepath.Base(ent.File), ent.Line)
	}
	ents += fmt.Sprintf("%v [", ent.Time.Format(timestampMilli))

	if enableColor {
		cl := levelColor(ent.Level)
		ents += color.New(cl).Sprint(ent.Level)
	} else {
		ents += ent.Level.String()
	}

	ents += "]"

	if ent.Component != "" {
		ents += fmt.Sprintf(" (%v)", quote(ent.Component))
	}

	ents += fmt.Sprintf(": %v", quote(ent.Message))

	fields := stringFields(ent)
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

func pinnedFields(ent slog.Entry) string {
	pinned := slogval.Map{}

	if ent.SpanContext != (trace.SpanContext{}) {
		pinned = pinned.Append("trace", slogval.String(ent.SpanContext.TraceID.String()))
		pinned = pinned.Append("span", slogval.String(ent.SpanContext.SpanID.String()))
	}

	return humanFields(pinned)
}

func stringFields(ent slog.Entry) string {
	pinned := pinnedFields(ent)
	fields := humanFields(slogval.Reflect(ent.Fields))

	if pinned == "" {
		return fields
	}

	if fields == "" {
		return pinned
	}

	return pinned + "\n" + fields
}

// Same as time.StampMilli but the days in the month padded by zeros.
const timestampMilli = "Jan 02 15:04:05.000"

func levelColor(level slog.Level) color.Attribute {
	switch level {
	case slog.LevelDebug, slog.LevelInfo:
		return color.FgBlue
	case slog.LevelWarn:
		return color.FgYellow
	case slog.LevelError, slog.LevelCritical, slog.LevelFatal:
		return color.FgRed
	}
	panic("humanfmt: unexpected level: " + string(level))
}

func IsTTY(w io.Writer) bool {
	f, ok := w.(interface {
		Fd() uintptr
	})
	return ok && terminal.IsTerminal(int(f.Fd()))
}
