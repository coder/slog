package core

import (
	"context"
	"errors"
	"fmt"
	"io"
	"runtime/debug"
	"testing"
	"time"

	structpb "github.com/golang/protobuf/ptypes/struct"
	"go.opencensus.io/trace"
	"golang.org/x/xerrors"

	"go.coder.com/m/lib/log"
	"go.coder.com/m/lib/multierr"
)

func TestLogger(t *testing.T) {
	t.Parallel()

	t.Run("loc", func(t *testing.T) {
		t.Parallel()

		var l Logger
		ctx := context.Background()

		var ent Entry
		func() {
			ent = l.Make(ctx, EntryConfig{
				Sev:  Info,
				Msg:  "my msg",
				Skip: 1,
			})
		}()

		s := ent.String()
		defer func() {
			if t.Failed() {
				t.Logf("entry:\n%v", s)
			}
		}()
		s, err := StripEntryTimestamp(s)
		if err != nil {
			t.Fatalf("failed to strip timestamp from entry: %v", err)
		}
		const exp = `log_test.go:36: [INFO]: my msg`
		if exp != s {
			t.Fatalf("expected entry to be:\n%v", exp)
		}
	})

	t.Run("fields", func(t *testing.T) {
		t.Parallel()

		type testCase struct {
			name string
			// If nil defaults to context.Background.
			ctx context.Context
			// If nil defaults to zero value logger.
			log func() Logger
			ec  EntryConfig
			exp string
		}

		// This table driven set of tests shouldn't be table driven. Its too limiting.
		// A solid set of helpers would be a much better way to structure these.
		testCases := []testCase{
			{
				name: "error",
				ec: EntryConfig{
					Sev: Info,
					Msg: "my_msg",
					Fields: map[string]interface{}{
						"someerr": xerrors.New("woww"),
					},
				},
				exp: fmt.Sprintf(`[INFO]: my_msg
someerr:
  - err: woww
    fun: go.coder.com/m/lib/log/internal/core.TestLogger.func2
    loc: %s:76`, Location(0).File),
			},
			{
				name: "complexError",
				ec: EntryConfig{
					Sev: Info,
					Msg: "my_msg",
					Fields: map[string]interface{}{
						"someerr": func() error {
							err := errors.New("wowza")
							err = xerrors.Errorf("woww: %v", err)
							err = xerrors.Errorf("my wrap msg: %v", err)
							err = xerrors.Errorf("my wrap msg 2: %v", err)

							err = multierr.Combine(err, io.EOF)

							return err
						}(),
					},
				},
				exp: fmt.Sprintf(`[INFO]: my_msg
someerr:
  -
    - err: my wrap msg 2
      fun: go.coder.com/m/lib/log/internal/core.TestLogger.func2.1
      loc: %[1]v:95
    - err: my wrap msg
      fun: go.coder.com/m/lib/log/internal/core.TestLogger.func2.1
      loc: %[1]v:94
    - err: woww
      fun: go.coder.com/m/lib/log/internal/core.TestLogger.func2.1
      loc: %[1]v:93
    - wowza
  - EOF`, Location(0).File),
			},
			{
				name: "unexportedFields",
				ec: EntryConfig{
					Sev: Info,
					Msg: "my_msg",
					Fields: map[string]interface{}{
						"some struct": struct {
							unexported    string
							unexportedMap map[string]string
						}{
							"wow",
							map[string]string{
								"my val": "xd",
							},
						},
					},
				},
				exp: `[INFO]: my_msg
some_struct:
  unexported: wow
  unexported_map:
    my_val: xd`,
			},
			{
				name: "types",
				ec: EntryConfig{
					Sev: Info,
					Msg: "my_msg",
					Fields: map[string]interface{}{
						"float": 1.5,
						"int": func() *int {
							n := 5
							return &n
						}(),
						"bool":    true,
						"nil":     (*error)(nil),
						"time":    time.Date(2099, 0, 0, 0, 0, 0, 0, time.UTC),
						"complex": complex(1, 1),
					},
				},
				exp: `[INFO]: my_msg
bool: true
complex: (1+1i)
float: 1.5
int: 5
nil: nil
time: 2098-11-30 00:00:00 +0000 UTC`,
			},
			{
				name: "closureOverride",
				ec: EntryConfig{
					Sev: Info,
					Msg: "my_msg",
					Fields: map[string]interface{}{
						"my int": func() interface{} {
							return 3
						},
						// Value interface is tested by usage of multier.Error in the complexError test.
						// We can test here too later.
					},
				},
				exp: `[INFO]: my_msg
my_int: 3`,
			},
			{
				name: "proto",
				ec: EntryConfig{
					Sev: Info,
					Msg: "my_msg",
					Fields: map[string]interface{}{
						"struct": pbstruct(&structpb.Struct{
							Fields: map[string]*structpb.Value{
								"wow": pbfloat64(53),
							},
						}),
					},
				},
				exp: `[INFO]: my_msg
struct:
  kind:
    struct_value:
      fields:
        wow:
          kind:
            number_value: 53`,
			},
			{
				name: "quoteMsg",
				ec: EntryConfig{
					Sev: Info,
					Msg: "aamam a a am mam ama \n\n adsadsamkldmkas \t\t \t",
				},
				exp: `[INFO]: "aamam a a am mam ama \n\n adsadsamkldmkas \t\t \t"`,
			},
			{
				name: "componentAndID",
				ctx: With(context.Background(), map[string]interface{}{
					log.Component: "my component",
					log.ID:        "my_id",
				}),
				ec: EntryConfig{
					Sev: Info,
					Msg: "my message",
					Fields: map[string]interface{}{
						log.Component:   "component",
						"Another Field": "wowow",
					},
				},
				exp: `[INFO] (my component.component): my message
id: my_id
Another_Field: wowow`,
			},
			{
				name: "emptyComponent",
				ctx: With(context.Background(), map[string]interface{}{
					log.Component: "",
				}),
				ec: EntryConfig{
					Sev: Info,
					Msg: "my message",
					Fields: map[string]interface{}{
						log.Component: "component",
						log.ID:        "my_id",
					},
				},
				exp: `[INFO] (component): my message
id: my_id`,
			},
			{
				name: "with",
				ctx: With(context.Background(), map[string]interface{}{
					log.Component: "wowwow",
				}),
				log: func() Logger {
					var l Logger
					return l.With(map[string]interface{}{
						log.Component: "izi",
					})
				},
				ec: EntryConfig{
					Sev: Info,
					Msg: "my message",
					Fields: map[string]interface{}{
						log.Component: "22",
					},
				},
				exp: `[INFO] (izi.wowwow.22): my message`,
			},
			{
				name: "fatal",
				ec: EntryConfig{
					Sev: Fatal,
					Msg: "msg",
				},
				exp: `[FATAL]: msg`,
			},
			{
				name: "error",
				ec: EntryConfig{
					Sev: Error,
					Msg: "msg",
				},
				exp: `[ERROR]: msg`,
			},
			{
				name: "critical",
				ec: EntryConfig{
					Sev: Critical,
					Msg: "msg",
				},
				exp: `[CRITICAL]: msg`,
			},
			func() testCase {
				ctx := context.Background()
				ctx, s := trace.StartSpan(ctx, "foobar")
				s.End()
				return testCase{
					name: "spanAndTrace",
					ctx:  ctx,
					ec: EntryConfig{
						Sev: Debug,
						Msg: "yoyo",
					},
					exp: fmt.Sprintf(`[DEBUG]: yoyo
span: %v
trace: %v`, s.SpanContext().SpanID, s.SpanContext().TraceID),
				}
			}(),
			{
				name: "emptyStruct",
				ec: EntryConfig{
					Sev: Info,
					Msg: "my_msg",
					Fields: map[string]interface{}{
						"struct": map[string]struct{}{
							"meow": {},
							"bar":  {},
							"lol":  {},
							"xd":   {},
						},
					},
				},
				exp: `[INFO]: my_msg
struct:
  bar:
  lol:
  meow:
  xd:`,
			},
		}

		for _, tc := range testCases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				defer func() {
					r := recover()
					if r != nil {
						t.Errorf("panic: %v\n%s", r, debug.Stack())
					}
				}()

				var l Logger
				if tc.log != nil {
					l = tc.log()
				}

				ctx := context.Background()
				if tc.ctx != nil {
					ctx = tc.ctx
				}

				ent := l.Make(ctx, tc.ec)
				// We test this elsewhere to keep the above tests simpler.
				ent.Loc.File = ""

				s := ent.String()
				defer func() {
					if t.Failed() {
						t.Logf("entry:\n%v", s)
					}
				}()
				s, err := StripEntryTimestamp(s)
				if err != nil {
					t.Fatalf("failed to strip timestamp from entry: %v", err)
				}
				if tc.exp != s {
					t.Fatalf("expected entry to be:\n%v", tc.exp)
				}
			})
		}
	})
}
