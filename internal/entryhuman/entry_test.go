package entryhuman_test

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"cdr.dev/slog"
	"cdr.dev/slog/internal/assert"
	"cdr.dev/slog/internal/entryhuman"
)

var kt = time.Date(2000, time.February, 5, 4, 4, 4, 4, time.UTC)

var updateGoldenFiles = flag.Bool("update-golden-files", false, "update golden files in testdata")

func TestEntry(t *testing.T) {
	t.Parallel()

	type tcase struct {
		name string
		ent  slog.SinkEntry
	}

	ents := []tcase{
		{
			"simpleNoFields",
			slog.SinkEntry{
				Message: "wowowow\tizi",
				Time:    kt,
				Level:   slog.LevelDebug,

				File: "myfile",
				Line: 100,
				Func: "mypkg.ignored",
			},
		},
		{
			"multilineMessage",
			slog.SinkEntry{
				Message: "line1\nline2",
				Level:   slog.LevelInfo,
			},
		},
		{
			"multilineField",
			slog.SinkEntry{
				Message: "msg",
				Level:   slog.LevelInfo,
				Fields:  slog.M(slog.F("field", "line1\nline2")),
			},
		},
		{
			"named",
			slog.SinkEntry{
				Level:       slog.LevelWarn,
				LoggerNames: []string{"named", "meow"},
			},
		},
		{
			"funky",
			slog.SinkEntry{
				Level: slog.LevelWarn,
				Fields: slog.M(
					slog.F("funky^%&^&^key", "value"),
					slog.F("funky^%&^&^key2", "@#\t \t \n"),
				),
			},
		},
	}
	if *updateGoldenFiles {
		ents, err := os.ReadDir("testdata")
		if err != nil {
			t.Fatal(err)
		}
		for _, ent := range ents {
			os.Remove("testdata/" + ent.Name())
		}
	}

	for _, tc := range ents {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			goldenPath := fmt.Sprintf("testdata/%s.golden", tc.name)

			var gotBuf bytes.Buffer
			entryhuman.Fmt(&gotBuf, ioutil.Discard, tc.ent)

			if *updateGoldenFiles {
				err := os.WriteFile(goldenPath, gotBuf.Bytes(), 0o644)
				if err != nil {
					t.Fatal(err)
				}
				return
			}

			wantByt, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, "entry matches", string(wantByt), gotBuf.String())
		})
	}
}

func BenchmarkFmt(b *testing.B) {
	bench := func(b *testing.B, color bool) {
		nfs := []int{1, 4, 16}
		for _, nf := range nfs {
			name := fmt.Sprintf("nf=%v", nf)
			if color {
				name = "Colored-" + name
			}
			b.Run(name, func(b *testing.B) {
				fs := make([]slog.Field, nf)
				for i := 0; i < nf; i++ {
					fs[i] = slog.F("key", "value")
				}
				se := slog.SinkEntry{
					Level: slog.LevelCritical,
					Fields: slog.M(
						fs...,
					),
				}
				w := io.Discard
				if color {
					w = entryhuman.ForceColorWriter
				}
				b.ResetTimer()
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					entryhuman.Fmt(io.Discard.(io.StringWriter), w, se)
				}
			})
		}
	}
	bench(b, true)
	bench(b, false)
}
