# money

[![githubb]][github]
[![codecovb]][codecov]
[![goreportb]][goreport]
[![godocb]][godoc]
[![licenseb]][license]
[![versionb]][version]
[![awesomeb]][awesome]

Package money implements immutable monetary amounts and exchange rates for Go.

## Features

- **Optimized Performance** - Utilizes `uint64` for coefficients, reducing heap
  allocations and memory consumption.
- **High Precision** - Supports 19 digits of precision, which allows representation
  of amounts between -99,999,999,999,999,999.99 and 99,999,999,999,999,999.99 inclusive.
- **Immutability** - Once an amount or exchange rate is set, it remains unchanged.
  This immutability ensures safe concurrent access across goroutines.
- **Banker's Rounding** - Methods use half-to-even rounding, also known as "banker's rounding",
  which minimizes cumulative rounding errors commonly seen in financial calculations.
- **No Panics** - All methods are designed to be panic-free.
  Instead of potentially crashing your application, they return errors for issues
  such as overflow, division by zero, or currency mismatch.
- **Correctness** - Fuzz testing is used to [cross-validate] arithmetic operations
  against the [cockroachdb] and [shopspring] decimal packages.

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
    // Constructors
    a, _ := money.NewAmount("USD", 8, 0)               // a = USD 8.00
    b, _ := money.ParseAmount("USD", "12.5")           // b = USD 12.50
    c, _ := money.NewAmountFromFloat64("USD", 2.567)   // c = USD 2.567
    d, _ := money.NewAmountFromInt64("USD", 7, 896, 3) // d = USD 7.896
    r, _ := money.NewExchRate("USD", "EUR", 9, 1)      // r = USD/EUR 0.9
    x, _ := decimal.New(2, 0)                          // x = 2

    // Operations
    fmt.Println(a.Add(b))          // USD 8.00 + USD 12.50
    fmt.Println(a.Sub(b))          // USD 8.00 - USD 12.50

    fmt.Println(a.Mul(x))          // USD 8.00 * 2
    fmt.Println(a.FMA(x, b))       // USD 8.00 * 2 + USD 12.50
    fmt.Println(r.Conv(a))         // USD 8.00 * USD/EUR 0.9

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

## Comparison

Comparison with other popular packages:

| Feature                         | govalues       | [rhymond] v1.0.10 | [bojanz] v1.2.1 |
| ------------------------------- | -------------- | ----------------- | --------------- |
| Speed                           | High           | Medium            | Medium          |
| Numeric Representation          | Floating Point | Fixed Point       | Floating Point  |
| Precision                       | 19 digits      | 18 digits         | 39 digits       |
| Default Rounding                | Half to even   | Not supported     | Half up         |
| Overflow Control                | Yes            | No[^wraparound]   | Yes             |
| Support for Division            | Yes            | No                | Yes             |
| Support for Currency Conversion | Yes            | No                | Yes             |

[^wraparound] [rhymond] does not detect overflow and returns an invalid result.
For example, 92,233,720,368,547,758.07 + 0.01 results in -92,233,720,368,547,758.08.

### Benchmarks

```text
goos: linux
goarch: amd64
pkg: github.com/govalues/money-tests
cpu: AMD Ryzen 7 3700C  with Radeon Vega Mobile Gfx 
```

| Test Case   | Expression                | govalues | [rhymond] v1.0.10 | [bojanz] v1.2.1 | govalues vs rhymond | govalues vs bojanz |
| ----------- | ------------------------- | -------: | ----------------: | --------------: | ------------------: | -----------------: |
| Add         | USD 2.00 + USD 3.00       |   22.95n |           218.30n |         144.10n |            +851.41% |           +528.02% |
| Mul         | USD 2.00 * 3              |   21.80n |           133.40n |         239.60n |            +511.79% |           +998.83% |
| QuoFinite   | USD 2.00 / 4              |   80.12n |       n/a[^nodiv] |         468.05n |                 n/a |           +484.19% |
| QuoInfinite | USD 2.00 / 3              |   602.1n |       n/a[^nodiv] |          512.4n |                 n/a |            -14.91% |
| Split       | USD 2.00 into 10 parts    |   374.9n |            897.0n |   n/a[^nosplit] |            +139.28% |                n/a |
| Conv        | USD 2.00 * USD/EUR 0.8000 |   30.88n |      n/a[^noconv] |         348.50n |                 n/a |          +1028.38% |
| Parse       | USD 1                     |   44.99n |           139.50n |          99.09n |            +210.07% |           +120.26% |
| Parse       | USD 123.456               |   61.45n |           148.60n |         240.90n |            +141.82% |           +292.03% |
| Parse       | USD 123456789.1234567890  |   131.2n |            204.4n |          253.0n |             +55.85% |            +92.87% |
| String      | USD 1                     |   38.48n |           200.70n |          89.92n |            +421.50% |           +133.65% |
| String      | USD 123.456               |   56.34n |           229.90n |         127.05n |            +308.02% |           +125.49% |
| String      | USD 123456789.1234567890  |   84.73n |           383.30n |         277.55n |            +352.38% |           +227.57% |
| Telco       | see [specification]       |   224.2n |   n/a[^nofracmul] |         1944.0n |                 n/a |           +766.89% |

[^nodiv]: [rhymond] does not support division.

[^noconv]: [rhymond] does not support currency conversion.

[^nofracmul]: [rhymond] does not support multiplication by a fraction.

[^nosplit]: [bojanz] does not support splitting into parts.

The benchmark results shown in the table are provided for informational purposes
only and may vary depending on your specific use case.

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
[rhymond]: https://pkg.go.dev/github.com/Rhymond/go-money
[bojanz]: https://pkg.go.dev/github.com/bojanz/currency
[cockroachdb]: https://pkg.go.dev/github.com/cockroachdb/apd
[shopspring]: https://pkg.go.dev/github.com/shopspring/decimal
[specification]: https://speleotrove.com/decimal/telcoSpec.html
[cross-validate]: https://github.com/govalues/decimal-tests/blob/main/decimal_fuzz_test.go
