package slog

import (
	"context"
	"testing"

	"cdr.dev/slog/internal/assert"
)

func TestExit(t *testing.T) {
	// This can't be parallel since it modifies a global variable.
	t.Run("defaultExitFn", func(t *testing.T) {
		var (
			ctx                 = context.Background()
			log                 Logger
			defaultExitFnCalled bool
		)

		prevExitFn := defaultExitFn
		t.Cleanup(func() { defaultExitFn = prevExitFn })

		defaultExitFn = func(_ int) {
			defaultExitFnCalled = true
		}

		log.Debug(ctx, "hi")
		log.Info(ctx, "hi")
		log.Warn(ctx, "hi")
		log.Error(ctx, "hi")
		log.Critical(ctx, "hi")
		log.Fatal(ctx, "hi")

		assert.True(t, "default exit fn used", defaultExitFnCalled)
	})
}
