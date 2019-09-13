package slogval_test

import (
	"encoding/json"
	"strings"
	"testing"

	"go.coder.com/slog"
	"go.coder.com/slog/internal/assert"
)

func TestMapJSON(t *testing.T) {
	m := slog.Map(
		slog.F("wow", slog.Map(
			slog.F("nested", true),
			slog.F("much", 3),
			slog.F("list", []string{
				"3",
				"5",
			}),
		)),
	)

	act, err := json.MarshalIndent(slog.Encode(m), "", strings.Repeat(" ", 2))
	if err != nil {
		t.Fatalf("failed to encode map to JSON: %+v", err)
	}

	exp := strings.TrimSpace(`
{
  "wow": {
    "nested": true,
    "much": 3,
    "list": [
      "3",
      "5"
    ]
  }
}
`)

	assert.Equalf(t, exp, string(act), "unexpected JSON")
}
