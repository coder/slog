package slogstackdriver

import (
	logpbtype "google.golang.org/genproto/googleapis/logging/type"

	"cdr.dev/slog/v2"
)

func Sev(level slog.Level) logpbtype.LogSeverity {
	return sev(level)
}
