/*
Package money implements immutable amounts and exchange rates.

# Representation

[Currency] is represented as an integer index in an in-memory array that
stores properties defined by [ISO 4217]:
  - Code: a three-letter alphabetic code.
  - Num: a three-digit numeric code.
  - Scale: a non-negative integer indicating the number of digits after
    the decimal point needed to represent minor units of the currency.

The currently supported currencies use scales of 0, 2, or 3:
  - A scale of 0 indicates currencies without minor units.
    For example, the [Japanese Yen] does not have minor units.
  - A scale of 2 indicates currencies with 2 digits for minor units.
    For example, the [US Dollar] represents its minor unit (1 cent)
    as 0.01 dollars.
  - A scale of 3 indicates currencies with 3 digits for minor units.
    For instance, the minor unit of the [Omani Rial], 1 baisa, is represented
    as 0.001 rials.

[Amount] is a struct with two fields:

  - Currency: a [Currency] in which the amount is denominated.
  - Amount: a [decimal.Decimal] representing the numeric value of the amount.

[ExchangeRate] is a struct with three fields:

  - Base: a [Currency] being exchanged.
  - Quote: a [Currency] being obtained in exchange for the base currency.
  - Rate: a positive [decimal.Decimal] representing how many units of the quote
    currency are needed to exchange for 1 unit of the base currency.

# Constraints

The range of an amount is determined by the scale of its currency.
Similarly, the range of an exchange rate is determined by the scale of its quote
currency.
Here are the ranges for scales 0, 2, and 3:

	| Example      | Scale | Minimum                              | Maximum                             |
	| ------------ | ----- | ------------------------------------ | ----------------------------------- |
	| Japanese Yen | 0     | -9,999,999,999,999,999,999           | 9,999,999,999,999,999,999           |
	| US Dollar    | 2     |    -99,999,999,999,999,999.99        |    99,999,999,999,999,999.99        |
	| Omani Rial   | 3     |     -9,999,999,999,999,999.999       |     9,999,999,999,999,999.999       |

Subnormal numbers are not supported by the underlying [decimal.Decimal] type.
Consequently, amounts and exchange rates between -0.00000000000000000005 and
0.00000000000000000005 inclusive are rounded to 0.

# Conversions

The package provides methods for converting:

  - from/to string:
    [ParseAmount], [Amount.String], [Amount.Format],
    [ParseExchRate], [ExchangeRate.String], [ExchangeRate.Format].
  - from/to float64:
    [NewAmountFromFloat64], [Amount.Float64],
    [NewExchRateFromFloat64], [ExchangeRate.Float64].
  - from/to int64:
    [NewAmount], [NewAmountFromInt64], [Amount.Int64],
    [NewAmountFromMinorUnits], [Amount.MinorUnits],
    [NewExchRate], [NewExchRateFromInt64], [ExchangeRate.Int64].
  - from/to decimal:
    [NewAmountFromDecimal], [Amount.Decimal],
    [NewExchRateFromDecimal], [ExchangeRate.Decimal].

See the documentation for each method for more details.

# Operations

Each arithmetic operation is performed in two steps:

 1. The operation is initially performed using uint64 arithmetic.
    If no overflow occurs, the exact result is immediately returned.
    If an overflow does occur, the operation proceeds to step 2.

 2. The operation is repeated with higher precision using [big.Int] arithmetic.
    The result is then rounded to 19 digits.
    If no significant digits are lost during rounding, the inexact result is returned.
    If any significant digits are lost, an overflow error is returned.

Step 1 was introduced to improve performance by avoiding heap allocation
for [big.Int] and the complexities associated with [big.Int] arithmetic.
In transactional financial systems, most arithmetic operations are expected to
compute an exact result during step 1.

The following rules determine the significance of digits during step 2:

  - [Amount.Add], [Amount.Sub], [Amount.SubAbs], [Amount.Mul], [Amount.AddMul],
    [Amount.AddQuo], [Amount.SubMul], [Amount.SubQuo],
    [Amount.Quo], [Amount.QuoRem], [ExchangeRate.Conv], [ExchangeRate.Mul]:
    All digits in the integer part are significant.
    In the fractional part, digits are significant up to the scale of
    the currency.

  - [Amount.Rat]:
    All digits in the integer part are significant, while digits in the
    fractional part are considered insignificant.

# Rounding

Implicit rounding applies when a result exceeds 19 digits.
In such cases, the result is rounded to 19 digits using half-to-even rounding.
This method ensures that rounding errors are evenly distributed between rounding up
and rounding down.

For all arithmetic operations, the result is the one that would be obtained by
computing the exact mathematical result with infinite precision and then
rounding it to 19 digits.

In addition to implicit rounding, the package provides several methods for
explicit rounding:

  - half-to-even rounding:
    [Amount.Round], [Amount.RoundToCurr], [Amount.Quantize], [Amount.Rescale],
    [ExchangeRate.Round], [ExchangeRate.Quantize], [ExchangeRate.Rescale].
  - rounding towards positive infinity:
    [Amount.Ceil], [Amount.CeilToCurr], [ExchangeRate.Ceil].
  - rounding towards negative infinity:
    [Amount.Floor], [Amount.FloorToCurr], [ExchangeRate.Floor].
  - rounding towards zero:
    [Amount.Trunc], [Amount.TruncToCurr], [ExchangeRate.Trunc].

See the documentation for each method for more details.

# Errors

All methods are panic-free and pure.
Errors are returned in the following cases:

  - Currency Mismatch.
    All arithmetic operations except for [Amount.Rat] return an error if
    the operands use different currencies.

  - Division by Zero.
    Unlike the standard library, [Amount.Quo], [Amount.QuoRem], [Amount.Rat], [Amount.AddQuo],
    and [Amount.SubQuo] do not panic when dividing by 0.
    Instead, they return an error.

  - Overflow.
    Unlike standard integers, there is no "wrap around" for amounts at certain sizes.
    Arithmetic operations return an error for out-of-range values.

  - Underflow.
    All arithmetic operations, except for [ExchangeRate.Mul],
    do not return an error in case of underflow.
    If the result is an amount between -0.00000000000000000005
    and 0.00000000000000000005 inclusive, it is rounded to 0.
    [ExchangeRate.Mul] return an error in cases of underflow,
    as the result of these operations is an exchange rate, and exchange rates
    cannot be 0.

[Japanese Yen]: https://en.wikipedia.org/wiki/Japanese_yen
[US Dollar]: https://en.wikipedia.org/wiki/United_States_dollar
[Omani Rial]: https://en.wikipedia.org/wiki/Omani_rial
[ISO 4217]: https://en.wikipedia.org/wiki/ISO_4217
[big.Int]: https://pkg.go.dev/math/big#Int
*/
package money
