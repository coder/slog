package core

import (
	"context"
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"time"

	structpb "github.com/golang/protobuf/ptypes/struct"
	"go.opencensus.io/trace"
	"golang.org/x/xerrors"
	logpb "google.golang.org/genproto/googleapis/logging/v2"

	"go.coder.com/m/lib/log"
	"go.coder.com/m/lib/m"
)

type Severity string

const (
	Debug    Severity = "DEBUG"
	Info     Severity = "INFO"
	Warn     Severity = "WARN"
	Error    Severity = "ERROR"
	Critical Severity = "CRITICAL"
	Fatal    Severity = "FATAL"
)

// The core logger implementation.
type Logger struct {
	fields map[string]*structpb.Value
}

type Entry struct {
	Time time.Time

	Sev Severity
	Msg string

	Component string
	ID        string
	Loc       *logpb.LogEntrySourceLocation
	SpanCtx   trace.SpanContext

	Fields map[string]*structpb.Value
}

// Strips and validates the timestamp from an entry to allow comparing entries
// deterministically in tests.
func StripEntryTimestamp(ent string) (string, error) {
	end := strings.Index(ent, " [")
	if end == -1 {
		return "", xerrors.New("entry is missing severity")
	}

	if end < len(timestampMilli) {
		return "", xerrors.Errorf("entry %q too small to have a timestamp", ent)
	}

	start := end - len(timestampMilli)
	timestamp := ent[start:end]

	t, err := time.Parse(timestampMilli, timestamp)
	if err != nil {
		return "", xerrors.Errorf("failed to parse timestamp %q in entry: %v", timestamp, err)
	}

	if t.IsZero() {
		return "", xerrors.New("entry has timestamp with zero value")
	}

	// Skip space after.
	before := ent[:start]
	after := ent[end+1:]

	return before + after, nil
}

func (ent Entry) pinnedFields() string {
	pinned := map[string]*structpb.Value{}

	if ent.ID != "" {
		pinned[log.ID] = PBString(ent.ID)
	}

	if ent.SpanCtx != (trace.SpanContext{}) {
		pinned[traceField] = PBString(ent.SpanCtx.TraceID.String())
		pinned[span] = PBString(ent.SpanCtx.SpanID.String())
	}

	return marshalFields(pinned)
}

func (ent Entry) stringFields() string {
	pinned := ent.pinnedFields()
	fields := marshalFields(ent.Fields)

	if pinned == "" {
		return fields
	}

	if fields == "" {
		return pinned
	}

	return pinned + "\n" + fields
}

// time.StampMilli but with 0 instead of a space before single digit month days.
const timestampMilli = "Jan 02 15:04:05.000"

func (ent Entry) String() string {
	var ents string
	if ent.Loc.File != "" {
		ents += fmt.Sprintf("%v:%v: ", filepath.Base(ent.Loc.File), ent.Loc.Line)
	}
	ents += fmt.Sprintf("%v [%v]", ent.Time.Format(timestampMilli), ent.Sev)

	if ent.Component != "" {
		ents += fmt.Sprintf(" (%v)", quote(ent.Component))
	}

	ents += fmt.Sprintf(": %v", quote(ent.Msg))

	fields := ent.stringFields()
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

type EntryConfig struct {
	Sev    Severity
	Msg    string
	Fields map[string]interface{}
	Skip   int
}

func (l Logger) Make(ctx context.Context, config EntryConfig) Entry {
	l = l.withContext(ctx)
	l = l.With(config.Fields)

	// Need to make sure any code that sets a field does not panic.
	// E.g. the setting of the message key in the stackdriver logger.
	if l.fields == nil {
		l.fields = make(map[string]*structpb.Value)
	}

	ent := Entry{
		Time:      time.Now(),
		Sev:       config.Sev,
		Component: l.component(),
		Msg:       config.Msg,
		SpanCtx:   trace.FromContext(ctx).SpanContext(),
		Fields:    l.fields,
		ID:        l.fields[log.ID].GetStringValue(),
	}

	delete(l.fields, log.Component)
	delete(l.fields, log.ID)

	ent.Loc = Location(config.Skip + 1 + skipFrom(ctx))
	return ent
}

func Location(skip int) *logpb.LogEntrySourceLocation {
	pc, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		panicf("failed to get caller information with skip %v", skip)
	}
	f := runtime.FuncForPC(pc)
	return &logpb.LogEntrySourceLocation{
		File:     file,
		Line:     int64(line),
		Function: f.Name(),
	}
}

func panicf(f string, v ...interface{}) {
	f = "log: " + f
	s := fmt.Sprintf(f, v...)
	panic(s)
}

func (l Logger) cloneFields() Logger {
	l2 := l
	l2.fields = make(map[string]*structpb.Value, len(l.fields))
	for k, v := range l.fields {
		l2.fields[k] = v
	}
	return l2
}

func (l Logger) With(fields map[string]interface{}) Logger {
	if len(fields) == 0 {
		return l
	}

	l = l.cloneFields()

	for k, v := range fields {
		switch k {
		case Message, span, traceField:
			panicf(`log key %q is reserved and cannot be used, usage is on value %#v of type %T`, k, v, v)
		case log.Component:
			name, ok := v.(string)
			if !ok {
				panicf("value for log key log.Component must always be a String but is instead %T and value %#v", v, v)
			}
			l.appendComponent(name)
			continue
		case log.ID:
			id, ok := v.(string)
			if !ok {
				panicf(`log key log.ID must always be a String but is instead %T and value %#v"`, v, v)
			}
			l.fields[log.ID] = PBString(id)
			continue
		}

		rv := reflect.ValueOf(v)
		l.fields[k] = pbval(rv)
	}
	return l
}

func (l Logger) component() string {
	return l.fields[log.Component].GetStringValue()
}

func (l Logger) appendComponent(name string) {
	if name == "" {
		return
	}

	if l.component() == "" {
		l.fields[log.Component] = PBString(name)
		return
	}

	l.fields[log.Component] = PBString(l.component() + "." + name)
}

func (l Logger) withContext(ctx context.Context) Logger {
	l2 := fromContext(ctx)
	if len(l2.fields) == 0 {
		return l
	}

	l = l.cloneFields()
	for k, v := range l2.fields {
		if k == log.Component {
			l.appendComponent(v.GetStringValue())
			continue
		}
		l.fields[k] = v
	}
	return l
}

// Special field keys for Logger.
// If this is ever updated, update the link in the docs on the Logger interface.
const (
	// Special for stackdriver so a user cannot use it.
	Message = "message"

	// Fields for logging trace info in text format.
	traceField = "trace"
	span       = "span"
)

func Inspect(ctx context.Context, l log.Logger, v ...interface{}) {
	m := make(m.M, len(v))
	for i, v := range v {
		m[fmt.Sprintf("%v", i)] = v
	}
	ctx = WithSkip(ctx, 1)
	l.Debug(ctx, "inspect", m)
}
