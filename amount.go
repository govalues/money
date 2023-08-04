package money

import (
	"errors"
	"fmt"
	"math"

	"github.com/govalues/decimal"
)

var errCurrencyMismatch = errors.New("currency mismatch")

// Amount type represents a monetary amount.
// The zero value corresponds to "XXX 0", where XXX indicates an unknown currency.
// This type is designed to be safe for concurrent use by multiple goroutines.
type Amount struct {
	curr  Currency        // an ISO 4217 currency code
	value decimal.Decimal // the monetary value
}

func newAmountUnsafe(curr Currency, amount decimal.Decimal) Amount {
	return Amount{curr: curr, value: amount}
}

// NewAmount returns a new amount with the specified currency and value.
// If the scale of the amount is less than the scale of the currency, the result
// will be zero-padded to the right.
func NewAmount(curr Currency, amount decimal.Decimal) (Amount, error) {
	amount, err := amount.Pad(curr.Scale())
	if err != nil {
		return Amount{}, fmt.Errorf("padding: %w", err)
	}
	return newAmountUnsafe(curr, amount), nil
}

// MustNewAmount is like [NewAmount] but panics if the amount cannot be created.
// This function simplifies safe initialization of global variables holding amounts.
func MustNewAmount(curr Currency, amount decimal.Decimal) Amount {
	a, err := NewAmount(curr, amount)
	if err != nil {
		panic(fmt.Sprintf("NewAmount(%v, %v) failed: %v", curr, amount, err))
	}
	return a
}

