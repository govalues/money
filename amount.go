package money

import (
	"errors"
	"fmt"
	"math"
	"strconv"

	"github.com/govalues/decimal"
)

var (
	errAmountOverflow   = errors.New("amount overflow")
	errCurrencyMismatch = errors.New("currency mismatch")
)

// Amount type represents a monetary amount.
// Its zero value corresponds to "XXX 0", where [XXX] indicates an unknown currency.
// Amount is designed to be safe for concurrent use by multiple goroutines.
type Amount struct {
	curr  Currency        // ISO 4217 currency
	value decimal.Decimal // monetary value
}

// newAmountUnsafe creates a new amount without checking the scale.
// Use it only if you are absolutely sure that the arguments are valid.
func newAmountUnsafe(c Currency, d decimal.Decimal) Amount {
	return Amount{curr: c, value: d}
}

// newAmountSafe creates a new amount and checks the scale.
func newAmountSafe(c Currency, d decimal.Decimal) (Amount, error) {
	if d.Scale() < c.Scale() {
		d = d.Pad(c.Scale())
		if d.Scale() < c.Scale() {
			return Amount{}, fmt.Errorf("padding amount: %w", errAmountOverflow)
		}
	}
	return newAmountUnsafe(c, d), nil
}

// NewAmount returns an amount equal to coef / 10^scale.
// If the scale of the amount is less than the scale of the currency, the result
// will be zero-padded to the right.
//
// NewAmount returns an error if:
//   - the currency code is not valid;
//   - the scale is negative or greater than [decimal.MaxScale];
//   - the integer part of the result has more than
//     ([decimal.MaxPrec] - [Currency.Scale]) digits.
//     For example, when currency is US Dollars, NewAmount will return an error
//     if the integer part of the result has more than 17 digits (19 - 2 = 17).
func NewAmount(curr string, coef int64, scale int) (Amount, error) {
	// Currency
	c, err := ParseCurr(curr)
	if err != nil {
		return Amount{}, fmt.Errorf("parsing currency: %w", err)
	}
	// Decimal
	d, err := decimal.New(coef, scale)
	if err != nil {
		return Amount{}, fmt.Errorf("converting coefficient: %w", err)
	}
	// Amount
	a, err := newAmountSafe(c, d)
	if err != nil {
		return Amount{}, fmt.Errorf("converting coefficient: %w", err)
	}
	return a, nil
}

// MustNewAmount is like [NewAmount] but panics if the amount cannot be constructed.
// It simplifies safe initialization of global variables holding amounts.
func MustNewAmount(curr string, coef int64, scale int) Amount {
	a, err := NewAmount(curr, coef, scale)
	if err != nil {
		panic(fmt.Sprintf("NewAmount(%q, %v, %v) failed: %v", curr, coef, scale, err))
	}
	return a
}

// NewAmountFromDecimal returns an amount with the specified currency and value.
// If the scale of the amount is less than the scale of the currency, the result
// will be zero-padded to the right. See also method [Amount.Decimal].
//
// NewAmountFromDecimal returns an error if the integer part of the result has more than
// ([decimal.MaxPrec] - [Currency.Scale]) digits.
// For example, when currency is US Dollars, NewAmountFromDecimal will return an error if
// the integer part of the result has more than 17 digits (19 - 2 = 17).
func NewAmountFromDecimal(curr Currency, amount decimal.Decimal) (Amount, error) {
	return newAmountSafe(curr, amount)
}

// NewAmountFromInt64 converts a pair of integers, representing the whole and
// fractional parts, to a (possibly rounded) amount equal to whole + frac / 10^scale.
// NewAmountFromInt64 deletes trailing zeros up to the scale of the currency.
// This method is useful for converting amounts from [protobuf] format.
// See also method [Amount.Int64].
//
// NewAmountFromInt64 returns an error if:
//   - the currency code is not valid;
//   - the whole and fractional parts have different signs;
//   - the scale is negative or greater than [decimal.MaxScale];
//   - frac / 10^scale is not within the range (-1, 1);
//   - the integer part of the result has more than
//     ([decimal.MaxPrec] - [Currency.Scale]) digits.
//     For example, when currency is US Dollars, NewAmountFromInt64 will return
//     an error if the integer part of the result has more than 17 digits (19 - 2 = 17).
//
// [protobuf]: https://github.com/googleapis/googleapis/blob/master/google/type/money.proto
func NewAmountFromInt64(curr string, whole, frac int64, scale int) (Amount, error) {
	// Currency
	c, err := ParseCurr(curr)
	if err != nil {
		return Amount{}, fmt.Errorf("parsing currency: %w", err)
	}
	// Whole
	d, err := decimal.New(whole, 0)
	if err != nil {
		return Amount{}, fmt.Errorf("converting integers: %w", err)
	}
	// Fraction
	f, err := decimal.New(frac, scale)
	if err != nil {
		return Amount{}, fmt.Errorf("converting integers: %w", err)
	}
	if !f.IsZero() {
		if !d.IsZero() && d.Sign() != f.Sign() {
			return Amount{}, fmt.Errorf("converting integers: inconsistent signs")
		}
		if !f.WithinOne() {
			return Amount{}, fmt.Errorf("converting integers: inconsistent fraction")
		}
		f = f.Trim(c.Scale())
		d, err = d.AddExact(f, c.Scale())
		if err != nil {
			return Amount{}, fmt.Errorf("converting integers: %w", err)
		}
	}
	// Amount
	return newAmountSafe(c, d)
}

