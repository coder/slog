// Package entryhuman contains the code to format slog.SinkEntry
// for humans.
package entryhuman

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"go.opencensus.io/trace"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/xerrors"

	"cdr.dev/slog"
)

// StripTimestamp strips the timestamp from entry and returns
// it and the rest of the entry.
func StripTimestamp(ent string) (time.Time, string, error) {
	ts := ent[:len(TimeFormat)]
	rest := ent[len(TimeFormat):]
	et, err := time.Parse(TimeFormat, ts)
	return et, rest, err
}

// TimeFormat is a simplified RFC3339 format.
const TimeFormat = "2006-01-02 15:04:05.000"

var (
	renderer = lipgloss.NewRenderer(os.Stdout, termenv.WithUnsafe())

	loggerNameStyle = renderer.NewStyle().Foreground(lipgloss.Color("#A47DFF"))
	timeStyle       = renderer.NewStyle().Foreground(lipgloss.Color("#606366"))
)

func render(w io.Writer, st lipgloss.Style, s string) string {
	if shouldColor(w) {
		ss := st.Render(s)
		return ss
	}
	return s
}

func reset(w io.Writer, termW io.Writer) {
	if shouldColor(termW) {
		fmt.Fprintf(w, termenv.CSI+termenv.ResetSeq+"m")
	}
}

// Fmt returns a human readable format for ent.
//
// We never return with a trailing newline because Go's testing framework adds one
// automatically and if we include one, then we'll get two newlines.
// We also do not indent the fields as go's test does that automatically
// for extra lines in a log so if we did it here, the fields would be indented
// twice in test logs. So the Stderr logger indents all the fields itself.
func Fmt(
	buf interface {
		io.StringWriter
		io.Writer
	}, termW io.Writer, ent slog.SinkEntry,
) {
	reset(buf, termW)
	ts := ent.Time.Format(TimeFormat)
	buf.WriteString(render(termW, timeStyle, ts+" "))

	level := ent.Level.String()
	if len(level) > 4 {
		level = level[:4]
	}
	level = "[" + level + "]"
	buf.WriteString(render(termW, levelStyle(ent.Level), level))
	buf.WriteString("\t")

	if len(ent.LoggerNames) > 0 {
		loggerName := "(" + quoteKey(strings.Join(ent.LoggerNames, ".")) + ")"
		buf.WriteString(render(termW, loggerNameStyle, loggerName))
		buf.WriteString("\t")
	}

	var multilineKey string
	var multilineVal string
	msg := strings.TrimSpace(ent.Message)
	if strings.Contains(msg, "\n") {
		multilineKey = "msg"
		multilineVal = msg
		msg = "..."
		msg = quote(msg)
	}
	buf.WriteString(msg)

	if ent.SpanContext != (trace.SpanContext{}) {
		ent.Fields = append(slog.M(
			slog.F("trace", ent.SpanContext.TraceID),
			slog.F("span", ent.SpanContext.SpanID),
		), ent.Fields...)
	}

	for i, f := range ent.Fields {
		if multilineVal != "" {
			break
		}

		var s string
		switch v := f.Value.(type) {
		case string:
			s = v
		case error, xerrors.Formatter:
			s = fmt.Sprintf("%+v", v)
		}
		s = strings.TrimSpace(s)
		if !strings.Contains(s, "\n") {
			continue
		}

		// Remove this field.
		ent.Fields = append(ent.Fields[:i], ent.Fields[i+1:]...)
		multilineKey = f.Name
		multilineVal = s
	}

	// Basic keyStyle off of the level makes it easy to distinguish individual
	// entries in a fast stream of logs where some are multi-line.
	// See logrus for an example.
	keyStyle := levelStyle(ent.Level).Copy().Bold(false)

	for i, f := range ent.Fields {
		if i < len(ent.Fields) {
			buf.WriteString("\t")
		}
		buf.WriteString(render(termW, keyStyle, quoteKey(f.Name)+"="))
		valueStr := fmt.Sprintf("%+v", f.Value)
		buf.WriteString(quote(valueStr))
	}

	if multilineVal != "" {
		if msg != "..." {
			buf.WriteString(" ...")
		}

		// Proper indentation.
		lines := strings.Split(multilineVal, "\n")
		for i, line := range lines[1:] {
			if line != "" {
				lines[i+1] = strings.Repeat(" ", len(multilineKey)+2) + line
			}
		}
		multilineVal = strings.Join(lines, "\n")

		multilineKey = render(termW, keyStyle, multilineKey)
		buf.WriteString("\n")
		buf.WriteString(multilineKey)
		buf.WriteString("= ")
		buf.WriteString(multilineVal)
	}
}

var (
	levelDebugStyle = renderer.NewStyle().Foreground(lipgloss.Color("#ffffff"))
	levelInfoStyle  = renderer.NewStyle().Foreground(lipgloss.Color("#0091FF"))
	levelWarnStyle  = renderer.NewStyle().Foreground(lipgloss.Color("#FFCF0D"))
	levelErrorStyle = renderer.NewStyle().Foreground(lipgloss.Color("#FF5A0D")).Bold(true)
)

func levelStyle(level slog.Level) lipgloss.Style {
	switch level {
	case slog.LevelDebug:
		return levelDebugStyle
	case slog.LevelInfo:
		return levelInfoStyle
	case slog.LevelWarn:
		return levelWarnStyle
	case slog.LevelError, slog.LevelFatal, slog.LevelCritical:
		return levelErrorStyle
	default:
		panic("unknown level")
	}
}

var forceColorWriter = io.Writer(&bytes.Buffer{})

// isTTY checks whether the given writer is a *os.File TTY.
func isTTY(w io.Writer) bool {
	if w == forceColorWriter {
		return true
	}
	f, ok := w.(interface {
		Fd() uintptr
	})
	return ok && terminal.IsTerminal(int(f.Fd()))
}

func shouldColor(w io.Writer) bool {
	return isTTY(w) || os.Getenv("FORCE_COLOR") != ""
}

// quotes quotes a string so that it is suitable
// as a key for a map or in general some output that
// cannot span multiple lines or have weird characters.
func quote(key string) string {
	// strconv.Quote does not quote an empty string so we need this.
	if key == "" {
		return `""`
	}

	var hasSpace bool
	for _, r := range key {
		if unicode.IsSpace(r) {
			hasSpace = true
			break
		}
	}
	quoted := strconv.Quote(key)
	// If the key doesn't need to be quoted, don't quote it.
	// We do not use strconv.CanBackquote because it doesn't
	// account tabs.
	if !hasSpace && quoted[1:len(quoted)-1] == key {
		return key
	}
	return quoted
}

func quoteKey(key string) string {
	// Replace spaces in the map keys with underscores.
	return quote(strings.ReplaceAll(key, " ", "_"))
}
