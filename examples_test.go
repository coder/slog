package slog_test

import (
	"context"
	"encoding/hex"
	"io"
	"math/rand"
	"os"
	"testing"
	"time"

	"golang.org/x/xerrors"

	"go.coder.com/slog"
	"go.coder.com/slog/sloggers/slogjson"
	"go.coder.com/slog/sloggers/slogtest"
)

func Example_test() {
	// Nil here but would be provided by the testing framework.
	var t testing.TB

	slogtest.Info(t, "my message here",
		slog.F("field_name", "something or the other"),
		slog.F("some_map", slog.Map(
			slog.F("nested_fields", "wowow"),
		)),
		slog.Error(
			xerrors.Errorf("wrap1: %w",
				xerrors.Errorf("wrap2: %w",
					io.EOF),
			)),
		slog.Component("test"),
	)

	// --- PASS: TestExample (0.00s)
	//    examples_test.go:49: [INFO] {slogtest.go:101} (test) Sep 09 16:17:46.925: my message here
	//        field_name: something or the other
	//        some_map:
	//          nested_fields: wowow
	//        error:
	//          - wrap1
	//            go.coder.com/slog_test.TestExample
	//              /Users/nhooyr/src/cdr/slog/examples_test.go:55
	//          - wrap2
	//            go.coder.com/slog_test.TestExample
	//              /Users/nhooyr/src/cdr/slog/examples_test.go:56
	//          - EOF
}

func TestExample(t *testing.T) {
	err := slog.Error(
		xerrors.Errorf(randStr()+": %w",
			xerrors.Errorf(randStr()+": %w",
				io.EOF),
		))
	for i := 0; i < 100; i++ {
		m := slog.Map(
			slog.F(randStr(), "something or the other"),
			slog.F("some_map", slog.Map(
				slog.F("nested_fields", "wowow"),
			)),
			slog.Component(randStr()),
			slog.F("hi", 3),
			slog.F("bool", true),
			slog.F("str", []string{randStr()}),
		)
		if rand.Intn(4)%4 == 0 {
			m = append(m, err)
		}
		slogjson.Make(os.Stderr).Info(context.Background(), randStr(), m...)
	}
}

func TestJSON(t *testing.T) {
	l := slogjson.Make(os.Stderr)
	l.Info(context.Background(), "my message\r here",
		slog.F("field_name", "something or the other"),
		slog.F("some_map", slog.Map(
			slog.F("nested_fields", "wowow"),
		)),
		slog.Error(
			xerrors.Errorf("wrap1: %w",
				xerrors.Errorf("wrap2: %w",
					io.EOF),
			)),
		slog.Component("test"),
	)
}

func randStr() string {
	p := make([]byte, rand.Intn(5)+3)
	rand.Read(p)
	return hex.EncodeToString(p)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