// NewAmountFromMinorUnits converts an integer, representing minor units of
// currency (e.g. cents, pennies, fens), to an amount.
// See also method [Amount.MinorUnits].
//
// NewAmountFromMinorUnits returns an error if currency code is not valid.
func NewAmountFromMinorUnits(curr string, units int64) (Amount, error) {
	// Currency
	c, err := ParseCurr(curr)
	if err != nil {
		return Amount{}, fmt.Errorf("parsing currency: %w", err)
	}
	// Decimal
	d, err := decimal.New(units, c.Scale())
	if err != nil {
		return Amount{}, fmt.Errorf("converting minor units: %w", err)
	}
	// Amount
	return newAmountSafe(c, d)
}

// NewAmountFromFloat64 converts a float to a (possibly rounded) amount.
// See also method [Amount.Float64].
//
// NewAmountFromFloat64 returns an error if:
//   - the currency code is not valid;
//   - the float is a special value (NaN or Inf);
//   - the integer part of the result has more than
//     ([decimal.MaxPrec] - [Currency.Scale]) digits.
//     For example, when currency is US Dollars, NewAmountFromFloat64 will
//     return an error if the integer part of the result has more than 17
//     digits (19 - 2 = 17).
func NewAmountFromFloat64(curr string, amount float64) (Amount, error) {
	// Float
	if math.IsNaN(amount) || math.IsInf(amount, 0) {
		return Amount{}, fmt.Errorf("converting float: special value %v", amount)
	}
	s := strconv.FormatFloat(amount, 'f', -1, 64)
	// Amount
	a, err := ParseAmount(curr, s)
	if err != nil {
		return Amount{}, fmt.Errorf("converting float: %w", err)
	}
	return a, nil
}

// ParseAmount converts currency and decimal strings to a (possibly rounded) amount.
// If the scale of the amount is less than the scale of the currency, the result
// will be zero-padded to the right.
// See also constructors [ParseCurr] and [decimal.Parse].
func ParseAmount(curr, amount string) (Amount, error) {
	// Currency
	c, err := ParseCurr(curr)
	if err != nil {
		return Amount{}, fmt.Errorf("parsing currency: %w", err)
	}
	// Decimal
	d, err := decimal.ParseExact(amount, c.Scale())
	if err != nil {
		return Amount{}, fmt.Errorf("parsing amount: %w", err)
	}
	// Amount
	return newAmountSafe(c, d)
}

// MustParseAmount is like [ParseAmount] but panics if any of the strings cannot be parsed.
// This function simplifies safe initialization of global variables holding amounts.
func MustParseAmount(curr, amount string) Amount {
	a, err := ParseAmount(curr, amount)
	if err != nil {
		panic(fmt.Sprintf("ParseAmount(%q, %q) failed: %v", curr, amount, err))
	}
	return a
}

// MinorUnits returns a (possibly rounded) amount in minor units of currency
// (e.g. cents, pennies, fens).
// If the scale of the amount is greater than the scale of the currency, then
// the fractional part is rounded using [rounding half to even] (banker's rounding).
// See also constructor [NewAmountFromMinorUnits].
//
// If the result cannot be represented as an int64, then false is returned.
//
// [rounding half to even]: https://en.wikipedia.org/wiki/Rounding#Rounding_half_to_even
func (a Amount) MinorUnits() (units int64, ok bool) {
	d := a.RoundToCurr().Decimal()
	u := d.Coef()
	if d.IsNeg() {
		if u > -math.MinInt64 {
			return 0, false
		}
		return -int64(u), true
	}
	if u > math.MaxInt64 {
		return 0, false
	}
	return int64(u), true
}

// Float64 returns the nearest binary floating-point number rounded
// using [rounding half to even] (banker's rounding).
// See also constructor [NewAmountFromFloat64].
//
// This conversion may lose data, as float64 has a smaller precision
// than the decimal type.
//
// [rounding half to even]: https://en.wikipedia.org/wiki/Rounding#Rounding_half_to_even
func (a Amount) Float64() (f float64, ok bool) {
	return a.Decimal().Float64()
}

