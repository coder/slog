package testlog_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"go.coder.com/m/lib/log"
	"go.coder.com/m/lib/log/internal/core"
	"go.coder.com/m/lib/log/testlog"
)

type tbMock struct {
	log   string
	error string
	fatal string
	testing.TB
}

func (tb *tbMock) Helper() {}

func (tb *tbMock) Log(v ...interface{}) {
	tb.log = fmt.Sprint(v...)
}

func (tb *tbMock) Error(v ...interface{}) {
	tb.error = fmt.Sprint(v...)
}

func (tb *tbMock) Fatal(v ...interface{}) {
	tb.fatal = fmt.Sprint(v...)
}

type tbMockCheckConfig struct {
	logged  bool
	errored bool
	fataled bool
}

func (tb *tbMock) assert(config tbMockCheckConfig) bool {
	return config.logged == (tb.log != "") &&
		config.errored == (tb.error != "") &&
		config.fataled == (tb.fatal != "")
}

func testOutput(t *testing.T) string {
	gotest := exec.Command(os.Args[0], "-test.run="+t.Name(), "-test.v", "-test.count=1")
	gotest.Env = []string{"TEST_OUTPUT=true"}

	outb, err := gotest.CombinedOutput()
	out := string(outb)
	if err != nil {
		failedCmd(t, gotest, err, out)
	}

	t.Logf("go test output:\n%v", out)

	lines := strings.Split(out, "\n")

	for i := 0; i < len(lines); i++ {
		l := lines[i]
		l = strings.TrimSpace(l)
		if strings.HasPrefix(l, "===") || strings.HasPrefix(l, "---") {
			lines = lines[i+1:]
			i--
		}
	}

	if len(lines) >= 2 {
		// Last two lines are like "\nPASS\n"
		lines = lines[:len(lines)-2]
	}

	indent := strings.IndexFunc(lines[0], func(r rune) bool {
		return r != ' '
	})
	var indentStr string
	if indent == -1 {
		indentStr = ""
	}
	indentStr = strings.Repeat(" ", indent)
	for i := range lines {
		lines[i] = strings.TrimPrefix(lines[i], indentStr)
	}

	out = strings.Join(lines, "\n")

	return out
}

func Test_testLogger(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("integration", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			name string
			fn   func(t *testing.T)
			exp  string
		}{
			{
				"fmt",
				func(t *testing.T) {
					testlog.Info(t, "msg", log.F{
						"field": "value",
					})
				},
				`test_test.go:110: [INFO]: msg
    field: value`,
			},
			{
				"stdlog",
				func(t *testing.T) {
					ctx := log.With(ctx, log.F{
						"field": "value",
					})
					log.Stdlib(ctx, testlog.Make(t)).Print("msg")
				},
				`stdlib_log_adapter.go:29: test_test.go:123: [INFO] (stdlog): msg
    field: value`,
			},
		}

		for _, tc := range testCases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				if os.Getenv("TEST_OUTPUT") == "true" {
					tc.fn(t)
					return
				}

				// This runs a test using `go test` instead of mocking out testing.TB because we want to ensure
				// the testing framework sees the location properly.
				out := testOutput(t)
				out, err := core.StripEntryTimestamp(out)
				if err != nil {
					t.Fatalf("failed to strip timestamp from output: %v", err)
				}

				if out != tc.exp {
					t.Fatalf("expected entry to be:\n%q\nbut got:\n%q", tc.exp, out)
				}
			})
		}
	})

	t.Run("failOnError", func(t *testing.T) {
		t.Parallel()

		tb := &tbMock{}
		l := testlog.Make(tb)

		l.Error(ctx, "msg", log.F{})
		if !tb.assert(tbMockCheckConfig{
			errored: true,
		}) {
			t.Fatalf("expected l.Error to cause testing error: %#v", tb)
		}

		l.Fatal(ctx, "msg", log.F{})
		if !tb.assert(tbMockCheckConfig{
			errored: true,
			fataled: true,
		}) {
			t.Fatalf("expected l.Fatal to cause testing fatal: %#v", tb)
		}
	})

	t.Run("ignoreError", func(t *testing.T) {
		t.Parallel()

		tb := &tbMock{}
		l := testlog.Make(tb, testlog.IgnoreError())

		l.Error(ctx, "msg", log.F{})
		if !tb.assert(tbMockCheckConfig{
			logged: true,
		}) {
			t.Fatalf("expected l.Error to cause testing log: %#v", tb)
		}
	})

	t.Run("fatalWithIgnoreError", func(t *testing.T) {
		t.Parallel()

		tb := &tbMock{}

		defer func() {
			if recover() == nil {
				t.Fatal("expected panic from l.Fatal with log.IgnoreError option")
			}
		}()
		l := testlog.Make(tb, testlog.IgnoreError())
		l.Fatal(ctx, "msg", log.F{})
	})
}

func failedCmd(t *testing.T, cmd *exec.Cmd, err error, out string) {
	lines := strings.Split(out, "\n")
	for i, line := range lines {
		lines[i] = "\t" + line
	}
	out = strings.Join(lines, "\n")

	t.Fatalf("failed to run %q: %v\nout:\n%v", cmd.Args, err, out)
}
