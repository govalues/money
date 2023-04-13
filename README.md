# Money

[![githubb]][github]
[![codecovb]][codecov]
[![goreportb]][goreport]
[![licenseb]][license]
[![godocb]][godoc]
[![versionb]][version]

Package money implements immutable monetary amounts for Go.

## Getting started

To install the money package into your Go workspace, you can use the go get command:

```bash
go get github.com/govalues/money
```

To use the money package in your Go project, you can import it as follows:

```go
import (
    "github.com/govalues/decimal"
    "github.com/govalues/money"
)
```

## Using Amount

To create a new amount, you can use one of the provided constructors,
such as `NewAmount`, `ParseAmount` or `MustParseAmount`.

```go
d := decimal.New(12345, 2)                  // d = 123.45
a, _ := money.NewAmount(money.USD, d)       // a = USD 123.45
b := money.MustParseAmount("USD", "123.45") // b = USD 123.45
```

Once you have an amount, you can perform arithmetic operations such as
addition, subtraction, multiplication, division, as well
as rounding operations such as ceiling, floor, truncation, and rounding.

```go
sum := a.Add(b)
difference := a.Sub(b)
product := a.Mul(d)
quotient := a.Quo(d)
ratio := a.Rat(b)
ceil := a.Ceil(2)
floor := a.Floor(2)
trunc := a.Trunc(2)
round := a.Round(2)
```

For more details on these and other methods, see the package documentation at [pkg.go.dev](https://pkg.go.dev/github.com/govalues/money).

## Benchmarks

For the benchmark results please check description of [decimal](https://github.com/govalues/decimal) package.

## Contributing to the project

The money package is hosted on [GitHub](https://github.com/govalues/money).
To contribute to the project, follow these steps:

 1. Fork the repository and clone it to your local machine.
 1. Make the desired changes to the code.
 1. Write tests for the changes you made.
 1. Ensure that all tests pass by running `go test`.
 1. Commit the changes and push them to your fork.
 1. Submit a pull request with a clear description of the changes you made.
 1. Wait for the maintainers to review and merge your changes.

Note: Before making any significant changes to the code, it is recommended to open an issue to discuss the proposed changes with the maintainers. This will help to ensure that the changes align with the project's goals and roadmap.

[codecov]: https://codecov.io/gh/govalues/money
[codecovb]: https://img.shields.io/codecov/c/github/govalues/money/main?color=brightcolor
[goreport]: https://goreportcard.com/report/github.com/govalues/money
[goreportb]: https://goreportcard.com/badge/github.com/govalues/money
[github]: https://github.com/govalues/money/actions/workflows/go.yml
[githubb]: https://img.shields.io/github/actions/workflow/status/govalues/money/go.yml
[godoc]: https://pkg.go.dev/github.com/govalues/money#section-documentation
[godocb]: https://img.shields.io/badge/go.dev-reference-blue
[version]: https://go.dev/dl
[versionb]: https://img.shields.io/github/go-mod/go-version/govalues/money?label=go
[license]: https://en.wikipedia.org/wiki/MIT_License
[licenseb]: https://img.shields.io/github/license/govalues/money?color=blue
