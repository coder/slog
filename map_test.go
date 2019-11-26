package slog

import (
	"encoding/json"
	"strings"
	"testing"

	"go.coder.com/slog/internal/assert"
)

func TestMapJSON(t *testing.T) {
	m := Map{
		{"wow", Map{
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

func Test_snakecase(t *testing.T) {
	t.Parallel()

	t.Run("table", func(t *testing.T) {
		t.Parallel()

		tcs := []struct {
			s   string
			exp string
		}{
			{
				"meowBar",
				"meow_bar",
			},
			{
				"MeowBar",
				"meow_bar",
			},
			{
				"MEOWBar",
				"meow_bar",
			},
			{
				"Meow123BAR",
				"meow_123_bar",
			},
			{
				"BöseÜberraschung",
				"böse_überraschung",
			},
			{
				"GL11Version",
				"gl_11_version",
			},
			{
				"SimpleXMLParser",
				"simple_xml_parser",
			},
			{
				"PDFLoader",
				"pdf_loader",
			},
			{
				"HTML",
				"html",
			},
		}

		for i, tc := range tcs {
			tc := tc
			i := i
			t.Run("", func(t *testing.T) {
				t.Parallel()

				out := snakecase(tc.s)
				assert.Equalf(t, tc.exp, out, "snakecase gave unexpected output for case %d", i)
			})
		}
	})
}
