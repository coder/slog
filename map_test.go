package slog_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"runtime"
	"strings"
	"testing"

	"golang.org/x/xerrors"

	"cdr.dev/slog"
	"cdr.dev/slog/internal/assert"
)

var _, mapTestFile, _, _ = runtime.Caller(0)

func TestMap(t *testing.T) {
	t.Parallel()

	test := func(t *testing.T, m slog.Map, exp string) {
		t.Helper()
		exp = indentJSON(t, exp)
		act := marshalJSON(t, m)
		assert.Equal(t, exp, act, "JSON")
	}

	t.Run("JSON", func(t *testing.T) {
		t.Parallel()

		type Meow struct {
			Wow       string `json:"meow"`
			Something int    `json:",omitempty"`
			Ignored   bool   `json:"-"`
		}

		test(t, slog.M(
			slog.Error(
				xerrors.Errorf("wrap1: %w",
					xerrors.Errorf("wrap2: %w",
						io.EOF,
					),
				),
			),
			slog.F("meow", struct {
				Izi string `json:"izi"`
				M   *Meow  `json:"Amazing"`

				ignored bool
			}{
				Izi: "sogood",
				M: &Meow{
					Wow:       "",
					Something: 0,
					Ignored:   true,
				},
			}),
		), `{
			"error": [
				{
					"msg": "wrap1",
					"fun": "cdr.dev/slog_test.TestMap.func2",
					"loc": "`+mapTestFile+`:41" 
				},
				{
					"msg": "wrap2",
					"fun": "cdr.dev/slog_test.TestMap.func2",
					"loc": "`+mapTestFile+`:42" 
				},
				"EOF"
			],
			"meow": {
				"izi": "sogood",
				"Amazing": {
					"meow": ""
				}
			}
		}`)
	})

	t.Run("badJSON", func(t *testing.T) {
		t.Parallel()

		mapTestFile := strings.Replace(mapTestFile, "_test", "", 1)

		test(t, slog.M(
			slog.F("meow", indentJSON),
		), `{
			"meow": {
				"error": [
					{
						"msg": "failed to marshal to JSON",
						"fun": "cdr.dev/slog.encode",
						"loc": "`+mapTestFile+`:105"
					},
					"json: unsupported type: func(*testing.T, string) string"
				],
				"type": "func(*testing.T, string) string",
				"value": "`+fmt.Sprint(interface{}(indentJSON))+`"
			}
		}`)
	})

	t.Run("basic", func(t *testing.T) {
		t.Parallel()

		test(t, slog.M(
			slog.F("wow", slog.M(
				slog.F("nested", true),
				slog.F("much", 3),
				slog.F("list", []string{
					"3",
					"5",
				}),
			)),
		), `{
			"wow": {
				"nested": true,
				"much": 3,
				"list": [
					"3",
					"5"
				]
			}
		}`)
	})

	t.Run("slice", func(t *testing.T) {
		t.Parallel()

		test(t, slog.M(
			slog.F("meow", []string{
				"1",
				"2",
				"3",
			}),
		), `{
			"meow": [
				"1",
				"2",
				"3"
			]
		}`)
	})

	t.Run("forceJSON", func(t *testing.T) {
		t.Parallel()

		test(t, slog.M(
			slog.F("error", slog.ForceJSON(io.EOF)),
		), `{
			"error": {}
		}`)
	})

	t.Run("value", func(t *testing.T) {
		t.Parallel()

		test(t, slog.M(
			slog.F("error", meow{1}),
		), `{
			"error": "xdxd"
		}`)
	})
}

type meow struct {
	a int
}

func (m meow) SlogValue() interface{} {
	return "xdxd"
}

func indentJSON(t *testing.T, j string) string {
	b := &bytes.Buffer{}
	err := json.Indent(b, []byte(j), "", strings.Repeat(" ", 4))
	assert.Success(t, err, "indent JSON")

	return b.String()
}

func marshalJSON(t *testing.T, m slog.Map) string {
	actb, err := json.Marshal(m)
	assert.Success(t, err, "marshal map to JSON")
	return indentJSON(t, string(actb))
}
