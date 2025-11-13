package entryhuman_test

import (
	"bytes"
	"database/sql"
	"flag"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"cdr.dev/slog"
	"cdr.dev/slog/internal/assert"
	"cdr.dev/slog/internal/entryhuman"
)

var kt = time.Date(2000, time.February, 5, 4, 4, 4, 4, time.UTC)

var updateGoldenFiles = flag.Bool("update-golden-files", false, "update golden files in testdata")

type testObj struct {
	foo int
	bar int
	dra []byte
}

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
				LoggerNames: []string{"some", "cat"},
				Message:     "meow",
				Fields: slog.M(
					slog.F("breath", "stinky"),
				),
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
		{
			"spacey",
			slog.SinkEntry{
				Level: slog.LevelWarn,
				Fields: slog.M(
					slog.F("space in my key", "value in my value"),
				),
			},
		},
		{
			"nil",
			slog.SinkEntry{
				Level: slog.LevelWarn,
				Fields: slog.M(
					slog.F("nan", nil),
				),
			},
		},
		{
			"bytes",
			slog.SinkEntry{
				Level: slog.LevelWarn,
				Fields: slog.M(
					slog.F("somefile", []byte("blah bla\x01h blah")),
				),
			},
		},
		{
			"driverValue",
			slog.SinkEntry{
				Level: slog.LevelWarn,
				Fields: slog.M(
					slog.F("val", sql.NullString{String: "dog", Valid: true}),
					slog.F("inval", sql.NullString{String: "cat", Valid: false}),
				),
			},
		},
		{
			"object",
			slog.SinkEntry{
				Level: slog.LevelWarn,
				Fields: slog.M(
					slog.F("obj", slog.M(
						slog.F("obj1", testObj{
							foo: 1,
							bar: 2,
							dra: []byte("blah"),
						}),
						slog.F("obj2", testObj{
							foo: 3,
							bar: 4,
							dra: []byte("blah"),
						}),
					)),
					slog.F("map", map[string]string{
						"key1": "value1",
					}),
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
			entryhuman.Fmt(&gotBuf, io.Discard, tc.ent)

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

	t.Run("isTTY during file close", func(t *testing.T) {
		t.Parallel()

		tmpdir := t.TempDir()
		f, err := os.CreateTemp(tmpdir, "slog")
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()

		done := make(chan struct{}, 2)
		go func() {
			entryhuman.Fmt(new(bytes.Buffer), f, slog.SinkEntry{
				Level: slog.LevelCritical,
				Fields: slog.M(
					slog.F("hey", "hi"),
				),
			})
			done <- struct{}{}
		}()
		go func() {
			_ = f.Close()
			done <- struct{}{}
		}()
		<-done
		<-done
	})
}

// Verifies that OptimizedFmt returtns the same result as Fmt.
// We can remove this if we are okay with removing the existing Fmt function and replacing it with OptimizedFmt (can return an error).
func TestEntry_Optimized(t *testing.T) {
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
				LoggerNames: []string{"some", "cat"},
				Message:     "meow",
				Fields: slog.M(
					slog.F("breath", "stinky"),
				),
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
		{
			"spacey",
			slog.SinkEntry{
				Level: slog.LevelWarn,
				Fields: slog.M(
					slog.F("space in my key", "value in my value"),
				),
			},
		},
		{
			"nil",
			slog.SinkEntry{
				Level: slog.LevelWarn,
				Fields: slog.M(
					slog.F("nan", nil),
				),
			},
		},
		{
			"bytes",
			slog.SinkEntry{
				Level: slog.LevelWarn,
				Fields: slog.M(
					slog.F("somefile", []byte("blah bla\x01h blah")),
				),
			},
		},
		{
			"driverValue",
			slog.SinkEntry{
				Level: slog.LevelWarn,
				Fields: slog.M(
					slog.F("val", sql.NullString{String: "dog", Valid: true}),
					slog.F("inval", sql.NullString{String: "cat", Valid: false}),
				),
			},
		},
		{
			"object",
			slog.SinkEntry{
				Level: slog.LevelWarn,
				Fields: slog.M(
					slog.F("obj", slog.M(
						slog.F("obj1", testObj{
							foo: 1,
							bar: 2,
							dra: []byte("blah"),
						}),
						slog.F("obj2", testObj{
							foo: 3,
							bar: 4,
							dra: []byte("blah"),
						}),
					)),
					slog.F("map", map[string]string{
						"key1": "value1",
					}),
				),
			},
		},
		{
			"primitiveTypes",
			slog.SinkEntry{
				Level:   slog.LevelInfo,
				Message: "primitives",
				Time:    kt,
				Fields: slog.M(
					slog.F("bool_true", true),
					slog.F("bool_false", false),
					slog.F("int", 42),
					slog.F("int8", int8(-8)),
					slog.F("int16", int16(-16)),
					slog.F("int32", int32(-32)),
					slog.F("int64", int64(-64)),
					slog.F("uint", uint(42)),
					slog.F("uint8", uint8(8)),
					slog.F("uint16", uint16(16)),
					slog.F("uint32", uint32(32)),
					slog.F("uint64", uint64(64)),
					slog.F("float32", float32(3.14)),
					slog.F("float64", 2.71828),
				),
			},
		},
		{
			"primitiveEdgeCases",
			slog.SinkEntry{
				Level:   slog.LevelWarn,
				Message: "edge cases",
				Time:    kt,
				Fields: slog.M(
					slog.F("zero_int", 0),
					slog.F("neg_int", -999),
					slog.F("max_int64", int64(9223372036854775807)),
					slog.F("min_int64", int64(-9223372036854775808)),
					slog.F("max_uint64", uint64(18446744073709551615)),
					slog.F("zero_float", 0.0),
					slog.F("neg_float", -123.456),
				),
			},
		},
		{
			"mixedPrimitiveAndComplex",
			slog.SinkEntry{
				Level:   slog.LevelDebug,
				Message: "mixed types",
				Time:    kt,
				Fields: slog.M(
					slog.F("count", 100),
					slog.F("name", "test"),
					slog.F("enabled", true),
					slog.F("ratio", 0.95),
					slog.F("data", []byte("bytes")),
					slog.F("nil_val", nil),
				),
			},
		},
		{
			"allLogLevels",
			slog.SinkEntry{
				Level:   slog.LevelCritical, // Test Critical
				Message: "critical",
				Time:    kt,
			},
		},
		{
			"fatalLevel",
			slog.SinkEntry{
				Level:   slog.LevelFatal, // Test Fatal
				Message: "fatal",
				Time:    kt,
			},
		},
	}

	for _, tc := range ents {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var fmtBuf bytes.Buffer
			var optBuf bytes.Buffer

			entryhuman.Fmt(&fmtBuf, io.Discard, tc.ent)
			entryhuman.OptimizedFmt(&optBuf, io.Discard, tc.ent)

			assert.Equal(t, "outputs match", fmtBuf.String(), optBuf.String())
		})
	}

	t.Run("isTTY during file close", func(t *testing.T) {
		t.Parallel()

		tmpdir := t.TempDir()
		f, err := os.CreateTemp(tmpdir, "slog")
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()

		done := make(chan struct{}, 2)
		go func() {
			entryhuman.OptimizedFmt(new(bytes.Buffer), f, slog.SinkEntry{
				Level: slog.LevelCritical,
				Fields: slog.M(
					slog.F("hey", "hi"),
				),
			})
			done <- struct{}{}
		}()
		go func() {
			_ = f.Close()
			done <- struct{}{}
		}()
		<-done
		<-done
	})
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
					entryhuman.OptimizedFmt(bytes.NewBuffer(nil), w, se)
				}
			})
		}
	}
	bench(b, true)
	bench(b, false)
}