// Int64 returns a pair of integers representing the whole and (possibly
// rounded) fractional parts of the amount.
// If given scale is greater than the scale of the amount, then the fractional part
// is zero-padded to the right.
// If given scale is smaller than the scale of the amount, then the fractional part
// is rounded using [rounding half to even] (banker's rounding).
// The relationship between the amount and the returned values can be expressed
// as a = whole + frac / 10^scale.
// This method is useful for converting amounts to [protobuf] format.
// See also constructor [NewAmountFromInt64].
//
// Int64 returns false if the result cannot be represented as a pair of int64 values.
//
// [rounding half to even]: https://en.wikipedia.org/wiki/Rounding#Rounding_half_to_even
// [protobuf]: https://github.com/googleapis/googleapis/blob/master/google/type/money.proto
func (a Amount) Int64(scale int) (whole, frac int64, ok bool) {
	return a.Decimal().Int64(scale)
}

// Curr returns the currency of the amount.
func (a Amount) Curr() Currency {
	return a.curr
}

// Decimal returns the decimal representation of the amount.
func (a Amount) Decimal() decimal.Decimal {
	return a.value
}

// Sign returns:
//
//	-1 if a < 0
//	 0 if a = 0
//	+1 if a > 0
func (a Amount) Sign() int {
	return a.Decimal().Sign()
}

// IsNeg returns:
//
//	true  if a < 0
//	false otherwise
func (a Amount) IsNeg() bool {
	return a.Decimal().IsNeg()
}

// IsPos returns:
//
//	true  if a > 0
//	false otherwise
func (a Amount) IsPos() bool {
	return a.Decimal().IsPos()
}

// Abs returns the absolute value of the amount.
func (a Amount) Abs() Amount {
	return newAmountUnsafe(a.Curr(), a.Decimal().Abs())
}

// Neg returns an amount with the opposite sign.
func (a Amount) Neg() Amount {
	return newAmountUnsafe(a.Curr(), a.Decimal().Neg())
}

// CopySign returns an amount with the same sign as amount b.
// The currency of amount b is ignored.
// CopySign treates 0 as positive.
// See also method [Amount.Sign].
func (a Amount) CopySign(b Amount) Amount {
	d, e := a.Decimal(), b.Decimal()
	return newAmountUnsafe(a.Curr(), d.CopySign(e))
}

// Scale returns the number of digits after the decimal point.
// See also method [Amount.MinScale].
func (a Amount) Scale() int {
	return a.Decimal().Scale()
}

// MinScale returns the smallest scale that the amount can be rescaled to
// without rounding.
// See also method [Amount.Trim].
func (a Amount) MinScale() int {
	return a.Decimal().MinScale()
}

// IsZero returns:
//
//	true  if a = 0
//	false otherwise
func (a Amount) IsZero() bool {
	return a.Decimal().IsZero()
}

// IsOne returns:
//
//	true  if a = -1 or a = 1
//	false otherwise
func (a Amount) IsOne() bool {
	return a.Decimal().IsOne()
}

// IsInt returns true if there are no significant digits after the decimal point.
func (a Amount) IsInt() bool {
	return a.Decimal().IsInt()
}

// WithinOne returns:
//
//	true  if -1 < a < 1
//	false otherwise
func (a Amount) WithinOne() bool {
	return a.Decimal().WithinOne()
}

// Add returns the (possibly rounded) sum of amounts a and b.
//
// Add returns an error if:
//   - amounts are denominated in different currencies;
//   - the integer part of the result has more than ([decimal.MaxPrec] - [Currency.Scale]) digits.
//     For example, when currency is US Dollars, Add will return an error if the integer
//     part of the result has more than 17 digits (19 - 2 = 17).
func (a Amount) Add(b Amount) (Amount, error) {
	c, err := a.add(b)
	if err != nil {
		return Amount{}, fmt.Errorf("computing [%v + %v]: %w", a, b, err)
	}
	return c, nil
}

func (a Amount) add(b Amount) (Amount, error) {
	if !a.SameCurr(b) {
		return Amount{}, errCurrencyMismatch
	}
	c, d, e := a.Curr(), a.Decimal(), b.Decimal()
	d, err := d.AddExact(e, c.Scale())
	if err != nil {
		return Amount{}, err
	}
	return newAmountSafe(c, d)
}

// Sub returns the (possibly rounded) difference between amounts a and b.
//
// Sub returns an error if:
//   - amounts are denominated in different currencies;
//   - the integer part of the result has more than ([decimal.MaxPrec] - [Currency.Scale]) digits.
//     For example, when currency is US Dollars, Sub will return an error if the integer
//     part of the result has more than 17 digits (19 - 2 = 17).
func (a Amount) Sub(b Amount) (Amount, error) {
	c, err := a.sub(b)
	if err != nil {
		return Amount{}, fmt.Errorf("computing [%v - %v]: %w", a, b, err)
	}
	return c, nil
}

