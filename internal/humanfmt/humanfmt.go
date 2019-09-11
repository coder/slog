// Package humanfmt contains the code to format slog.Entry
// into a human readable format.
package humanfmt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"go.opencensus.io/trace"
	"golang.org/x/crypto/ssh/terminal"

	"go.coder.com/slog"
	"go.coder.com/slog/slogval"
)

// Entry returns a human readable format for ent.
func Entry(ent slog.Entry, enableColor bool) string {
	var ents string
	ts := ent.Time.Format(timestampMilli)
	ents += ts + " "

	level := ent.Level.String()
	if enableColor {
		level = color.New(levelColor(ent.Level)).Sprint(ent.Level)
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
	m, ok := slogval.Encode(ent.Fields, nil).(slogval.Map)
	if ok {
		fields, err := json.MarshalIndent(m, "", "")
		if err == nil {
			fields = bytes.ReplaceAll(fields, []byte(",\n"), []byte(", "))
			fields = bytes.ReplaceAll(fields, []byte("\n"), []byte(""))
			fields = highlightJSON(fields)
			ents += "\t" + string(fields)
		} else {
			ents += fmt.Sprintf("\thumanfmt: failed to marshal fields: %+v", err)
		}
	}

	// We never return with a trailing newline because Go's testing framework adds one
	// automatically and if we include one, then we'll get two newlines.
	// We also do not indent the fields as go's test does that automatically
	// for extra lines in a log so if we did it here, the fields would be indented
	// twice in test logs. So the Stderr logger indents all the fields itself.
	return ents
}

// Same as time.StampMilli but the days in the month padded by zeros.
const timestampMilli = "Jan 02 15:04:05.000"

func levelColor(level slog.Level) color.Attribute {
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

// quotes quotes a string so that it is suitable
// as a key for a map or in general some output that
// cannot span multiple lines or have weird characters.
func quote(key string) string {
	// strconv.Quote does not quote an empty string so we need this.
	if key == "" {
		return `""`
	}

	quoted := strconv.Quote(key)
	// If the key doesn't need to be quoted, don't quote it.
	// We do not use strconv.CanBackquote because it doesn't
	// account tabs.
	if quoted[1:len(quoted)-1] == key {
		return key
	}
	return quoted
}

func quoteKey(key string) string {
	// Replace spaces in the map keys with underscores.
	return strings.ReplaceAll(key, " ", "_")
}
