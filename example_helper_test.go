package slog_test

import (
	"context"
	"net/http"
	"os"

	"cdr.dev/slog"
	"cdr.dev/slog/sloggers/sloghuman"
)

func httpLogHelper(ctx context.Context, status int) {
	slog.Helper()

	l.Info(ctx, "sending HTTP response",
		slog.F("status", status),
	)
}

var l = slog.Make(sloghuman.Sink(os.Stdout))

func ExampleHelper() {
	ctx := context.Background()
	httpLogHelper(ctx, http.StatusBadGateway)

	// 2019-12-07 21:18:42.567 [INFO]	<example_helper_test.go:24>	sending HTTP response	{"status": 502}
}
