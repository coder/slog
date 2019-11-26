package slog_test

import (
	"encoding/json"
	"strings"
	"testing"

	"go.coder.com/slog"
	"go.coder.com/slog/internal/assert"
)

func TestMapJSON(t *testing.T) {
	m := slog.Map{
		{"wow", slog.Map{
			{"nested", true},
			{"much", 3},
			{"list", []string{
				"3",
				"5",
			}},
		}},
	}

	act, err := json.MarshalIndent(m, "", strings.Repeat(" ", 2))
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