// SubAbs returns the (possibly rounded) absolute difference between amounts a and b.
//
// SubAbs returns an error if:
//   - amounts are denominated in different currencies;
//   - the integer part of the result has more than ([decimal.MaxPrec] - [Currency.Scale]) digits.
//     For example, when currency is US Dollars, SubAbs will return an error if the integer
//     part of the result has more than 17 digits (19 - 2 = 17).
func (a Amount) SubAbs(b Amount) (Amount, error) {
	c, err := a.sub(b)
	if err != nil {
		return Amount{}, fmt.Errorf("computing [abs(%v - %v)]: %w", a, b, err)
	}
	return c.Abs(), nil
}

func (a Amount) sub(b Amount) (Amount, error) {
	if !a.SameCurr(b) {
		return Amount{}, errCurrencyMismatch
	}
	c, d, e := a.Curr(), a.Decimal(), b.Decimal()
	d, err := d.SubExact(e, c.Scale())
	if err != nil {
		return Amount{}, err
	}
	return newAmountSafe(c, d)
}

// FMA returns the (possibly rounded) [fused multiply-addition] of amounts a, b, and factor e.
// It computes a * e + b without any intermediate rounding.
// This method is useful for improving the accuracy and performance of algorithms
// that involve the accumulation of products, such as daily interest accrual.
//
// FMA returns an error if:
//   - amounts are denominated in different currencies;
//   - the integer part of the result has more than ([decimal.MaxPrec] - [Currency.Scale]) digits.
//     For example, when currency is US Dollars, FMA will return an error if the integer
//     part of the result has more than 17 digits (19 - 2 = 17).
//
// [fused multiply-addition]: https://en.wikipedia.org/wiki/Multiply%E2%80%93accumulate_operation#Fused_multiply%E2%80%93add
func (a Amount) FMA(e decimal.Decimal, b Amount) (Amount, error) {
	c, err := a.fma(e, b)
	if err != nil {
		return Amount{}, fmt.Errorf("computing [%v * %v + %v]: %w", a, e, b, err)
	}
	return c, nil
}

func (a Amount) fma(e decimal.Decimal, b Amount) (Amount, error) {
	if !a.SameCurr(b) {
		return Amount{}, errCurrencyMismatch
	}
	c, d, f := a.Curr(), a.Decimal(), b.Decimal()
	d, err := d.FMAExact(e, f, c.Scale())
	if err != nil {
		return Amount{}, err
	}
	return newAmountSafe(c, d)
}

// Mul returns the (possibly rounded) product of amount a and factor e.
//
// Mul returns an error if the integer part of the result has more than
// ([decimal.MaxPrec] - [Currency.Scale]) digits.
// For example, when currency is US Dollars, Mul will return an error if the integer
// part of the result has more than 17 digits (19 - 2 = 17).
func (a Amount) Mul(e decimal.Decimal) (Amount, error) {
	c, err := a.mul(e)
	if err != nil {
		return Amount{}, fmt.Errorf("computing [%v * %v]: %w", a, e, err)
	}
	return c, nil
}

func (a Amount) mul(e decimal.Decimal) (Amount, error) {
	c, d := a.Curr(), a.Decimal()
	d, err := d.MulExact(e, c.Scale())
	if err != nil {
		return Amount{}, err
	}
	return newAmountSafe(c, d)
}

// Quo returns the (possibly rounded) quotient of amount a and divisor e.
// See also methods [Amount.QuoRem], [Amount.Rat], and [Amount.Split].
//
// Quo returns an error if:
//   - the divisor is 0;
//   - the integer part of the result has more than ([decimal.MaxPrec] - [Currency.Scale]) digits.
//     For example, when currency is US Dollars, Quo will return an error if the integer
//     part of the result has more than 17 digits (19 - 2 = 17).
func (a Amount) Quo(e decimal.Decimal) (Amount, error) {
	c, err := a.quo(e)
	if err != nil {
		return Amount{}, fmt.Errorf("computing [%v / %v]: %w", a, e, err)
	}
	return c, nil
}

func (a Amount) quo(e decimal.Decimal) (Amount, error) {
	c, d := a.Curr(), a.Decimal()
	d, err := d.QuoExact(e, c.Scale())
	if err != nil {
		return Amount{}, err
	}
	return newAmountSafe(c, d)
}

// QuoRem returns the quotient q and remainder r of amount a and divisor e
// such that a = e * q + r, where q has scale equal to the scale of its currency
// and the sign of the reminder r is the same as the sign of the dividend d.
// See also methods [Amount.Quo], [Amount.Rat], and [Amount.Split].
//
// QuoRem returns an error if:
//   - the divisor is 0;
//   - the integer part of the result has more than [decimal.MaxPrec] digits.
func (a Amount) QuoRem(e decimal.Decimal) (q, r Amount, err error) {
	q, r, err = a.quoRem(e)
	if err != nil {
		return Amount{}, Amount{}, fmt.Errorf("computing [%v div %v] and [%v mod %v]: %w", a, e, a, e, err)
	}
	return q, r, nil
}

