# slog

[![GoDoc](https://godoc.org/go.coder.com/slog?status.svg)](https://godoc.org/go.coder.com/slog)
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/cdr/slog?color=critical&sort=semver)](https://github.com/cdr/slog/releases)
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
- First class [\*testing.T](https://godoc.org/go.coder.com/slog/testlog) support

## Example

```
testlog.Info(t, "my message here",
    slog.F("field_name", "something or the other"),
    slog.F("some_map", map[string]interface{}{
        "nested_fields": "wowow",
    }),
    slog.F("some slice", []interface{}{
        1,
        "foof",
        "bar",
        true,
    }),
    slog.Component("test"),

    slog.F("name", slog.ValueFunc(func() interface{} {
        return "wow"
    })),
)

// --- PASS: TestExampleTest (0.00s)
//    test_test.go:38: Sep 06 14:33:52.628 [INFO] (test): my_message_here
//        field_name: something or the other
//        some_map:
//          nested_fields: wowow
//        error:
//          - msg: wrap2
//            loc: /Users/nhooyr/src/cdr/slog/test_test.go:43
//            fun: go.coder.com/slog_test.TestExampleTest
//          - msg: wrap1
//            loc: /Users/nhooyr/src/cdr/slog/test_test.go:44
//            fun: go.coder.com/slog_test.TestExampleTest
//          - EOF
//        name: wow
```

## Design justifications

See [#9](https://github.com/cdr/slog/issues/9)

## Comparison

### zap

https://github.com/uber-go/zap

See [#6][https://github.com/cdr/slog/issues/6]

## Contributing

See [.github/CONTRIBUTING.md](.github/CONTRIBUTING.md).

## Users

If your company or project is using this library, feel free to open an issue or PR to amend this list.

- [Coder](https://github.com/cdr)
