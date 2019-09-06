package console

import (
	"fmt"
	"path/filepath"

	"go.opencensus.io/trace"

	"go.coder.com/slog/slogcore"
)

func Entry(ent slogcore.Entry) string {
	var ents string
	if ent.File != "" {
		ents += fmt.Sprintf("%v:%v: ", filepath.Base(ent.File), ent.Line)
	}
	ents += fmt.Sprintf("%v [%v]", ent.Time.Format(timestampMilli), ent.Level)

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

func pinnedFields(ent slogcore.Entry) string {
	pinned := slogcore.Map{}

	if ent.SpanContext != (trace.SpanContext{}) {
		pinned = pinned.Append("trace", slogcore.String(ent.SpanContext.TraceID.String()))
		pinned = pinned.Append("span", slogcore.String(ent.SpanContext.SpanID.String()))
	}

	return Fields(pinned)
}

func stringFields(ent slogcore.Entry) string {
	pinned := pinnedFields(ent)
	fields := Fields(ent.Fields)

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

func panicf(f string, v ...interface{}) {
	f = "slogcore: " + f
	s := fmt.Sprintf(f, v...)
	panic(s)
}
