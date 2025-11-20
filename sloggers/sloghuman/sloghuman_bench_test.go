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
func BenchmarkHumanSinkLogEntry(b *testing.B) {
	s := sloghuman.Sink(io.Discard)
	testcases := []struct {
		name  string
		entry slog.SinkEntry
	}{
		{
			"SingleLine",
			slog.SinkEntry{
				Time:    time.Unix(0, 0),
				Level:   slog.LevelInfo,
				Message: "hello world",
				Fields: slog.M(
					slog.F("k1", "v1"),
					slog.F("k2", 123),
				),
			},
		},
		{
			"MultiLineFieldSmall",
			slog.SinkEntry{
				Time:  time.Unix(0, 0),
				Level: slog.LevelInfo,
				Fields: slog.M(
					slog.F("msg", "..."),
					slog.F("stack", multiline(10)), // forces multiline field handling in Fmt
				),
			},
		},
		{
			"MultilineMultifieldLarge",
			slog.SinkEntry{
				Time:  time.Unix(0, 0),
				Level: slog.LevelInfo,
				Fields: slog.M(
					slog.F("msg", "..."),
					slog.F("stack", multiline(1000)),
				),
			},
		},
		{
			"MultilineMessage",
			slog.SinkEntry{
				Time:    time.Unix(0, 0),
				Level:   slog.LevelInfo,
				Message: "line1\nline2\nline3\nline4\nline5",
				Fields:  nil,
			},
		},
		{
			"StringHeavy",
			slog.SinkEntry{
				Time:    time.Unix(0, 0),
				Level:   slog.LevelInfo,
				Message: "user action",
				Fields: slog.M(
					slog.F("user_id", "abc-123-def-456-ghi-789"),
					slog.F("action", "login"),
					slog.F("ip_address", "192.168.1.100"),
					slog.F("user_agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"),
					slog.F("session_id", "sess_9d8f7a6b5c4e3d2a1b0c"),
					slog.F("request_id", "req_1a2b3c4d5e6f7g8h9i0j"),
					slog.F("endpoint", "/api/v1/users/login"),
					slog.F("method", "POST"),
				),
			},
		},
	}
	for _, tc := range testcases {
		b.Run(tc.name, func(b *testing.B) {
			ctx := context.Background()
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				s.LogEntry(ctx, tc.entry)
			}
		})
	}
}

func genMultiline(n int) string {
	var b strings.Builder
	b.Grow(n * 8)
	for i := 0; i < n; i++ {
		b.WriteString("line-")
		b.WriteString(strconv.Itoa(i))
		b.WriteByte('\n')
	}
	return b.String()
}

func BenchmarkFmt(b *testing.B) {
	testcases := []struct {
		name  string
		entry slog.SinkEntry
	}{
		{
			"WithSpan",
			slog.SinkEntry{
				Time:    time.Unix(0, 0),
				Level:   slog.LevelInfo,
				Message: "msg",
				Fields:  slog.M(slog.F("k", "v")),
			},
		},
		{
			"WithValidSpan",
			slog.SinkEntry{
				Time:    time.Unix(0, 0),
				Level:   slog.LevelInfo,
				Message: "msg",
				Fields:  slog.M(slog.F("k", "v")),
			},
		},
		{
			"WithNames",
			slog.SinkEntry{
				Time:        time.Unix(0, 0),
				Level:       slog.LevelInfo,
				Message:     "msg",
				LoggerNames: []string{"svc", "sub", "component"},
				Fields:      slog.M(slog.F("k", "v")),
			},
		},
		{
			"SingleLine",
			slog.SinkEntry{
				Time:    time.Unix(0, 0),
				Level:   slog.LevelInfo,
				Message: "hello world",
				Fields: slog.M(
					slog.F("k1", "v1"),
					slog.F("k2", 123),
				),
			},
		},
		{
			"MultilineMsg",
			slog.SinkEntry{
				Time:    time.Unix(0, 0),
				Level:   slog.LevelInfo,
				Message: "line1\nline2\nline3\nline4\nline5",
			},
		},
		{
			"MultilineMultifieldSmall",
			slog.SinkEntry{
				Time:  time.Unix(0, 0),
				Level: slog.LevelInfo,
				Fields: slog.M(
					slog.F("msg", "..."),
					slog.F("stack", genMultiline(10)),
				),
			},
		},
		{
			"MultilineMultifieldLarge",
			slog.SinkEntry{
				Time:  time.Unix(0, 0),
				Level: slog.LevelInfo,
				Fields: slog.M(
					slog.F("msg", "..."),
					slog.F("stack", genMultiline(1000)),
				),
			},
		},
	}
	for _, tc := range testcases {
		b.Run(tc.name, func(b *testing.B) {
			w := io.Discard
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				entryhuman.Fmt(bytes.NewBuffer(nil), w, tc.entry)
			}
		})
	}
}