func (a Amount) quoRem(e decimal.Decimal) (q, r Amount, err error) {
	// Quotient
	q, err = a.Quo(e)
	if err != nil {
		return Amount{}, Amount{}, err
	}

	// T-Division
	q = q.TruncToCurr()

	// Reminder
	r, err = q.Mul(e)
	if err != nil {
		return Amount{}, Amount{}, err
	}
	r, err = a.Sub(r)
	if err != nil {
		return Amount{}, Amount{}, err
	}
	return q, r, nil
}

// Rat returns the (possibly rounded) ratio between amounts a and b.
// This method is particularly useful for calculating exchange rates between
// two currencies or determining percentages within a single currency.
// See also methods [Amount.Quo], [Amount.QuoRem], and [Amount.Split].
//
// Rat returns an error if:
//   - the divisor is 0;
//   - the integer part of the result has more than [decimal.MaxPrec] digits.
func (a Amount) Rat(b Amount) (decimal.Decimal, error) {
	d, e := a.Decimal(), b.Decimal()
	d, err := d.Quo(e)
	if err != nil {
		return decimal.Decimal{}, fmt.Errorf("computing [%v / %v]: %w", a, b, err)
	}
	return d, nil
}

// Split returns a slice of amounts that sum up to the original amount,
// ensuring the parts are as equal as possible.
// If the original amount cannot be divided equally among the specified number
// of parts, the remainder is distributed among the first parts of the slice.
// See also methods [Amount.Quo], [Amount.QuoRem], and [Amount.Rat].
//
// Split returns an error if the number of parts is not a positive integer.
func (a Amount) Split(parts int) ([]Amount, error) {
	r, err := a.split(parts)
	if err != nil {
		return nil, fmt.Errorf("splitting %v into %v parts: %w", a, parts, err)
	}
	return r, nil
}

func (a Amount) split(parts int) ([]Amount, error) {
	// Parts
	par, err := decimal.New(int64(parts), 0)
	if err != nil {
		return nil, err
	}
	if !par.IsPos() {
		return nil, fmt.Errorf("number of parts must be positive")
	}

	// Quotient
	quo, err := a.Quo(par)
	if err != nil {
		return nil, err
	}
	quo = quo.Trunc(a.Scale())

	// Reminder
	rem, err := quo.Mul(par)
	if err != nil {
		return nil, err
	}
	rem, err = a.Sub(rem)
	if err != nil {
		return nil, err
	}
	ulp := rem.ULP().CopySign(rem)

	res := make([]Amount, parts)
	for i := 0; i < parts; i++ {
		res[i] = quo
		// Reminder distribution
		if !rem.IsZero() {
			rem, err = rem.Sub(ulp)
			if err != nil {
				return nil, err
			}
			res[i], err = res[i].Add(ulp)
			if err != nil {
				return nil, err
			}
		}
	}
	return res, nil
}

// One returns an amount with a value of 1, having the same currency and scale
// as amount a.
// See also methods [Amount.Zero], [Amount.ULP].
func (a Amount) One() Amount {
	return newAmountUnsafe(a.Curr(), a.Decimal().One())
}

// Zero returns an amount with a value of 0, having the same currency and scale
// as amount a.
// See also methods [Amount.One], [Amount.ULP].
func (a Amount) Zero() Amount {
	return newAmountUnsafe(a.Curr(), a.Decimal().Zero())
}

// ULP (Unit in the Last Place) returns the smallest representable positive difference
// between two amounts with the same scale as amount a.
// It can be useful for implementing rounding and comparison algorithms.
// See also methods [Amount.Zero], [Amount.One].
func (a Amount) ULP() Amount {
	return newAmountUnsafe(a.Curr(), a.Decimal().ULP())
}

// Ceil returns an amount rounded up to the specified number of digits after
// the decimal point using [rounding toward positive infinity].
// See also methods [Amount.CeilToCurr], [Amount.Floor].
//
// [rounding toward positive infinity]: https://en.wikipedia.org/wiki/Rounding#Rounding_up
func (a Amount) Ceil(scale int) Amount {
	c, d := a.Curr(), a.Decimal()
	d = d.Ceil(scale).Pad(c.Scale())
	return newAmountUnsafe(c, d)
}

// CeilToCurr returns an amount rounded up to the scale of its currency
// using [rounding toward positive infinity].
// See also methods [Amount.Ceil], [Amount.SameScaleAsCurr].
//
// [rounding toward positive infinity]: https://en.wikipedia.org/wiki/Rounding#Rounding_up
func (a Amount) CeilToCurr() Amount {
	return a.Ceil(a.Curr().Scale())
}

// Floor returns an amount rounded down to the specified number of digits after
// the decimal point using [rounding toward negative infinity].
// See also methods [Amount.FloorToCurr], [Amount.Ceil].
//
// [rounding toward negative infinity]: https://en.wikipedia.org/wiki/Rounding#Rounding_down
func (a Amount) Floor(scale int) Amount {
	c, d := a.Curr(), a.Decimal()
	d = d.Floor(scale).Pad(c.Scale())
	return newAmountUnsafe(c, d)
}

