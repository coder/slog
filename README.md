# slog

[![GitHub Release](https://img.shields.io/github/v/release/cdr/slog?color=6b9ded&sort=semver)](https://github.com/cdr/slog/releases)
[![GoDoc](https://godoc.org/cdr.dev/slog?status.svg)](https://godoc.org/cdr.dev/slog)
[![Coveralls](https://img.shields.io/coveralls/github/cdr/slog?color=65d6a4)](https://coveralls.io/github/cdr/slog)
[![CI Status](https://github.com/cdr/slog/workflows/ci/badge.svg)](https://github.com/cdr/slog/actions)

slog is a minimal structured logging library for Go.

## Install

```bash
go get cdr.dev/slog
```

## Features

- Minimal API
- Tiny codebase
- First class [context.Context](https://blog.golang.org/context) support
- First class [testing.TB](https://godoc.org/cdr.dev/slog/slogtest) support
- Beautiful human readable logging output
  - Prints multiline fields and errors nicely
- Machine readable JSON output
- [GCP Stackdriver](https://godoc.org/cdr.dev/slog/sloggers/slogstackdriver) support
- [Tee](https://godoc.org/cdr.dev/slog#Tee) multiple loggers
- [Stdlib](https://godoc.org/cdr.dev/slog#Stdlib) log adapter
- Skip caller frames with [slog.Helper](https://godoc.org/cdr.dev/slog#Helper)
- Can encode any Go structure including private fields
- Transparently logs [opencensus](https://godoc.org/go.opencensus.io/trace) trace and span IDs
- [Single dependency](https://godoc.org/cdr.dev/slog?imports) on go.opencensus.io

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

These are the main reasons we decided it was worth creating another log package for Go:

1. Very large API surface. Compare [zap](https://godoc.org/go.uber.org/zap) and
   [zapcore](https://godoc.org/go.uber.org/zap/zapcore) to [slog](https://godoc.org/cdr.dev/slog)

   - The sprawling API has made it hard to understand, use and extend.

1. zap's typed API is too verbose.

   - It does offer a [sugared API](https://godoc.org/go.uber.org/zap#hdr-Choosing_a_Logger)
     but it's too easy to pass an invalid fields list since there is no static type checking.
     Furthermore, it's harder to read as there is no syntax grouping for each key value pair.
   - We wanted an API that only accepted the equivalent of [zap.Any](https://godoc.org/go.uber.org/zap#Any) for every field.
     This is [slog.F](https://godoc.org/cdr.dev/slog#F).

1. zap's human readable format is not easy to read.

   - Lack of appropriate colors for different levels and fields
     - slog colors distinct parts of each line to make it easier to scan logs. Even the JSON that represents
       the fields in each log is syntax highlighted so that is very easy to scan. See the screenshot above.
   - zap logs multiline fields and errors stack traces as JSON strings which made them unreadable in a terminal.
     - slog automatically prints one multiline field after the log to make errors and such much easier to read.
       slog also automatically prints a Go 1.13 error chain as an array. See screenshot above.

1. Lack of [context.Context](https://blog.golang.org/context) support.

   - We wanted to be able to pull up all relevant logs for a given trace, user or request. With zap, we were plugging
     these fields in for every relevant log or passing around a logger with the fields set. This became very verbose.
     - `slog` lets you set fields in a `context.Context` such that any log with the context prints those fields.

1. zap is hard and confusing to extend. There are too many structures and configuration options.

   - With slog, the extension model is based on the single method Sink interface. Logger is the
     concrete type around Sink used to provide the high level API.

1. We found ourselves often implementing zap's [ObjectMarshaler](https://godoc.org/go.uber.org/zap/zapcore#ObjectMarshaler)
   to log Go structures.

   - This was very verbose and most of the time we ended up only implementing `fmt.Stringer` and using `zap.Stringer`
     instead.
   - slog handles Go structures transparently, including private fields. One
     may implement [`slog.Value`](https://godoc.org/cdr.dev/slog#Value) to override the representation,
     use struct tags to ignore or rename fields and even reuse the
     [`json.Marshal`](https://golang.org/pkg/encoding/json/#Marshal) representation
     with [`slog.JSON`](https://godoc.org/cdr.dev/slog#JSON).

1. We had many helper functions for logging but we wanted the line reported to be of the parent function.
   zap has an [API](https://godoc.org/go.uber.org/zap#AddCallerSkip) for this but it's verbose and requires
   passing the logger around explicitly.

   - slog takes inspriation from Go's stdlib and implements [`slog.Helper`](https://godoc.org/cdr.dev/slog#Helper) which works just like
     [`t.Helper`](https://golang.org/pkg/testing/#T.Helper)

1. We wanted tighter integration with stdlib's [`testing`](https://golang.org/pkg/testing) package.
   - zap has [zaptest](https://godoc.org/go.uber.org/zap/zaptest) but the API surface is large and doesn't
     integrate well. It has no support for failing on any ERROR logs nor does it have a stateless API
     that logs directly to a [`testing.TB`](https://golang.org/pkg/testing/#TB) without creating a
     intermediary logger.

## Users

If your company or project is using slog, feel free to open an issue or PR to amend this list.

- [Coder](https://github.com/cdr)
