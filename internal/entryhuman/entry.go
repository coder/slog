// Package entryhuman contains the code to format slog.SinkEntry
// for humans.
package entryhuman

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"golang.org/x/term"
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

	timeStyle = renderer.NewStyle().Foreground(lipgloss.Color("#606366"))
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

func formatValue(v interface{}) string {
	if vr, ok := v.(driver.Valuer); ok {
		var err error
		v, err = vr.Value()
		if err != nil {
			return fmt.Sprintf("error calling Value: %v", err)
		}
	}
	if v == nil {
		return "<nil>"
	}
	typ := reflect.TypeOf(v)
	switch typ.Kind() {
	case reflect.Struct, reflect.Map:
		byt, err := json.Marshal(v)
		if err != nil {
			panic(err)
		}
		return string(byt)
	case reflect.Slice:
		// Byte slices are optimistically readable.
		if typ.Elem().Kind() == reflect.Uint8 {
			return fmt.Sprintf("%q", v)
		}
		fallthrough
	default:
		return quote(fmt.Sprintf("%+v", v))
	}
}

const tab = "  "

// bracketedLevel is an optimization to avoid extra allocations and calls to strings.ToLower
// when we want to translate/print the lowercase version of a log level.
func bracketedLevel(l slog.Level) string {
	switch l {
	case slog.LevelDebug:
		return "[debu]"
	case slog.LevelInfo:
		return "[info]"
	case slog.LevelWarn:
		return "[warn]"
	case slog.LevelError:
		return "[erro]"
	case slog.LevelCritical:
		return "[crit]"
	case slog.LevelFatal:
		return "[fata]"
	default:
		return "[unkn]"
	}
}

func writeSignedInt(w io.Writer, n int64) (bool, error) {
	var a [20]byte
	_, err := w.Write(strconv.AppendInt(a[:0], n, 10))
	return true, err
}

func writeUnsignedInt(w io.Writer, n uint64) (bool, error) {
	var a [20]byte
	_, err := w.Write(strconv.AppendUint(a[:0], n, 10))
	return true, err
}

// Optimization to avoid allocation of heap allocations/temporary strings via formatValue when dealing with primitive types.
// It returns (handled, error). When handled is false, the caller should fall back to formatValue.
func writeValueFast(w io.Writer, v interface{}) (bool, error) {
	switch x := v.(type) {
	case string:
		_, err := w.Write([]byte(quote(x)))
		return true, err
	case bool:
		if x {
			_, err := w.Write([]byte("true"))
			return true, err
		}
		_, err := w.Write([]byte("false"))
		return true, err

	// signed ints
	case int:
		return writeSignedInt(w, int64(x))
	case int8:
		return writeSignedInt(w, int64(x))
	case int16:
		return writeSignedInt(w, int64(x))
	case int32:
		return writeSignedInt(w, int64(x))
	case int64:
		return writeSignedInt(w, x)

	// unsigned ints
	case uint:
		return writeUnsignedInt(w, uint64(x))
	case uint8:
		return writeUnsignedInt(w, uint64(x))
	case uint16:
		return writeUnsignedInt(w, uint64(x))
	case uint32:
		return writeUnsignedInt(w, uint64(x))
	case uint64:
		return writeUnsignedInt(w, x)

	// floats: prefer 'g' to keep output bounded (matches fmt default)
	case float32:
		var a [32]byte
		_, err := w.Write(strconv.AppendFloat(a[:0], float64(x), 'g', -1, 32))
		return true, err
	case float64:
		var a [32]byte
		_, err := w.Write(strconv.AppendFloat(a[:0], x, 'g', -1, 64))
		return true, err
	default:
		return false, nil
	}
}

