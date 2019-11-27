# slog

[![GitHub Release](https://img.shields.io/github/v/release/cdr/slog?color=6b9ded&sort=semver)](https://github.com/cdr/slog/releases)
[![GoDoc](https://godoc.org/go.coder.com/slog?status.svg)](https://godoc.org/go.coder.com/slog)
[![Coveralls](https://img.shields.io/coveralls/github/cdr/slog?color=65d6a4)](https://coveralls.io/github/cdr/slog)
[![CI Status](https://github.com/cdr/slog/workflows/ci/badge.svg)](https://github.com/cdr/slog/actions)

slog is a minimal structured logging library for Go.

## Install

```bash
go get go.coder.com/slog
```

## Features

- Minimal API
- Tiny codebase
- First class [context.Context](https://blog.golang.org/context) support
- First class [testing.TB](https://godoc.org/go.coder.com/slog/slogtest) support
- Beautiful human readable logging output
  - Prints multiline fields and errors nicely
- Machine readable JSON output
- [GCP Stackdriver](https://godoc.org/go.coder.com/slog/sloggers/slogstackdriver) support
- [Tee](https://godoc.org/go.coder.com/slog#Tee) multiple loggers
- [Stdlib](https://godoc.org/go.coder.com/slog#Stdlib) log adapter
- Skip caller frames with [slog.Helper](https://godoc.org/go.coder.com/slog#Helper)
- Can encode any Go structure including private fields

## Example

```go
slogtest.Info(t, "my message here",
    slog.F{"field_name", "something or the other"},
    slog.F{"some_map", slog.Map{
        {"nested_fields", "wowow"},
    }},
    slog.Error(
        xerrors.Errorf("wrap1: %w",
            xerrors.Errorf("wrap2: %w",
                io.EOF,
            ),
        ),
    ),
)
```

![Example output screenshot](https://i.imgur.com/o8uW4Oy.png)

## Why?

The logging library of choice at [Coder](https://github.com/cdr) has been Uber's [zap](https://github.com/uber-go/zap)
for several years now.

It's a fantastic library for performance but the API and developer experience is not great.

First, the API surface is very large. See [godoc](https://godoc.org/go.uber.org/zap).
That's not including zap's subpackage [zapcore](https://godoc.org/go.uber.org/zap/zapcore) which
is itself very big.

The sprawling API has made it hard to understand, use and extend. zap's regular API is also very verbose
to the explicit typing. While it does offer a [sugared API](https://godoc.org/go.uber.org/zap#hdr-Choosing_a_Logger)
the API is in our opinion too dynamic. It's too easy to miss the key of a value since there is no static type
checking. It's less verbose but harder to read as there is less grouping each key value pair.

We wanted an API that only accepted the equivalent of [zap.Any](https://godoc.org/go.uber.org/zap#Any) for every field.
This is [slog.F](https://godoc.org/go.coder.com/slog#F).

Second, we found the human readable format to be hard to read due to the lack appropriate colors for different levels
and fields. `slog` colors distinct parts of each line to make it easier to scan logs. Even the JSON that represents
the fields in each log is syntax highlighted so that is very easy to scan. See the screenshot above.

Third, zap logged multiline fields and errors stack traces as JSON strings which made them pretty much unreadable in the
console. When using the human logger, slog automatically prints one multiline field after the log to make errors and
such much easier to read. slog also automatically prints a Go 1.13 error chain as an array. See screenshot above.

Fourth, zap does not support `context.Context`. We wanted to be able to pull up all relevant logs for a given trace,
user or request. With zap, we'd have to manually plug these fields in for every relevant log or use `With` on `zap.Logger`
to set the appropriate fields and pass it around. This got very verbose. `slog` lets you set fields in a `context.Context`
such that any log with the context prints those fields.

Fifth, we found it hard and confusing to extend zap. There are too many structures and configuration options. We wanted
a very simple and easy to understand extension model. With slog, there is only the idea of a Sink and Logger. Logger
is the concrete type used to provide the high level API. Sink is the interface that must be implemented for a new
logging backend.

Sixth, we found ourselves often implementing zap's [ObjectMarshaler](https://godoc.org/go.uber.org/zap/zapcore#ObjectMarshaler) to
log Go structures. This was very verbose and most of the time we just ended up only implementing `fmt.Stringer` and using `zap.Stringer`
instead. We wanted it to be automatic, even for private fields. slog handles this transparently for us. It will automatically
log Go structures with their fields separate, including private fields.

Seventh, t.Helper style APIf

Eight, tighter integration with `*testing.T`.

These are the main reasons we decided it was worth creating another log package for Go.

## Contributing

See [.github/CONTRIBUTING.md](.github/CONTRIBUTING.md).

## Users

If your company or project is using slog, feel free to open an issue or PR to amend this list.

- [Coder](https://github.com/cdr)