// FloorToCurr returns an amount rounded down to the scale of its currency
// using [rounding toward negative infinity].
// See also methods [Amount.Floor], [Amount.SameScaleAsCurr].
//
// [rounding toward negative infinity]: https://en.wikipedia.org/wiki/Rounding#Rounding_down
func (a Amount) FloorToCurr() Amount {
	return a.Floor(a.Curr().Scale())
}

// Trunc returns an amount truncated to the specified number of digits after
// the decimal point using [rounding toward zero].
// See also method [Amount.TruncToCurr].
//
// [rounding toward zero]: https://en.wikipedia.org/wiki/Rounding#Rounding_toward_zero
func (a Amount) Trunc(scale int) Amount {
	c, d := a.Curr(), a.Decimal()
	d = d.Trunc(scale).Pad(c.Scale())
	return newAmountUnsafe(c, d)
}

// TruncToCurr returns an amount truncated to the scale of its currency
// using [rounding toward zero].
// See also methods [Amount.Trunc], [Amount.SameScaleAsCurr].
//
// [rounding toward zero]: https://en.wikipedia.org/wiki/Rounding#Rounding_toward_zero
func (a Amount) TruncToCurr() Amount {
	return a.Trunc(a.Curr().Scale())
}

// Round returns an amount rounded to the specified number of digits after
// the decimal point using [rounding half to even] (banker's rounding).
// See also methods [Amount.Rescale], [Amount.RoundToCurr].
//
// [rounding half to even]: https://en.wikipedia.org/wiki/Rounding#Rounding_half_to_even
func (a Amount) Round(scale int) Amount {
	c, d := a.Curr(), a.Decimal()
	d = d.Round(scale).Pad(c.Scale())
	return newAmountUnsafe(c, d)
}

// RoundToCurr returns an amount rounded to the scale of its currency
// using [rounding half to even] (banker's rounding).
// See also methods [Amount.Round], [Amount.SameScaleAsCurr].
//
// [rounding half to even]: https://en.wikipedia.org/wiki/Rounding#Rounding_half_to_even
func (a Amount) RoundToCurr() Amount {
	return a.Round(a.Curr().Scale())
}

// Quantize returns an amount rescaled to the same scale as amount b.
// The currency and the sign of amount b are ignored.
// See also methods [Amount.Scale], [Amount.SameScale], [Amount.Rescale].
func (a Amount) Quantize(b Amount) Amount {
	return a.Rescale(b.Scale())
}

// Rescale returns an amount rounded or zero-padded to the given number of digits
// after the decimal point.
// See also method [Amount.Round].
func (a Amount) Rescale(scale int) Amount {
	c, d := a.Curr(), a.Decimal()
	d = d.Rescale(scale).Pad(c.Scale())
	return newAmountUnsafe(c, d)
}

// Trim returns an amount with trailing zeros removed up to the given scale.
// If the given scale is less than the scale of the currency, the zeros will be
// removed up to the scale of the currency instead.
// See also method [Amount.TrimToCurr].
func (a Amount) Trim(scale int) Amount {
	c, d := a.Curr(), a.Decimal()
	scale = max(scale, c.Scale())
	d = d.Trim(scale)
	return newAmountUnsafe(c, d)
}

// TrimToCurr returns an amount with trailing zeros removed up the scale of its currency.
// See also method [Amount.Trim].
func (a Amount) TrimToCurr() Amount {
	return a.Trim(a.Curr().Scale())
}

// SameCurr returns true if amounts are denominated in the same currency.
// See also method [Amount.Curr].
func (a Amount) SameCurr(b Amount) bool {
	return a.Curr() == b.Curr()
}

// SameScale returns true if amounts have the same scale.
// See also methods [Amount.Scale], [Amount.Quantize].
func (a Amount) SameScale(b Amount) bool {
	d, e := a.Decimal(), b.Decimal()
	return d.SameScale(e)
}

// SameScaleAsCurr returns true if the scale of the amount is equal to the scale of
// its currency.
// See also methods [Amount.Scale], [Currency.Scale], [Amount.RoundToCurr].
func (a Amount) SameScaleAsCurr() bool {
	return a.Scale() == a.Curr().Scale()
}

// String implements the [fmt.Stringer] interface and returns a string
// representation of an amount.
// See also methods [Currency.String], [Decimal.String], [Amount.Format].
//
// [fmt.Stringer]: https://pkg.go.dev/fmt#Stringer
// [Decimal.String]: https://pkg.go.dev/github.com/govalues/decimal#Decimal.String
func (a Amount) String() string {
	var buf [32]byte
	pos := len(buf) - 1
	coef := a.Decimal().Coef()
	scale := a.Decimal().Scale()

	// Coefficient
	for {
		buf[pos] = byte(coef%10) + '0'
		pos--
		coef /= 10
		if scale > 0 {
			scale--
			// Decimal point
			if scale == 0 {
				buf[pos] = '.'
				pos--
				// Leading 0
				if coef == 0 {
					buf[pos] = '0'
					pos--
				}
			}
		}
		if coef == 0 && scale == 0 {
			break
		}
	}

	// Sign
	if a.Decimal().IsNeg() {
		buf[pos] = '-'
		pos--
	}

	// Delimiter
	buf[pos] = ' '
	pos--

	// Currency
	curr := a.Curr().Code()
	for i := len(curr) - 1; i >= 0; i-- {
		buf[pos] = curr[i]
		pos--
	}

	return string(buf[pos+1:])
}

