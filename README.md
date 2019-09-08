# slog

[![GoDoc](https://godoc.org/go.coder.com/slog?status.svg)](https://godoc.org/go.coder.com/slog)
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/cdr/slog?include_prereleases&sort=semver)](https://github.com/cdr/slog/releases)
[![Codecov](https://img.shields.io/codecov/c/github/cdr/slog.svg?color=success)](https://codecov.io/gh/cdr/slog)
[![CI](https://img.shields.io/circleci/build/github/cdr/slog?label=ci)](https://github.com/cdr/slog/commits/master)

slog is a minimal structured logging library for Go.

## Install

```bash
go get go.coder.com/slog
```

## Features

- Minimal API
- Tiny codebase
- First class [context.Context](https://blog.golang.org/context) support
- Beautiful logging output by default
- Multiple adapters
- First class [testing.TB](https://godoc.org/go.coder.com/slog/slogtest) support

## Example

```go
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
//    examples_test.go:46: Sep 08 13:54:34.532 [INFO] (test): my_message_here
//        field_name: something or the other
//        some_map:
//          nested_fields: wowow
//        error:
//          - wrap1
//            go.coder.com/slog_test.TestExample
//              /Users/nhooyr/src/cdr/slog/examples_test.go:52
//          - wrap2
//            go.coder.com/slog_test.TestExample
//              /Users/nhooyr/src/cdr/slog/examples_test.go:53
//          - EOF
```

## Design justifications

See [#9](https://github.com/cdr/slog/issues/9)

## Comparison

### zap

https://github.com/uber-go/zap

See [#6](https://github.com/cdr/slog/issues/6).

## Contributing

See [.github/CONTRIBUTING.md](.github/CONTRIBUTING.md).

## Users

If your company or project is using this library, feel free to open an issue or PR to amend this list.

- [Coder](https://github.com/cdr)
