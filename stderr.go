package stderrlog

import (
	"context"
	"io"
	"os"
	"strings"
	"sync"

	"go.coder.com/m/lib/log"
	"go.coder.com/m/lib/log/internal/core"
)

var stderrLog = makeLogger(os.Stderr)

// Debug logs the given msg and fields on the debug level to stderr.
func Debug(ctx context.Context, msg string, fields log.F) {
	Make().Debug(core.WithSkip(ctx, 1), msg, fields)
}

// Info logs the given msg and fields on the info level to stderr.
func Info(ctx context.Context, msg string, fields log.F) {
	Make().Info(core.WithSkip(ctx, 1), msg, fields)
}

// Warn logs the given msg and fields on the warn level to stderr.
func Warn(ctx context.Context, msg string, fields log.F) {
	Make().Warn(core.WithSkip(ctx, 1), msg, fields)
}

// Error logs the given msg and fields on the error level to stderr.
func Error(ctx context.Context, msg string, fields log.F) {
	Make().Error(core.WithSkip(ctx, 1), msg, fields)
}

// Critical logs the given msg and fields on the critical level to stderr.
func Critical(ctx context.Context, msg string, fields log.F) {
	Make().Critical(core.WithSkip(ctx, 1), msg, fields)
}

// Fatal logs the given msg and fields on the fatal level to stderr.
func Fatal(ctx context.Context, msg string, fields log.F) {
	Make().Fatal(core.WithSkip(ctx, 1), msg, fields)
}

// Inspect is useful for one off debug statements.
func Inspect(v ...interface{}) {
	ctx := context.Background()
	ctx = core.WithSkip(ctx, 1)
	core.Inspect(context.Background(), stderrLog, v)
}

// Stderr creates a logger that writes logs to os.Stderr.
func Make() log.Logger {
	return stderrLog
}

func makeLogger(w io.Writer) log.Logger {
	return stderrLogger{
		&sync.Mutex{},
		w,
		core.Logger{},
	}
}

type stderrLogger struct {
	// This is specifically a pointer because we use the logger as a value
	// to make the With functions simpler as the outer logger is shallow
	// cloned automatically.
	// (nhooyr): I wonder if this is really necessary.
	mu *sync.Mutex
	w  io.Writer
	l  core.Logger
}

func (sl stderrLogger) Debug(ctx context.Context, msg string, fields log.F) {
	sl.log(ctx, core.Debug, msg, fields)
}

func (sl stderrLogger) Info(ctx context.Context, msg string, fields log.F) {
	sl.log(ctx, core.Info, msg, fields)
}

func (sl stderrLogger) Warn(ctx context.Context, msg string, fields log.F) {
	sl.log(ctx, core.Warn, msg, fields)
}

func (sl stderrLogger) Error(ctx context.Context, msg string, fields log.F) {
	sl.log(ctx, core.Error, msg, fields)
}

func (sl stderrLogger) Critical(ctx context.Context, msg string, fields log.F) {
	sl.log(ctx, core.Critical, msg, fields)
}

func (sl stderrLogger) Fatal(ctx context.Context, msg string, fields log.F) {
	sl.log(ctx, core.Fatal, msg, fields)
	os.Exit(1)
}

func (sl stderrLogger) With(fields log.F) log.Logger {
	sl.l = sl.l.With(fields)
	return sl
}

func (sl stderrLogger) log(ctx context.Context, sev core.Severity, msg string, fields log.F) {
	ent := sl.l.Make(ctx, core.EntryConfig{
		Sev:    sev,
		Msg:    msg,
		Fields: fields,
		Skip:   2,
	})

	sl.write(ent)

	if ent.Sev == core.Fatal {
		os.Exit(1)
	}
}

func (sl stderrLogger) write(ent core.Entry) {
	s := ent.String()
	lines := strings.Split(s, "\n")

	fieldsLines := lines[1:]
	for i, line := range fieldsLines {
		if line == "" {
			continue
		}
		fieldsLines[i] = "\t" + line
	}

	s = strings.Join(lines, "\n")

	sl.writeString(s + "\n")
}

func (sl stderrLogger) writeString(s string) {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	io.WriteString(sl.w, s)
}