// Cmp compares amounts and returns:
//
//	-1 if a < b
//	 0 if a = b
//	+1 if a > b
//
// See also methods [Amount.CmpAbs], [Amount.CmpTotal].
//
// Cmp returns an error if amounts are denominated in different currencies.
func (a Amount) Cmp(b Amount) (int, error) {
	if !a.SameCurr(b) {
		return 0, fmt.Errorf("comparing [%v] and [%v]: %w", a, b, errCurrencyMismatch)
	}
	d, e := a.Decimal(), b.Decimal()
	return d.Cmp(e), nil
}

// CmpAbs compares absolute values of amounts and returns:
//
//	-1 if |a| < |b|
//	 0 if |a| = |b|
//	+1 if |a| > |b|
//
// See also methods [Amount.Cmp], [Amount.CmpTotal].
//
// CmpAbs returns an error if amounts are denominated in different currencies.
func (a Amount) CmpAbs(b Amount) (int, error) {
	if !a.SameCurr(b) {
		return 0, fmt.Errorf("comparing [abs(%v)] and [abs(%v)]: %w", a, b, errCurrencyMismatch)
	}
	d, e := a.Decimal(), b.Decimal()
	return d.CmpAbs(e), nil
}

// CmpTotal compares the representation of amounts and returns:
//
//	-1 if a < b
//	-1 if a = b and a.scale > b.scale
//	 0 if a = b and a.scale = b.scale
//	+1 if a = b and a.scale < b.scale
//	+1 if a > b
//
// See also methods [Amount.Cmp], [Amount.CmpAbs].
//
// CmpTotal returns an error if amounts are denominated in different currencies.
func (a Amount) CmpTotal(b Amount) (int, error) {
	if !a.SameCurr(b) {
		return 0, fmt.Errorf("comparing [%v] and [%v]: %w", a, b, errCurrencyMismatch)
	}
	d, e := a.Decimal(), b.Decimal()
	return d.CmpTotal(e), nil
}

// Min returns the smaller amount.
// See also method [Amount.CmpTotal].
//
// Min returns an error if amounts are denominated in different currencies.
func (a Amount) Min(b Amount) (Amount, error) {
	switch c, err := a.CmpTotal(b); {
	case err != nil:
		return Amount{}, err
	case c <= 0: // a <= b
		return a, nil
	default:
		return b, nil
	}
}

// Max returns the larger amount.
// See also method [Amount.CmpTotal].
//
// Max returns an error if amounts are denominated in different currencies.
func (a Amount) Max(b Amount) (Amount, error) {
	switch c, err := a.CmpTotal(b); {
	case err != nil:
		return Amount{}, err
	case c >= 0: // a >= b
		return a, nil
	default:
		return b, nil
	}
}

// Clamp compares amounts and returns:
//
//	min if a < min
//	max if a > max
//	  d otherwise
//
// See also method [Amount.CmpTotal].
//
// Clamp returns an error if:
//   - amounts are denominated in different currencies;
//   - min is greater than max numerically.
func (a Amount) Clamp(min, max Amount) (Amount, error) {
	switch c, err := min.Cmp(max); {
	case err != nil:
		return Amount{}, err
	case c > 0: // min > max
		return Amount{}, fmt.Errorf("clamping %v: invalid range", a)
	}
	switch c, err := min.CmpTotal(max); {
	case err != nil:
		return Amount{}, err
	case c > 0: // min > max
		// Numerically min and max are equal but have different scales.
		// Swaping min and max to ensure total ordering.
		min, max = max, min
	}
	switch c, err := a.CmpTotal(min); {
	case err != nil:
		return Amount{}, err
	case c < 0: // a < min
		return min, nil
	}
	switch c, err := a.CmpTotal(max); {
	case err != nil:
		return Amount{}, err
	case c > 0: // a > max
		return max, nil
	}
	return a, nil
}