// Fmt returns a human readable format for ent. Assumes we have a bytes.Buffer
// which we will more easily be able to assume underlying reallocation of it's size is possible
// if necessary than for an arbitrary io.Writer/io.StringWriter
// Note that while bytes.Buffer can in theory return an error for writes, it only does so if the buffer size will
// exceed our architectures max integer size. If the system is actually OOM and more memory cannot be allocated
// it will panic instead.
//
// We never return with a trailing newline because Go's testing framework adds one
// automatically and if we include one, then we'll get two newlines.
// We also do not indent the fields as go's test does that automatically
// for extra lines in a log so if we did it here, the fields would be indented
// twice in test logs. So the Stderr logger indents all the fields itself.
func Fmt(buf *bytes.Buffer, termW io.Writer, ent slog.SinkEntry) {
	reset(buf, termW)

	// Timestamp + space
	buf.WriteString(render(termW, timeStyle, ent.Time.Format(TimeFormat)))
	buf.WriteString(" ")

	// Level label + two spaces
	lvl := bracketedLevel(ent.Level) // e.g. "[debu]", "[info]"
	buf.WriteString(render(termW, levelStyle(ent.Level), lvl))
	buf.WriteString("  ")

	// Logger names: name1.name2.name3: (no strings.Join allocation)
	if len(ent.LoggerNames) > 0 {
		for i, name := range ent.LoggerNames {
			if i > 0 {
				buf.WriteString(".")
			}
			buf.WriteString(quoteKey(name))
		}
		buf.WriteString(": ")
	}

	// Message (detect multiline)
	var multilineKey string
	var multilineVal string
	msg := strings.TrimSpace(ent.Message)
	if strings.Contains(msg, "\n") {
		multilineKey = "msg"
		multilineVal = msg
		msg = quote("...")
	}
	buf.WriteString(msg)

	keyStyle := timeStyle
	equalsStyle := timeStyle

	// Write trace/span directly (do not mutate ent.Fields)
	if ent.SpanContext.IsValid() {
		buf.WriteString(tab)

		buf.WriteString(render(termW, keyStyle, quoteKey("trace")))
		buf.WriteString(render(termW, equalsStyle, "="))
		buf.WriteString(ent.SpanContext.TraceID().String())
		buf.WriteString(tab)
		buf.WriteString(render(termW, keyStyle, quoteKey("span")))
		buf.WriteString(render(termW, equalsStyle, "="))
		buf.WriteString(ent.SpanContext.SpanID().String())
	}

	// Find a multiline field without mutating ent.Fields.
	multiIdx := -1
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
		multiIdx = i
		multilineKey = f.Name
		multilineVal = s
		break
	}

	// Print fields (skip multiline field index).
	for i, f := range ent.Fields {
		if i == multiIdx {
			continue
		}
		if i < len(ent.Fields) {
			buf.WriteString(tab)
		}

		buf.WriteString(render(termW, keyStyle, quoteKey(f.Name)))
		buf.WriteString(render(termW, equalsStyle, "="))

		if ok, err := writeValueFast(buf, f.Value); err != nil {
			// return err
		} else if !ok {
			buf.WriteString(formatValue(f.Value))
		}
	}

	// Multiline value block
	if multilineVal != "" {
		if msg != "..." {
			buf.WriteString(" ...")
		}

		buf.WriteString("\n")
		buf.WriteString(render(termW, keyStyle, multilineKey))
		buf.WriteString("= ")

		// First line up to first newline
		s := multilineVal
		if n := strings.IndexByte(s, '\n'); n >= 0 {
			buf.WriteString(s[:n])
			s = s[n+1:]
		} else {
			buf.WriteString(s)
			s = ""
		}

		indent := strings.Repeat(" ", len(multilineKey)+2)
		for len(s) > 0 {
			buf.WriteString("\n")
			// Only indent non-empty lines.
			if s[0] != '\n' {
				buf.WriteString(indent)
			}
			if n := strings.IndexByte(s, '\n'); n >= 0 {
				buf.WriteString(s[:n])
				s = s[n+1:]
			} else {
				buf.WriteString(s)
				break
			}
		}
	}
}

var (
	levelDebugStyle = timeStyle.Copy()
	levelInfoStyle  = renderer.NewStyle().Foreground(lipgloss.Color("#0091FF"))
	levelWarnStyle  = renderer.NewStyle().Foreground(lipgloss.Color("#FFCF0D"))
	levelErrorStyle = renderer.NewStyle().Foreground(lipgloss.Color("#FF5A0D"))
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
	// SyscallConn is safe during file close.
	if sc, ok := w.(interface {
		SyscallConn() (syscall.RawConn, error)
	}); ok {
		conn, err := sc.SyscallConn()
		if err != nil {
			return false
		}
		var isTerm bool
		err = conn.Control(func(fd uintptr) {
			isTerm = term.IsTerminal(int(fd))
		})
		if err != nil {
			return false
		}
		return isTerm
	}
	// Fallback to unsafe Fd.
	f, ok := w.(interface{ Fd() uintptr })
	return ok && term.IsTerminal(int(f.Fd()))
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