// ParseAmount converts currency and decimal strings to (possibly rounded) amount.
// If the scale of the amount is less than the scale of the currency, the result
// will be zero-padded to the right.
// See also methods [ParseCurr] and [decimal.Parse].
func ParseAmount(curr, amount string) (Amount, error) {
	c, err := ParseCurr(curr)
	if err != nil {
		return Amount{}, fmt.Errorf("parsing currency: %w", err)
	}
	d, err := decimal.ParseExact(amount, c.Scale())
	if err != nil {
		return Amount{}, fmt.Errorf("parsing amount: %w", err)
	}
	a, err := NewAmount(c, d)
	if err != nil {
		return Amount{}, fmt.Errorf("new amount: %w", err)
	}
	return a, nil
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

// Coef returns the coefficient of the amount.
// See also methods [Amount.Prec] and [Amount.MinorUnits].
func (a Amount) Coef() uint64 {
	return a.value.Coef()
}

// MinorUnits returns the (potentially rounded) amount in minor units of currency.
// If the result cannot be represented as an int64, then false is returned.
// See also method [Amount.RoundToCurr].
func (a Amount) MinorUnits() (m int64, ok bool) {
	coef := a.RoundToCurr().Coef()
	if a.IsNeg() {
		if coef > -math.MinInt64 {
			return 0, false
		}
		return -int64(coef), true
	}
	if coef > math.MaxInt64 {
		return 0, false
	}
	return int64(coef), true
}

// Float64 returns a float64 representation of the amount.
// This conversion may lose data, as float64 has a limited precision
// compared to the decimal type.
func (a Amount) Float64() (f float64, ok bool) {
	return a.value.Float64()
}

// Int64 returns a pair of int64 values representing the integer part b and the
// fractional part f of the amount.
// The relationship can be expressed as a = b + f / 10^scale.
// If the result cannot be accurately represented as a pair of int64 values,
// the method returns false.
func (a Amount) Int64(scale int) (b, f int64, ok bool) {
	return a.value.Int64(scale)
}

// Prec returns the number of digits in the coefficient.
// See also method [Amount.Coef].
func (a Amount) Prec() int {
	return a.value.Prec()
}

// Curr returns the currency of the amount.
func (a Amount) Curr() Currency {
	return a.curr
}

// Sign returns:
//
//	-1 if a < 0
//	 0 if a == 0
//	+1 if a > 0
func (a Amount) Sign() int {
	return a.value.Sign()
}

// IsNeg returns:
//
//	true  if a < 0
//	false otherwise
func (a Amount) IsNeg() bool {
	return a.value.IsNeg()
}

// IsPos returns:
//
//	true  if a > 0
//	false otherwise
func (a Amount) IsPos() bool {
	return a.value.IsPos()
}

// Abs returns the absolute value of the amount.
func (a Amount) Abs() Amount {
	d := a.value
	return newAmountUnsafe(a.Curr(), d.Abs())
}

// Neg returns the amount with the opposite sign.
func (a Amount) Neg() Amount {
	d := a.value
	return newAmountUnsafe(a.Curr(), d.Neg())
}

// CopySign returns the amount with the same sign as amount b.
// If amount b is zero, the sign of the result remains unchanged.
func (a Amount) CopySign(b Amount) Amount {
	d, e := a.value, b.value
	return newAmountUnsafe(a.Curr(), d.CopySign(e))
}

// Scale returns the number of digits after the decimal point.
func (a Amount) Scale() int {
	return a.value.Scale()
}

// IsZero returns:
//
//	true  if a == 0
//	false otherwise
func (a Amount) IsZero() bool {
	return a.value.IsZero()
}

// IsOne returns:
//
//	true  if a == -1 || a == 1
//	false otherwise
func (a Amount) IsOne() bool {
	return a.value.IsOne()
}

// IsInt returns true if the fractional part of the amount is equal to zero.
func (a Amount) IsInt() bool {
	return a.value.IsInt()
}

// WithinOne returns:
//
//	true  if -1 < a < 1
//	false otherwise
func (a Amount) WithinOne() bool {
	return a.value.WithinOne()
}

// Add returns the (possibly rounded) sum of amounts a and b.
//
// Add returns an error if:
//   - amounts are denominated in different currencies.
//   - the integer part of the result exceeds the maximum precision allowed by the currency.
//     This limit is calculated as ([decimal.MaxPrec] - a.Curr().Scale()).
//     For example, when dealing with US Dollars, Add will return an error if the integer
//     part of the result has more than 17 digits (19 - 2 = 17).
func (a Amount) Add(b Amount) (Amount, error) {
	if !a.SameCurr(b) {
		return Amount{}, fmt.Errorf("%q + %q: %w", a, b, errCurrencyMismatch)
	}
	d, e := a.value, b.value
	f, err := d.AddExact(e, a.Curr().Scale())
	if err != nil {
		return Amount{}, fmt.Errorf("%q + %q: %w", a, b, err)
	}
	return NewAmount(a.Curr(), f)
}

// Sub returns the (possibly rounded) difference of amounts a and b.
//
// Sub returns an error if:
//   - amounts are denominated in different currencies.
//   - the integer part of the result exceeds the maximum precision allowed by the currency.
//     This limit is calculated as ([decimal.MaxPrec] - a.Curr().Scale()).
//     For example, when dealing with US Dollars, Sub will return an error if the integer
//     part of the result has more than 17 digits (19 - 2 = 17).
func (a Amount) Sub(b Amount) (Amount, error) {
	if !a.SameCurr(b) {
		return Amount{}, fmt.Errorf("%q - %q: %w", a, b, errCurrencyMismatch)
	}
	d, e := a.value, b.value
	f, err := d.SubExact(e, a.Curr().Scale())
	if err != nil {
		return Amount{}, fmt.Errorf("%q - %q: %w", a, b, err)
	}
	return NewAmount(a.Curr(), f)
}

// FMA returns the (possibly rounded) [fused multiply-addition] of amounts a, b, and factor e.
// It computes a * e + b without any intermeddiate rounding.
// This method is useful for improving the accuracy and performance of algorithms
// that involve the accumulation of products, such as daily interest accrual.
//
// FMA returns an error if:
//   - amounts are denominated in different currencies.
//   - the integer part of the result exceeds the maximum precision allowed by the currency.
//     This limit is calculated as ([decimal.MaxPrec] - a.Curr().Scale()).
//     For example, when dealing with US Dollars, FMA will return an error if the integer
//     part of the result has more than 17 digits (19 - 2 = 17).
//
// [fused multiply-addition]: https://en.wikipedia.org/wiki/Multiply%E2%80%93accumulate_operation#Fused_multiply%E2%80%93add
func (a Amount) FMA(e decimal.Decimal, b Amount) (Amount, error) {
	if !a.SameCurr(b) {
		return Amount{}, fmt.Errorf("%q * %v + %q: %w", a, e, b, errCurrencyMismatch)
	}
	d, f := a.value, b.value
	d, err := d.FMAExact(e, f, a.Curr().Scale())
	if err != nil {
		return Amount{}, fmt.Errorf("%q * %v + %q: %w", a, e, b, err)
	}
	return NewAmount(a.Curr(), d)
}

// Mul returns the (possibly rounded) product of amount a and factor e.
//
// Mul returns an error if the integer part of the result exceeds the maximum precision
// allowed by the currency.
// This limit is calculated as ([decimal.MaxPrec] - a.Curr().Scale()).
// For example, when dealing with US Dollars, Mul will return an error if the integer
// part of the result has more than 17 digits (19 - 2 = 17).
func (a Amount) Mul(e decimal.Decimal) (Amount, error) {
	d := a.value
	f, err := d.MulExact(e, a.Curr().Scale())
	if err != nil {
		return Amount{}, fmt.Errorf("%q * %v: %w", a, e, err)
	}
	return NewAmount(a.Curr(), f)
}

// Quo returns the (possibly rounded) quotient of amount a and divisor e.
//
// Quo returns an error if:
//   - divisor is zero.
//   - the integer part of the result exceeds the maximum precision allowed by the currency.
//     This limit is calculated as ([decimal.MaxPrec] - a.Curr().Scale()).
//     For example, when dealing with US Dollars, Quo will return an error if the integer
//     part of the result has more than 17 digits (19 - 2 = 17).
func (a Amount) Quo(e decimal.Decimal) (Amount, error) {
	d := a.value
	d, err := d.QuoExact(e, a.Curr().Scale())
	if err != nil {
		return Amount{}, fmt.Errorf("%q / %v: %w", a, e, err)
	}
	return NewAmount(a.Curr(), d)
}

// Rat returns the (possibly rounded) ratio between amounts a and b.
// This method is particularly useful for calculating exchange rates between
// two currencies or determining percentages within a single currency.
//
// Rat returns an error if:
//   - divisor is zero.
//   - the integer part of the result exceeds the maximum precision.
//     This limit is set to [decimal.MaxPrec] digits.
func (a Amount) Rat(b Amount) (decimal.Decimal, error) {
	d, e := a.value, b.value
	f, err := d.Quo(e)
	if err != nil {
		return decimal.Decimal{}, fmt.Errorf("%q / %q: %w", a, b, err)
	}
	return f, nil
}

// Split returns a slice of amounts that sum up to the original amount,
// ensuring the parts are as equal as possible.
// If the original amount cannot be divided equally among the specified number of parts, the
// remainder is distributed among the first parts of the slice.
//
// Split returns an error if the number of parts is not a positive integer.
func (a Amount) Split(parts int) ([]Amount, error) {
	// Parts
	div, err := decimal.New(int64(parts), 0)
	if err != nil {
		return nil, err
	}
	if !div.IsPos() {
		return nil, fmt.Errorf("number of parts must be positive")
	}

	// Quotient
	quo, err := a.Quo(div)
	if err != nil {
		return nil, err
	}
	quo = quo.Trunc(a.Scale())

	// Reminder
	rem, err := quo.Mul(div)
	if err != nil {
		return nil, err
	}
	rem, err = a.Sub(rem)
	if err != nil {
		return nil, err
	}
	ulp := rem.ULP().CopySign(rem)

	// Distribute remainder
	res := make([]Amount, parts)
	for i := 0; i < parts; i++ {
		res[i] = quo
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

// One returns an amount with a value of 1, having the same currency and scale as amount a.
// See also method [Amount.ULP].
func (a Amount) One() Amount {
	d := a.value
	return newAmountUnsafe(a.Curr(), d.One())
}

// Zero returns an amount with a value of 0, having the same currency and scale as amount a.
func (a Amount) Zero() Amount {
	d := a.value
	return newAmountUnsafe(a.Curr(), d.Zero())
}

// ULP (Unit in the Last Place) returns the smallest representable positive difference
// between two amounts with the same scale as amount a.
// It can be useful for implementing rounding and comparison algorithms.
// See also method [Amount.One].
func (a Amount) ULP() Amount {
	d := a.value
	return newAmountUnsafe(a.Curr(), d.ULP())
}

// Ceil returns an amount rounded up to the specified number of digits after
// the decimal point.
// If the given scale is less than the scale of the currency,
// the amount will be rounded up to the scale of the currency instead.
// See also method [Amount.CeilToCurr].
func (a Amount) Ceil(scale int) Amount {
	if scale < a.Curr().Scale() {
		scale = a.Curr().Scale()
	}
	d := a.value
	return newAmountUnsafe(a.Curr(), d.Ceil(scale))
}

// CeilToCurr returns an amount rounded up to the scale of its currency.
// See also method [Amount.SameScaleAsCurr].
func (a Amount) CeilToCurr() Amount {
	return a.Ceil(a.Curr().Scale())
}

// Floor returns an amount rounded down to the specified number of digits after
// the decimal point.
// If the given scale is less than the scale of the currency,
// the amount will be rounded down to the scale of the currency instead.
// See also method [Amount.FloorToCurr].
func (a Amount) Floor(scale int) Amount {
	if scale < a.Curr().Scale() {
		scale = a.Curr().Scale()
	}
	d := a.value
	return newAmountUnsafe(a.Curr(), d.Floor(scale))
}

// FloorToCurr returns an amount rounded down to the scale of its currency.
// See also method [Amount.SameScaleAsCurr].
func (a Amount) FloorToCurr() Amount {
	return a.Floor(a.Curr().Scale())
}

// Trunc returns an amount truncated to the specified number of digits after
// the decimal point.
// If the given scale is less than the scale of the currency,
// the amount will be truncated to the scale of the currency instead.
// See also method [Amount.TruncToCurr].
func (a Amount) Trunc(scale int) Amount {
	if scale < a.Curr().Scale() {
		scale = a.Curr().Scale()
	}
	d := a.value
	return newAmountUnsafe(a.Curr(), d.Trunc(scale))
}

// TruncToCurr returns an amount truncated to the scale of its currency.
// See also method [Amount.SameScaleAsCurr].
func (a Amount) TruncToCurr() Amount {
	return a.Trunc(a.Curr().Scale())
}

// Round returns an amount rounded to the specified number of digits after
// the decimal point.
// If the given scale is less than the scale of the currency,
// the amount will be rounded to the currency's scale instead.
// See also method [Amount.RoundToCurr].
func (a Amount) Round(scale int) Amount {
	if scale < a.Curr().Scale() {
		scale = a.Curr().Scale()
	}
	d := a.value
	return newAmountUnsafe(a.Curr(), d.Round(scale))
}

// RoundToCurr returns an amount rounded to the scale of its currency.
// See also method [Amount.SameScaleAsCurr].
func (a Amount) RoundToCurr() Amount {
	return a.Round(a.Curr().Scale())
}

// Rescale returns an amount rounded or zero-padded to the given number of digits
// after the decimal point.
// If the specified scale is less than the scale of the currency,
// the amount will be rounded to the currency's scale instead.
//
// Rescale returns an error if the integer part of the result exceeds
// the maximum precision, calculated as ([decimal.MaxPrec] - scale).
func (a Amount) Rescale(scale int) (Amount, error) {
	if scale < a.Curr().Scale() {
		scale = a.Curr().Scale()
	}
	d := a.value
	d, err := d.Rescale(scale)
	if err != nil {
		return Amount{}, fmt.Errorf("rescaling %q to %v decimal places: %w", a, scale, err)
	}
	return NewAmount(a.Curr(), d)
}

// Quantize returns an amount rounded to the same scale as amount b.
// The sign and value of amount b are ignored.
// See also method [Amount.Rescale].
//
// Quantize returns an error if:
//   - amounts are denominated in different currencies.
//   - the integer part of the result exceeds the maximum precision,
//     calculated as ([decimal.MaxPrec] - b.Scale()).
func (a Amount) Quantize(b Amount) (Amount, error) {
	if !a.SameCurr(b) {
		return Amount{}, errCurrencyMismatch
	}
	return a.Rescale(b.Scale())
}

// Trim returns an amount with trailing zeros removed up to the given scale.
// If the given scale is less than the scale of the currency,
// the zeros will be removed up to the scale of the currency instead.
// See also method [Amount.TrimToCurr].
func (a Amount) Trim(scale int) Amount {
	if scale < a.Curr().Scale() {
		scale = a.Curr().Scale()
	}
	d := a.value
	return newAmountUnsafe(a.Curr(), d.Trim(scale))
}

// TrimToCurr returns an amount with trailing zeros removed up the scale of its currency.
func (a Amount) TrimToCurr() Amount {
	return a.Trim(a.Curr().Scale())
}

// SameCurr returns true if amounts are denominated in the same currency.
// See also method [Amount.Curr].
func (a Amount) SameCurr(b Amount) bool {
	return a.Curr() == b.Curr()
}

// SameScale returns true if amounts have the same scale.
// See also method [Amount.Scale].
func (a Amount) SameScale(b Amount) bool {
	return a.Scale() == b.Scale()
}

// SameScaleAsCurr returns true if the scale of the amount matches the scale of its currency.
// See also method [Amount.RoundToCurr].
func (a Amount) SameScaleAsCurr() bool {
	return a.Scale() == a.Curr().Scale()
}

// String implements the [fmt.Stringer] interface and returns a string
// representation of an amount.
// See also methods [Currency.String] and [Decimal.String].
//
// [fmt.Stringer]: https://pkg.go.dev/fmt#Stringer
// [Decimal.String]: https://pkg.go.dev/github.com/govalues/decimal#Decimal.String
func (a Amount) String() string {
	return a.Curr().String() + " " + a.value.String()
}

// Cmp compares amounts numerically and returns:
//
//	-1 if a < b
//	 0 if a == b
//	+1 if a > b
//
// Cmp returns an error if amounts are denominated in different currencies.
func (a Amount) Cmp(b Amount) (int, error) {
	if !a.SameCurr(b) {
		return 0, errCurrencyMismatch
	}
	d, e := a.value, b.value
	return d.Cmp(e), nil
}

// CmpTotal compares the representation of amounts and returns:
//
//	-1 if a < b
//	-1 if a == b && a.scale >  b.scale
//	 0 if a == b && a.scale == b.scale
//	+1 if a == b && a.scale <  b.scale
//	+1 if a > b
//
// CmpTotal returns an error if amounts are denominated in different currencies.
// See also method [Amount.Cmp].
func (a Amount) CmpTotal(b Amount) (int, error) {
	if !a.SameCurr(b) {
		return 0, errCurrencyMismatch
	}
	d, e := a.value, b.value
	return d.CmpTotal(e), nil
}

// Min returns the smaller amount.
// See also method [Amount.CmpTotal].
//
// Min returns an error if amounts are denominated in different currencies.
func (a Amount) Min(b Amount) (Amount, error) {
	switch s, err := a.CmpTotal(b); {
	case err != nil:
		return Amount{}, err
	case s <= 0:
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
	switch s, err := a.CmpTotal(b); {
	case err != nil:
		return Amount{}, err
	case s >= 0:
		return a, nil
	default:
		return b, nil
	}
}

// Format implements the [fmt.Formatter] interface.
// The following [format verbs] are available:
//
//	%s, %v: USD -123.456
//	%q:    "USD -123.456"
//	%f:    -123.46
//	%d:    -12346
//	%c:     USD
//
// The '-' format flag can be used with all verbs.
// The '+', ' ', '0' format flags can be used with all verbs except %c.
//
// Precision is only supported for the %f verb.
// The default precision is equal to the scale of the currency.
//
// [format verbs]: https://pkg.go.dev/fmt#hdr-Printing
// [fmt.Formatter]: https://pkg.go.dev/fmt#Formatter
func (a Amount) Format(state fmt.State, verb rune) {
	// Rescaling
	tzeroes := 0
	if verb == 'f' || verb == 'F' || verb == 'd' || verb == 'D' {
		scale := 0
		switch p, ok := state.Precision(); {
		case verb == 'd' || verb == 'D':
			scale = a.Curr().Scale()
		case ok:
			scale = p
		case verb == 'f' || verb == 'F':
			scale = a.Curr().Scale()
		}
		switch {
		case scale < a.Scale():
			a = a.Round(scale)
		case scale > a.Scale():
			tzeroes = scale - a.Scale()
		}
	}

	// Integer and fractional digits
	intdigs, fracdigs := 0, 0
	switch aprec := a.Prec(); verb {
	case 'c', 'C':
		// skip
	case 'd', 'D':
		intdigs = aprec
		if a.IsZero() {
			intdigs++ // leading 0
		}
	default:
		fracdigs = a.Scale()
		if aprec > fracdigs {
			intdigs = aprec - fracdigs
		}
		if a.WithinOne() {
			intdigs++ // leading 0
		}
	}

	// Decimal point
	dpoint := 0
	if fracdigs > 0 || tzeroes > 0 {
		dpoint = 1
	}

	// Arithmetic sign
	rsign := 0
	if verb != 'c' && verb != 'C' && (a.IsNeg() || state.Flag('+') || state.Flag(' ')) {
		rsign = 1
	}

	// Currency symbols
	curr := ""
	switch verb {
	case 'f', 'F', 'd', 'D':
		// skip
	case 'c', 'C':
		curr = a.Curr().String()
	default:
		curr = a.Curr().String() + " "
	}
	currlen := len(curr)

	// Quotes
	lquote, tquote := 0, 0
	if verb == 'q' || verb == 'Q' {
		lquote, tquote = 1, 1
	}

	// Padding
	width := lquote + rsign + intdigs + dpoint + fracdigs + tzeroes + currlen + tquote
	lspaces, lzeroes, tspaces := 0, 0, 0
	if w, ok := state.Width(); ok && w > width {
		switch {
		case state.Flag('-'):
			tspaces = w - width
		case state.Flag('0') && verb != 'c' && verb != 'C':
			lzeroes = w - width
		default:
			lspaces = w - width
		}
		width = w
	}

	// Writing buffer
	buf := make([]byte, width)
	pos := width - 1
	for i := 0; i < tspaces; i++ {
		buf[pos] = ' '
		pos--
	}
	if tquote > 0 {
		buf[pos] = '"'
		pos--
	}
	for i := 0; i < tzeroes; i++ {
		buf[pos] = '0'
		pos--
	}
	coef := a.Coef()
	for i := 0; i < fracdigs; i++ {
		buf[pos] = byte(coef%10) + '0'
		pos--
		coef /= 10
	}
	if dpoint > 0 {
		buf[pos] = '.'
		pos--
	}
	for i := 0; i < intdigs; i++ {
		buf[pos] = byte(coef%10) + '0'
		pos--
		coef /= 10
	}
	for i := 0; i < lzeroes; i++ {
		buf[pos] = '0'
		pos--
	}
	if rsign > 0 {
		if a.IsNeg() {
			buf[pos] = '-'
		} else if state.Flag(' ') {
			buf[pos] = ' '
		} else {
			buf[pos] = '+'
		}
		pos--
	}
	for i := currlen; i > 0; i-- {
		buf[pos] = curr[i-1]
		pos--
	}
	if lquote > 0 {
		buf[pos] = '"'
		pos--
	}
	for i := 0; i < lspaces; i++ {
		buf[pos] = ' '
		pos--
	}

	// Writing result
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