// Format implements the [fmt.Formatter] interface.
// The following [format verbs] are available:
//
//	| Verb   | Example     | Description                |
//	| ------ | ----------- | -------------------------- |
//	| %s, %v | USD 5.678   | Currency and amount        |
//	| %q     | "USD 5.678" | Quoted currency and amount |
//	| %f     | 5.678       | Amount                     |
//	| %d     | 568         | Amount in minor units      |
//	| %c     | USD         | Currency                   |
//
// The '-' format flag can be used with all verbs.
// The '+', ' ', '0' format flags can be used with all verbs except %c.
//
// Precision is only supported for the %f verb.
// The default precision is equal to the actual scale of the amount.
//
// [format verbs]: https://pkg.go.dev/fmt#hdr-Printing
// [fmt.Formatter]: https://pkg.go.dev/fmt#Formatter
//
//gocyclo:ignore
func (a Amount) Format(state fmt.State, verb rune) {
	c, d := a.Curr(), a.Decimal()

	// Rescaling
	tzeros := 0
	if verb == 'f' || verb == 'F' || verb == 'd' || verb == 'D' {
		scale := 0
		switch p, ok := state.Precision(); {
		case verb == 'd' || verb == 'D':
			scale = c.Scale()
		case ok:
			scale = p
		case verb == 'f' || verb == 'F':
			scale = d.Scale()
		}
		scale = max(scale, c.Scale())
		switch {
		case scale < d.Scale():
			d = d.Round(scale)
		case scale > d.Scale():
			tzeros = scale - d.Scale()
		}
	}

	// Integer and fractional digits
	intdigs, fracdigs := 0, 0
	switch aprec := d.Prec(); verb {
	case 'c', 'C':
		// skip
	case 'd', 'D':
		intdigs = aprec
		if d.IsZero() {
			intdigs++ // leading 0
		}
	default:
		fracdigs = d.Scale()
		if aprec > fracdigs {
			intdigs = aprec - fracdigs
		}
		if d.WithinOne() {
			intdigs++ // leading 0
		}
	}

	// Decimal point
	dpoint := 0
	if fracdigs > 0 || tzeros > 0 {
		dpoint = 1
	}

	// Arithmetic sign
	rsign := 0
	if verb != 'c' && verb != 'C' && (d.IsNeg() || state.Flag('+') || state.Flag(' ')) {
		rsign = 1
	}

	// Currency code and delimiter
	curr, currsyms, currdel := "", 0, 0
	switch verb {
	case 'f', 'F', 'd', 'D':
		// skip
	case 'c', 'C':
		curr = c.Code()
		currsyms = len(curr)
	default:
		curr = c.Code()
		currsyms = len(curr)
		currdel = 1
	}

	// Opening and closing quotes
	lquote, tquote := 0, 0
	if verb == 'q' || verb == 'Q' {
		lquote, tquote = 1, 1
	}

	// Calculating padding
	width := lquote + currsyms + currdel + rsign + intdigs + dpoint + fracdigs + tzeros + tquote
	lspaces, lzeros, tspaces := 0, 0, 0
	if w, ok := state.Width(); ok && w > width {
		switch {
		case state.Flag('-'):
			tspaces = w - width
		case state.Flag('0') && verb != 'c' && verb != 'C':
			lzeros = w - width
		default:
			lspaces = w - width
		}
		width = w
	}

	buf := make([]byte, width)
	pos := width - 1

	// Trailing spaces
	for i := 0; i < tspaces; i++ {
		buf[pos] = ' '
		pos--
	}

	// Closing quote
	if tquote > 0 {
		buf[pos] = '"'
		pos--
	}

	// Trailing zeros
	for i := 0; i < tzeros; i++ {
		buf[pos] = '0'
		pos--
	}

	// Fractional digits
	coef := d.Coef()
	for i := 0; i < fracdigs; i++ {
		buf[pos] = byte(coef%10) + '0'
		pos--
		coef /= 10
	}

	// Decimal point
	if dpoint > 0 {
		buf[pos] = '.'
		pos--
	}

	// Integer digits
	for i := 0; i < intdigs; i++ {
		buf[pos] = byte(coef%10) + '0'
		pos--
		coef /= 10
	}

	// Leading zeros
	for i := 0; i < lzeros; i++ {
		buf[pos] = '0'
		pos--
	}

	// Arithmetic sign
	if rsign > 0 {
		if d.IsNeg() {
			buf[pos] = '-'
		} else if state.Flag(' ') {
			buf[pos] = ' '
		} else {
			buf[pos] = '+'
		}
		pos--
	}

	// Currency delimiter
	if currdel > 0 {
		buf[pos] = ' '
		pos--
	}

	// Currency code
	for i := currsyms; i > 0; i-- {
		buf[pos] = curr[i-1]
		pos--
	}

	// Opening quote
	if lquote > 0 {
		buf[pos] = '"'
		pos--
	}

	// Leading spaces
	for i := 0; i < lspaces; i++ {
		buf[pos] = ' '
		pos--
	}

	// Writing result
	//nolint:errcheck
	switch verb {
	case 'q', 'Q', 's', 'S', 'v', 'V', 'f', 'F', 'd', 'D', 'c', 'C':
		state.Write(buf)
	default:
		state.Write([]byte("%!"))
		state.Write([]byte{byte(verb)})
		state.Write([]byte("(money.Amount="))
		state.Write(buf)
		state.Write([]byte(")"))
	}
}
