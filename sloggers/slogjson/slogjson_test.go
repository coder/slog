package slogjson_test

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"runtime"
	"testing"

	"go.opencensus.io/trace"

	"cdr.dev/slog"
	"cdr.dev/slog/internal/assert"
	"cdr.dev/slog/sloggers/slogjson"
)

var _, slogjsonTestFile, _, _ = runtime.Caller(0)

var bg = context.Background()

func TestMake(t *testing.T) {
	t.Parallel()

	ctx, s := trace.StartSpan(bg, "meow")
	_ = s
	b := &bytes.Buffer{}
	l := slogjson.Make(b)
	l.Error(ctx, "line1\n\nline2", slog.F("wowow", "me\nyou"))

	j := filterJSONTimestamp(b.String())
	exp := fmt.Sprintf(`{"level":"ERROR","component":"","msg":"line1\n\nline2","caller":"`+slogjsonTestFile+`:29","func":"cdr.dev/slog/sloggers/slogjson_test.TestMake","trace":"%v","span":"%v","fields":{"wowow":"me\nyou"}}
`, s.SpanContext().TraceID, s.SpanContext().SpanID)
	assert.Equal(t, exp, j, "entry")
}

func filterJSONTimestamp(j string) string {
	return regexp.MustCompile(`"ts":[^,]+,`).ReplaceAllString(j, "")
}
