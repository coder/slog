package slog_test

import (
	"bytes"
	"encoding/json"
	"io"
	"runtime"
	"strings"
	"testing"
	"time"

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
		assert.Equal(t, "JSON", exp, act)
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
			slog.F("meow", complexJSON(complex(10, 10))),
		), `{
			"meow": {
				"error": [
					{
						"msg": "failed to marshal to JSON",
						"fun": "cdr.dev/slog.encodeJSON",
						"loc": "`+mapTestFile+`:131"
					},
					"json: error calling MarshalJSON for type slog_test.complexJSON: json: unsupported type: complex128"
				],
				"type": "slog_test.complexJSON",
				"value": "(10+10i)"
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

	t.Run("array", func(t *testing.T) {
		t.Parallel()

		test(t, slog.M(
			slog.F("meow", [3]string{
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

	t.Run("nilSlice", func(t *testing.T) {
		t.Parallel()

		test(t, slog.M(
			slog.F("slice", []string(nil)),
		), `{
			"slice": null
		}`)
	})

	t.Run("nil", func(t *testing.T) {
		t.Parallel()

		test(t, slog.M(
			slog.F("val", nil),
		), `{
			"val": null
		}`)
	})

	t.Run("json.Marshaler", func(t *testing.T) {
		t.Parallel()

		test(t, slog.M(
			slog.F("val", time.Date(2000, 02, 05, 4, 4, 4, 0, time.UTC)),
		), `{
			"val": "2000-02-05T04:04:04Z"
		}`)
	})

	t.Run("complex", func(t *testing.T) {
		t.Parallel()

		test(t, slog.M(
			slog.F("val", complex(10, 10)),
		), `{
			"val": "(10+10i)"
		}`)
	})

	t.Run("privateStruct", func(t *testing.T) {
		t.Parallel()

		test(t, slog.M(
			slog.F("val", struct {
				meow string
				bar  int
				far  uint
			}{
				meow: "hi",
				bar:  23,
				far:  600,
			}),
		), `{
			"val": "{meow:hi bar:23 far:600}"
		}`)
	})
}

type meow struct {
	a int
}

func indentJSON(t *testing.T, j string) string {
	b := &bytes.Buffer{}
	err := json.Indent(b, []byte(j), "", strings.Repeat(" ", 4))
	assert.Success(t, "indent JSON", err)

	return b.String()
}

func marshalJSON(t *testing.T, m slog.Map) string {
	actb, err := json.Marshal(m)
	assert.Success(t, "marshal map to JSON", err)
	return indentJSON(t, string(actb))
}

type complexJSON complex128

func (c complexJSON) MarshalJSON() ([]byte, error) {
	return json.Marshal(complex128(c))
}
