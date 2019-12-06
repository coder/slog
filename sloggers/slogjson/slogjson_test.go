package slogjson_test

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"testing"

	"go.opencensus.io/trace"

	"cdr.dev/slog"
	"cdr.dev/slog/internal/assert"
	"cdr.dev/slog/sloggers/slogjson"
)

var bg = context.Background()

func TestMake(t *testing.T) {
	t.Parallel()

	ctx, s := trace.StartSpan(bg, "meow")
	_ = s
	b := &bytes.Buffer{}
	l := slogjson.Make(b)
	l.Error(ctx, "line1\n\nline2", slog.F("wowow", "me\nyou"))

	j := filterJSONTimestamp(b.String())
	exp := fmt.Sprintf(`{"level":"ERROR","component":"","msg":"line1\n\nline2","caller":"/Users/nhooyr/src/cdr/slog/sloggers/slogjson/slogjson_test.go:26","func":"cdr.dev/slog/sloggers/slogjson_test.TestMake","trace":"%v","span":"%v","fields":{"wowow":"me\nyou"}}
`, s.SpanContext().TraceID, s.SpanContext().SpanID)
	assert.Equal(t, exp, j, "entry")
}

func filterJSONTimestamp(j string) string {
	return regexp.MustCompile(`"ts":[^,]+,`).ReplaceAllString(j, "")
}
