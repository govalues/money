# money

[![githubb]][github]
[![codecovb]][codecov]
[![goreportb]][goreport]
[![godocb]][godoc]
[![licenseb]][license]
[![versionb]][version]
[![awesomeb]][awesome]

Package money implements immutable monetary amounts for Go.

## Getting started

### Installation

To add the money package to your Go workspace:

```bash
go get github.com/govalues/money
```

### Usage

Create amount using one of the constructors.
After creating a monetary amount, various operations can be performed:

```go
package main

import (
    "fmt"
    "github.com/govalues/decimal"
    "github.com/govalues/money"
)

func main() {
    x, _ := decimal.New(2, 0)                          // x = 2

    // Constructors
    a, _ := money.NewAmount("USD", 8, 0)               // a = USD 8.00
    b, _ := money.ParseAmount("USD", "12.5")           // b = USD 12.50
    c, _ := money.NewAmountFromFloat64("USD", 2.567)   // c = USD 2.567
    d, _ := money.NewAmountFromInt64("USD", 7, 896, 3) // d = USD 7.896
    r, _ := money.NewExchRate("USD", "EUR", 9, 1)      // r = USD/EUR 0.9

    // Operations
    fmt.Println(a.Add(b))          // USD 8.00 + USD 12.50
    fmt.Println(a.Sub(b))          // USD 8.00 - USD 12.50

    fmt.Println(a.Mul(x))          // USD 8.00 * 2
    fmt.Println(a.FMA(x, b))       // USD 8.00 * 2 + USD 12.50
    fmt.Println(r.Conv(a))         // USD/EUR 0.9 * USD 8.00

    fmt.Println(a.Quo(x))          // USD 8.00 / 2
    fmt.Println(a.QuoRem(x))       // USD 8.00 div 2, USD 8.00 mod 2
    fmt.Println(a.Rat(b))          // USD 8.00 / USD 12.50
    fmt.Println(a.Split(3))        // USD 8.00 into 3 parts

    // Rounding to 2 decimal places
    fmt.Println(c.RoundToCurr())   // 2.57
    fmt.Println(c.CeilToCurr())    // 2.57
    fmt.Println(c.FloorToCurr())   // 2.56
    fmt.Println(c.TruncToCurr())   // 2.56

    // Conversions
    fmt.Println(d.Int64(9))        // 7 896000000
    fmt.Println(d.Float64())       // 7.896
    fmt.Println(d.String())        // USD 7.896

    // Formatting
    fmt.Printf("%v\n", d)          // USD 7.896
    fmt.Printf("%[1]f %[1]c\n", d) // 7.896 USD
    fmt.Printf("%f\n", d)          // 7.896
    fmt.Printf("%c\n", d)          // USD
    fmt.Printf("%d\n", d)          // 790
}
```

## Documentation

For detailed documentation and additional examples, visit the package
[documentation](https://pkg.go.dev/github.com/govalues/money#pkg-examples).

## Contributing

Interested in contributing? Here's how to get started:

1. Fork and clone the repository.
1. Implement your changes.
1. Write tests to cover your changes.
1. Ensure all tests pass with `go test`.
1. Commit and push to your fork.
1. Open a pull request detailing your changes.

**Note**: If you're considering significant changes, please open an issue first to
discuss with the maintainers.
This ensures alignment with the project's objectives and roadmap.

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
[awesome]: https://github.com/avelino/awesome-go#financial
[awesomeb]: https://awesome.re/mentioned-badge.svg
