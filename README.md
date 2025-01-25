# money

[![githubb]][github]
[![codecovb]][codecov]
[![goreportb]][goreport]
[![godocb]][godoc]
[![licenseb]][license]
[![versionb]][version]
[![awesomeb]][awesome]

Package money implements immutable monetary amounts and exchange rates for Go.

## Key Features

- **Immutability** - Once set, an amount or exchange rate remains constant,
  ensuring safe concurrent access across goroutines.
- **Banker's Rounding** - Uses half-to-even rounding, also known as "banker's rounding",
  to minimize cumulative rounding errors commonly seen in financial calculations.
- **No Panics** - All methods are panic-free, returning errors instead of crashing
  your application in cases such as overflow, division by zero, or currency mismatch.
- **Zero Heap Allocation** - Optimized to avoid heap allocations,
  reducing garbage collector impact during arithmetic operations.
- **High Precision** - Supports up to 19 digits of precision, representing amounts
  from -99,999,999,999,999,999.99 to 99,999,999,999,999,999.99 inclusive.
- **Correctness** - Arithmetic operations are cross-validated against the
  [cockroachdb/apd] and [shopspring/decimal] packages through extensive [fuzz testing].

## Getting Started

### Installation

To add the money package to your Go workspace:

```bash
go get github.com/govalues/money
```

### Basic Usage

Create an amount using one of the constructors.
After creating an amount, you can perform various operations as shown below:

```go
package main

import (
    "fmt"
    "github.com/govalues/decimal"
    "github.com/govalues/money"
)

func main() {
    // Constructors
    a, _ := money.NewAmount("USD", 8, 0)               // a = $8.00
    b, _ := money.ParseAmount("USD", "12.5")           // b = $12.50
    c, _ := money.NewAmountFromFloat64("USD", 2.567)   // c = $2.567
    d, _ := money.NewAmountFromInt64("USD", 7, 896, 3) // d = $7.896
    r, _ := money.NewExchRate("USD", "EUR", 9, 1)      // r = $/€ 0.9
    x, _ := decimal.New(2, 0)                          // x = 2

    // Operations
    fmt.Println(a.Add(b))          // $8.00 + $12.50
    fmt.Println(a.Sub(b))          // $8.00 - $12.50
    fmt.Println(a.SubAbs(b))       // abs($8.00 - $12.50)

    fmt.Println(a.Mul(x))          // $8.00 * 2
    fmt.Println(a.AddMul(b, x))    // $8.00 + $12.50 * 2
    fmt.Println(a.SubMul(b, x))    // $8.00 - $12.50 * 2
    fmt.Println(r.Conv(a))         // $8.00 * $/€ 0.9

    fmt.Println(a.Quo(x))          // $8.00 / 2
    fmt.Println(a.AddQuo(b, x))    // $8.00 + $12.50 / 2
    fmt.Println(a.SubQuo(b, x))    // $8.00 - $12.50 / 2
    fmt.Println(a.QuoRem(x))       // $8.00 div 2, $8.00 mod 2
    fmt.Println(a.Rat(b))          // $8.00 / $12.50
    fmt.Println(a.Split(3))        // $8.00 into 3 parts

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
    fmt.Printf("%v", d)            // USD 7.896
    fmt.Printf("%[1]f %[1]c", d)   // 7.896 USD
    fmt.Printf("%f", d)            // 7.896
    fmt.Printf("%c", d)            // USD
    fmt.Printf("%d", d)            // 790
}
```

## Documentation

For detailed documentation and additional examples, visit the package
[documentation](https://pkg.go.dev/github.com/govalues/money#section-documentation).

## Comparison

Comparison with other popular packages:

| Feature                         | govalues       | [rhymond] v1.0.10 | [bojanz] v1.2.1 |
| ------------------------------- | -------------- | ----------------- | --------------- |
| Speed                           | High           | Medium            | Medium          |
| Numeric Representation          | Floating Point | Fixed Point       | Floating Point  |
| Precision                       | 19 digits      | 18 digits         | 39 digits       |
| Default Rounding                | Half to even   | Not supported     | Half up         |
| Overflow Control                | Yes            | No[^wrap]         | Yes             |
| Support for Division            | Yes            | No                | Yes             |
| Support for Currency Conversion | Yes            | No                | Yes             |

[^wrap]: [rhymond] does not detect overflow and returns an invalid result.
For example, 92,233,720,368,547,758.07 + 0.01 results in -92,233,720,368,547,758.08.

### Benchmarks

```text
goos: linux
goarch: amd64
pkg: github.com/govalues/money-tests
cpu: AMD Ryzen 7 3700C  with Radeon Vega Mobile Gfx 
```

| Test Case     | Expression            | govalues | [rhymond] v1.0.10 | [bojanz] v1.2.1 | govalues vs rhymond | govalues vs bojanz |
| ------------- | --------------------- | -------: | ----------------: | --------------: | ------------------: | -----------------: |
| Add           | $2.00 + $3.00         |   22.95n |           218.30n |         144.10n |            +851.41% |           +528.02% |
| Mul           | $2.00 * 3             |   21.80n |           133.40n |         239.60n |            +511.79% |           +998.83% |
| Quo (exact)   | $2.00 / 4             |   80.12n |               n/a |         468.05n |                 n/a |           +484.19% |
| Quo (inexact) | $2.00 / 3             |   602.1n |               n/a |          512.4n |                 n/a |            -14.91% |
| Split         | $2.00 into 10 parts   |   374.9n |            897.0n |             n/a |            +139.28% |                n/a |
| Conv          | $2.00 to €            |   30.88n |               n/a |         348.50n |                 n/a |          +1028.38% |
| Parse         | $1                    |   44.99n |           139.50n |          99.09n |            +210.07% |           +120.26% |
| Parse         | $123.456              |   61.45n |           148.60n |         240.90n |            +141.82% |           +292.03% |
| Parse         | $123456789.1234567890 |   131.2n |            204.4n |          253.0n |             +55.85% |            +92.87% |
| String        | $1                    |   38.48n |           200.70n |          89.92n |            +421.50% |           +133.65% |
| String        | $123.456              |   56.34n |           229.90n |         127.05n |            +308.02% |           +125.49% |
| String        | $123456789.1234567890 |   84.73n |           383.30n |         277.55n |            +352.38% |           +227.57% |
| Telco         | see [specification]   |   224.2n |               n/a |         1944.0n |                 n/a |           +766.89% |

The benchmark results shown in the table are provided for informational purposes
only and may vary depending on your specific use case.

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
[cockroachdb/apd]: https://pkg.go.dev/github.com/cockroachdb/apd
[shopspring/decimal]: https://pkg.go.dev/github.com/shopspring/decimal
[specification]: https://speleotrove.com/decimal/telcoSpec.html
[fuzz testing]: https://github.com/govalues/decimal-tests
