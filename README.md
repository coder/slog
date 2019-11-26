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
                io.EOF),
        ),
    ),
)
```

![Example output screenshot](https://i.imgur.com/o8uW4Oy.png)

## Why?

We have been using Go at [Coder](https://github.com/cdr) for several years during
which we used Uber's [zap](https://github.com/uber-go/zap) for logging.

It's a fantastic library for performance but the API and developer experience is not great.

The API surface is extremely large. See [godoc](https://godoc.org/go.uber.org/zap). And that's not including
zap's subpackage [zapcore](https://godoc.org/go.uber.org/zap/zapcore) which is a beast of its own.

The sprawling API makes it hard to understand, use and extend.

## Contributing

See [.github/CONTRIBUTING.md](.github/CONTRIBUTING.md).

## Users

If your company or project is using this library, feel free to open an issue or PR to amend this list.

- [Coder](https://github.com/cdr)
