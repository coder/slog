package sloghuman_test

import (
	"bytes"
	"context"
	"io"
	"strconv"
	"strings"
	"testing"
	"time"

	"cdr.dev/slog"
	"cdr.dev/slog/internal/entryhuman"
	"cdr.dev/slog/sloggers/sloghuman"
	"go.opentelemetry.io/otel/trace"
)

func multiline(n int) string {
	var b strings.Builder
	b.Grow(n * 8)
	for i := 0; i < n; i++ {
		b.WriteString("line-")
		b.WriteString(strconv.Itoa(i))
		b.WriteByte('\n')
	}
	return b.String()
}

// Benchmarks target the human sink path: humanSink.LogEntry -> entryhuman.Fmt -> bufio.Scanner indent.
func BenchmarkHumanSinkLogEntry_SingleLine(b *testing.B) {
	s := sloghuman.Sink(io.Discard)
	ent := slog.SinkEntry{
		Time:    time.Unix(0, 0),
		Level:   slog.LevelInfo,
		Message: "hello world",
		Fields: slog.M(
			slog.F("k1", "v1"),
			slog.F("k2", 123),
		),
	}
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.LogEntry(ctx, ent)
	}
}

func BenchmarkHumanSinkLogEntry_MultilineField_Small(b *testing.B) {
	s := sloghuman.Sink(io.Discard)
	ml := multiline(10)
	ent := slog.SinkEntry{
		Time:  time.Unix(0, 0),
		Level: slog.LevelInfo,
		Fields: slog.M(
			slog.F("msg", "..."),
			slog.F("stack", ml), // forces multiline field handling in Fmt
		),
	}
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.LogEntry(ctx, ent)
	}
}

func BenchmarkHumanSinkLogEntry_MultilineField_Large(b *testing.B) {
	s := sloghuman.Sink(io.Discard)
	ml := multiline(1000) // many short lines; avoids Scanner token limit
	ent := slog.SinkEntry{
		Time:  time.Unix(0, 0),
		Level: slog.LevelInfo,
		Fields: slog.M(
			slog.F("msg", "..."),
			slog.F("stack", ml),
		),
	}
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.LogEntry(ctx, ent)
	}
}

func BenchmarkHumanSinkLogEntry_MultilineMessage(b *testing.B) {
	s := sloghuman.Sink(io.Discard)
	msg := "line1\nline2\nline3\nline4\nline5"
	ent := slog.SinkEntry{
		Time:    time.Unix(0, 0),
		Level:   slog.LevelInfo,
		Message: msg, // triggers multiline message handling in Fmt
		Fields:  nil,
	}
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.LogEntry(ctx, ent)
	}
}

func bmMultiline(n int) string {
	var b strings.Builder
	b.Grow(n * 8)
	for i := 0; i < n; i++ {
		b.WriteString("line-")
		b.WriteString(strconv.Itoa(i))
		b.WriteByte('\n')
	}
	return b.String()
}

func BenchmarkFmt_SingleLine(b *testing.B) {
	ent := slog.SinkEntry{
		Time:    time.Unix(0, 0),
		Level:   slog.LevelInfo,
		Message: "hello world",
		Fields: slog.M(
			slog.F("k1", "v1"),
			slog.F("k2", 123),
		),
	}
	w := io.Discard
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entryhuman.OptimizedFmt(bytes.NewBuffer(nil), w, ent)
	}
}

func BenchmarkFmt_MultilineField_Small(b *testing.B) {
	ml := bmMultiline(10)
	ent := slog.SinkEntry{
		Time:  time.Unix(0, 0),
		Level: slog.LevelInfo,
		Fields: slog.M(
			slog.F("msg", "..."),
			slog.F("stack", ml),
		),
	}
	w := io.Discard
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entryhuman.OptimizedFmt(bytes.NewBuffer(nil), w, ent)
	}
}

func BenchmarkFmt_MultilineField_Large(b *testing.B) {
	ml := bmMultiline(1000)
	ent := slog.SinkEntry{
		Time:  time.Unix(0, 0),
		Level: slog.LevelInfo,
		Fields: slog.M(
			slog.F("msg", "..."),
			slog.F("stack", ml),
		),
	}
	w := io.Discard
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entryhuman.OptimizedFmt(bytes.NewBuffer(nil), w, ent)
	}
}

func BenchmarkFmt_MultilineMessage(b *testing.B) {
	msg := "line1\nline2\nline3\nline4\nline5"
	ent := slog.SinkEntry{
		Time:    time.Unix(0, 0),
		Level:   slog.LevelInfo,
		Message: msg,
	}
	w := io.Discard
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entryhuman.OptimizedFmt(bytes.NewBuffer(nil), w, ent)
	}
}

func BenchmarkFmt_WithNames(b *testing.B) {
	ent := slog.SinkEntry{
		Time:        time.Unix(0, 0),
		Level:       slog.LevelInfo,
		Message:     "msg",
		LoggerNames: []string{"svc", "sub", "component"},
		Fields:      slog.M(slog.F("k", "v")),
	}
	w := io.Discard
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entryhuman.OptimizedFmt(bytes.NewBuffer(nil), w, ent)
	}
}

func BenchmarkFmt_WithSpan(b *testing.B) {
	ent := slog.SinkEntry{
		Time:    time.Unix(0, 0),
		Level:   slog.LevelInfo,
		Message: "msg",
		Fields:  slog.M(slog.F("k", "v")),
	}
	w := io.Discard
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entryhuman.OptimizedFmt(bytes.NewBuffer(nil), w, ent)
	}
}

func BenchmarkFmt_WithValidSpan(b *testing.B) {
	ent := slog.SinkEntry{
		Time:    time.Unix(0, 0),
		Level:   slog.LevelInfo,
		Message: "msg",
		Fields:  slog.M(slog.F("k", "v")),
	}
	// Create a valid SpanContext
	tid := trace.TraceID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	sid := trace.SpanID{1, 2, 3, 4, 5, 6, 7, 8}
	ent.SpanContext = trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    tid,
		SpanID:     sid,
		TraceFlags: trace.FlagsSampled,
	})

	w := io.Discard
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entryhuman.OptimizedFmt(bytes.NewBuffer(nil), w, ent)
	}
}
