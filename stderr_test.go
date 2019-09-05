package stderrlog

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"

	"go.coder.com/m/lib/log"
	"go.coder.com/m/lib/log/internal/core"
)

type captureWriter struct {
	entries []string
}

func (ew *captureWriter) Write(p []byte) (n int, err error) {
	ew.entries = append(ew.entries, string(p))
	return len(p), nil
}

func TestStderr(t *testing.T) {
	t.Parallel()

	var ew captureWriter
	l := makeLogger(&ew)

	ctx := context.Background()

	l.Error(ctx, "my msg", log.F{
		"field": "value",
	})

	log.Stdlib(ctx, l).Print("mama mia")

	for i, ent := range ew.entries {
		stripped, err := core.StripEntryTimestamp(ent)
		if err != nil {
			t.Fatalf("failed to strip entry timestamp: %+v\nentry:\n%s", err, ent)
		}
		ew.entries[i] = stripped
	}
	expEntries := []string{
		`stderr_test.go:31: [ERROR]: my msg
	field: value
`, `stderr_test.go:35: [INFO] (stdlog): mama mia
`,
	}
	if !reflect.DeepEqual(ew.entries, expEntries) {
		t.Fatalf("expected entries to be the same: %v", cmp.Diff(expEntries, ew.entries))
	}
}
