package money

import (
	"fmt"
	"math"
	"strconv"

	"github.com/govalues/decimal"
)

var errRateOverflow = fmt.Errorf("rate overflow")

// ExchangeRate represents a unidirectional exchange rate between two currencies.
// The zero value corresponds to an exchange rate of "XXX/XXX 0", where [XXX] indicates
// an unknown currency.
// This type is designed to be safe for concurrent use by multiple goroutines.
type ExchangeRate struct {
	base  Currency        // currency being exchanged
	quote Currency        // currency being obtained in exchange for the base currency
	value decimal.Decimal // how many units of the quote currency are needed to exchange for 1 unit of the base currency
}

// newExchRateUnsafe creates a new rate without checking the sign and the scale.
// Use it only if you are absolutely sure that the arguments are valid.
func newExchRateUnsafe(b, q Currency, d decimal.Decimal) ExchangeRate {
	return ExchangeRate{base: b, quote: q, value: d}
}

// newExchRateSafe creates a new rate and checks the sign and the scale.
func newExchRateSafe(b, q Currency, d decimal.Decimal) (ExchangeRate, error) {
	if d.IsZero() {
		return ExchangeRate{}, fmt.Errorf("exchange rate cannot be 0")
	}
	if d.IsNeg() {
		return ExchangeRate{}, fmt.Errorf("exchange rate must be positive")
	}
	if b == q && !d.IsOne() {
		return ExchangeRate{}, fmt.Errorf("exchange rate between identical currencies must be equal to 1")
	}
	if d.Scale() < q.Scale() {
		d = d.Pad(q.Scale())
		if d.Scale() < q.Scale() {
			return ExchangeRate{}, fmt.Errorf("padding exchange rate: %w", errRateOverflow)
		}
	}
	return newExchRateUnsafe(b, q, d), nil
}

// NewExchRate returns a rate equal to coef / 10^scale.
// If the scale of the rate is less than the scale of the quote currency, the
// result will be zero-padded to the right.
//
// NewExchRate returns an error if:
//   - the currency codes are not valid;
//   - the coefficient is 0 or negative
//   - the scale is negative or greater than [decimal.MaxScale];
//   - the integer part of the result has more than
//     ([decimal.MaxPrec] - [Currency.Scale]) digits.
//     For example, when the quote currency is US Dollars, NewAmount will return an error
//     if the integer part of the result has more than 17 digits (19 - 2 = 17).
func NewExchRate(base, quote string, coef int64, scale int) (ExchangeRate, error) {
	// Currency
	b, err := ParseCurr(base)
	if err != nil {
		return ExchangeRate{}, fmt.Errorf("parsing currency: %w", err)
	}
	q, err := ParseCurr(quote)
	if err != nil {
		return ExchangeRate{}, fmt.Errorf("parsing currency: %w", err)
	}
	// Decimal
	d, err := decimal.New(coef, scale)
	if err != nil {
		return ExchangeRate{}, fmt.Errorf("converting coefficient: %w", err)
	}
	// Rate
	r, err := newExchRateSafe(b, q, d)
	if err != nil {
		return ExchangeRate{}, fmt.Errorf("converting coefficient: %w", err)
	}
	return r, nil
}

// MustNewExchRate is like [NewExchRate] but panics if the rate cannot be constructed.
// It simplifies safe initialization of global variables holding rates.
func MustNewExchRate(base, quote string, coef int64, scale int) ExchangeRate {
	r, err := NewExchRate(base, quote, coef, scale)
	if err != nil {
		panic(fmt.Sprintf("NewExchRate(%q, %q, %v, %v) failed: %v", base, quote, coef, scale, err))
	}
	return r
}

// NewExchRateFromDecimal returns a rate with the specified currencies and value.
// If the scale of the rate is less than the scale of quote currency, the result
// will be zero-padded to the right. See also method [ExchangeRate.Decimal].
//
// NewExchRateFromDecimal returns an error if the integer part of the result has more than
// ([decimal.MaxPrec] - [Currency.Scale]) digits.
// For example, when the quote currency is US Dollars, NewExchRateFromDecimal will return an error if
// the integer part of the result has more than 17 digits (19 - 2 = 17).
func NewExchRateFromDecimal(base, quote Currency, rate decimal.Decimal) (ExchangeRate, error) {
	return newExchRateSafe(base, quote, rate)
}

// NewExchRateFromInt64 converts a pair of integers, representing the whole and
// fractional parts, to a (possibly rounded) rate equal to whole + frac / 10^scale.
// NewExchRateFromInt64 deletes trailing zeros up to the scale of the quote currency.
// This method is useful for converting rates from [protobuf] format.
// See also method [ExchangeRate.Int64].
//
// NewExchRateFromInt64 returns an error if:
//   - the currency codes are not valid;
//   - the whole part is negative;
//   - the fractional part is negative;
//   - the scale is negative or greater than [decimal.MaxScale];
//   - frac / 10^scale is not within the range (-1, 1);
//   - the integer part of the result has more than
//     ([decimal.MaxPrec] - [Currency.Scale]) digits.
//     For example, when the quote currency is US Dollars, NewAmountFromInt64 will return
//     an error if the integer part of the result has more than 17 digits (19 - 2 = 17).
//
// [protobuf]: https://github.com/googleapis/googleapis/blob/master/google/type/money.proto
func NewExchRateFromInt64(base, quote string, whole, frac int64, scale int) (ExchangeRate, error) {
	// Currency
	b, err := ParseCurr(base)
	if err != nil {
		return ExchangeRate{}, fmt.Errorf("parsing currency: %w", err)
	}
	q, err := ParseCurr(quote)
	if err != nil {
		return ExchangeRate{}, fmt.Errorf("parsing currency: %w", err)
	}
	// Whole
	d, err := decimal.New(whole, 0)
	if err != nil {
		return ExchangeRate{}, fmt.Errorf("converting integers: %w", err)
	}
	// Fraction
	f, err := decimal.New(frac, scale)
	if err != nil {
		return ExchangeRate{}, fmt.Errorf("converting integers: %w", err)
	}
	if !f.IsZero() {
		if !d.IsZero() && d.Sign() != f.Sign() {
			return ExchangeRate{}, fmt.Errorf("converting integers: inconsistent signs")
		}
		if !f.WithinOne() {
			return ExchangeRate{}, fmt.Errorf("converting integers: inconsistent fraction")
		}
		f = f.Trim(q.Scale())
		d, err = d.AddExact(f, q.Scale())
		if err != nil {
			return ExchangeRate{}, fmt.Errorf("converting integers: %w", err)
		}
	}
	// Amount
	return newExchRateSafe(b, q, d)
}

// NewExchRateFromFloat64 converts a float to a (possibly rounded) rate.
// See also method [ExchangeRate.Float64].
//
// NewExchRateFromFloat64 returns an error if:
//   - the currency codes are not valid;
//   - the float is a 0 or negative;
//   - the float is a special value (NaN or Inf);
//   - the integer part of the result has more than
//     ([decimal.MaxPrec] - [Currency.Scale]) digits.
//     For example, when the quote currency is US Dollars, NewExchRateFromFloat64 will
//     return an error if the integer part of the result has more than 17
//     digits (19 - 2 = 17).
func NewExchRateFromFloat64(base, quote string, rate float64) (ExchangeRate, error) {
	// Float
	if math.IsNaN(rate) || math.IsInf(rate, 0) {
		return ExchangeRate{}, fmt.Errorf("converting float: special value %v", rate)
	}
	s := strconv.FormatFloat(rate, 'f', -1, 64)
	// Rate
	r, err := ParseExchRate(base, quote, s)
	if err != nil {
		return ExchangeRate{}, fmt.Errorf("converting float: %w", err)
	}
	return r, nil
}

// ParseExchRate converts currency and decimal strings to a (possibly rounded) rate.
// If the scale of the rate is less than the scale of the quote currency,
// the result will be zero-padded to the right.
// See also constructors [ParseCurr] and [decimal.Parse].
func ParseExchRate(base, quote, rate string) (ExchangeRate, error) {
	// Currency
	b, err := ParseCurr(base)
	if err != nil {
		return ExchangeRate{}, fmt.Errorf("parsing base currency: %w", err)
	}
	q, err := ParseCurr(quote)
	if err != nil {
		return ExchangeRate{}, fmt.Errorf("parsing quote currency: %w", err)
	}
	// Decimal
	d, err := decimal.ParseExact(rate, q.Scale())
	if err != nil {
		return ExchangeRate{}, fmt.Errorf("parsing exchange rate: %w", err)
	}
	// Rate
	r, err := newExchRateSafe(b, q, d)
	if err != nil {
		return ExchangeRate{}, fmt.Errorf("parsing exchange rate: %w", err)
	}
	return r, nil
}

// MustParseExchRate is like [ParseExchRate] but panics if any of the strings cannot be parsed.
// It simplifies safe initialization of global variables holding exchange rates.
func MustParseExchRate(base, quote, rate string) ExchangeRate {
	r, err := ParseExchRate(base, quote, rate)
	if err != nil {
		panic(fmt.Sprintf("ParseExchRate(%q, %q, %q) failed: %v", base, quote, rate, err))
	}
	return r
}

// Base returns the currency being exchanged.
func (r ExchangeRate) Base() Currency {
	return r.base
}

// Quote returns the currency being obtained in exchange for the base currency.
func (r ExchangeRate) Quote() Currency {
	return r.quote
}

// Decimal returns the decimal representation of the rate.
// It is equal to the number of units of the quote currency needed
// to exchange for 1 unit of the base currency.
func (r ExchangeRate) Decimal() decimal.Decimal {
	return r.value
}

// Float64 returns the nearest binary floating-point number rounded
// using [rounding half to even] (banker's rounding).
// See also constructor [NewExchRateFromFloat64].
//
// This conversion may lose data, as float64 has a smaller precision
// than the decimal type.
//
// [rounding half to even]: https://en.wikipedia.org/wiki/Rounding#Rounding_half_to_even
func (r ExchangeRate) Float64() (f float64, ok bool) {
	return r.Decimal().Float64()
}

// Int64 returns a pair of integers representing the whole and (possibly
// rounded) fractional parts of the rate.
// If given scale is greater than the scale of the rate, then the fractional part
// is zero-padded to the right.
// If given scale is smaller than the scale of the rate, then the fractional part
// is rounded using [rounding half to even] (banker's rounding).
// The relationship between the rate and the returned values can be expressed
// as r = whole + frac / 10^scale.
// This method is useful for converting rates to [protobuf] format.
// See also constructor [NewExchRateFromInt64].
//
// Int64 returns false if the result cannot be represented as a pair of int64 values.
//
// [rounding half to even]: https://en.wikipedia.org/wiki/Rounding#Rounding_half_to_even
// [protobuf]: https://github.com/googleapis/googleapis/blob/master/google/type/money.proto
func (r ExchangeRate) Int64(scale int) (whole, frac int64, ok bool) {
	return r.Decimal().Int64(scale)
}

// CanConv returns true if [ExchangeRate.Conv] can be used to convert the given amount.
func (r ExchangeRate) CanConv(b Amount) bool {
	return r.Base() == b.Curr() &&
		r.Base() != XXX &&
		r.Quote() != XXX &&
		r.IsPos()
}

// Conv returns a (possibly rounded) amount converted from the base currency to
// the quote currency.
// See also method [ExchangeRate.CanConv].
//
// Conv returns an error if:
//   - the base currency of the exchange rate does not match the currency of the given amount.
//   - the integer part of the result has more than
//     ([decimal.MaxPrec] - [Currency.Scale]) digits.
//     For example, when converting to US Dollars, Conv will return an error
//     if the integer part of the result has more than 17 digits (19 - 2 = 17).
func (r ExchangeRate) Conv(b Amount) (Amount, error) {
	c, err := r.conv(b)
	if err != nil {
		return Amount{}, fmt.Errorf("converting [%v] to [%v]: %w", b, r.Quote(), err)
	}
	return c, nil
}

func (r ExchangeRate) conv(b Amount) (Amount, error) {
	if !r.CanConv(b) {
		return Amount{}, errCurrencyMismatch
	}
	q, d, e := r.Quote(), r.Decimal(), b.Decimal()
	d, err := d.MulExact(e, q.Scale())
	if err != nil {
		return Amount{}, err
	}
	return newAmountSafe(q, d)
}

// Mul returns an exchange rate with the same base and quote currencies,
// but with the rate multiplied by a factor.
//
// Mul returns an error if:
//   - the result is 0 or negative;
//   - the integer part of the result has more than
//     ([decimal.MaxPrec] - [Currency.Scale]) digits.
//     For example, when the quote currency is US Dollars, Mul will return an error
//     if the integer part of the result has more than 17 digits (19 - 2 = 17).
func (r ExchangeRate) Mul(e decimal.Decimal) (ExchangeRate, error) {
	q, err := r.mul(e)
	if err != nil {
		return ExchangeRate{}, fmt.Errorf("computing [%v * %v]: %w", r, e, err)
	}
	return q, nil
}

func (r ExchangeRate) mul(e decimal.Decimal) (ExchangeRate, error) {
	b, q, d := r.Base(), r.Quote(), r.Decimal()
	d, err := d.MulExact(e, q.Scale())
	if err != nil {
		return ExchangeRate{}, err
	}
	return newExchRateSafe(b, q, d)
}

// Inv returns the inverse of the exchange rate.
//
// Inv returns an error if:
//   - the rate is 0;
//   - the inverse of the rate is 0;
//   - the integer part of the result has more than
//     ([decimal.MaxPrec] - [Currency.Scale]) digits.
//     For example, when the base currency is US Dollars, Inv will return an error
//     if the integer part of the result has more than 17 digits (19 - 2 = 17).
func (r ExchangeRate) Inv() (ExchangeRate, error) {
	q, err := r.inv()
	if err != nil {
		return ExchangeRate{}, fmt.Errorf("inverting %v: %w", r, err)
	}
	return q, nil
}

func (r ExchangeRate) inv() (ExchangeRate, error) {
	b, q, d, e := r.Base(), r.Quote(), r.Decimal(), decimal.One
	d, err := e.QuoExact(d, b.Scale())
	if err != nil {
		return ExchangeRate{}, err
	}
	return newExchRateSafe(q, b, d)
}

// SameCurr returns true if exchange rates are denominated in the same base
// and quote currencies.
// See also methods [ExchangeRate.Base] and [ExchangeRate.Quote].
func (r ExchangeRate) SameCurr(q ExchangeRate) bool {
	return r.Base() == q.Base() && r.Quote() == q.Quote()
}

// SameScale returns true if exchange rates have the same scale.
// See also method [ExchangeRate.Scale].
func (r ExchangeRate) SameScale(q ExchangeRate) bool {
	d, e := r.Decimal(), q.Decimal()
	return d.SameScale(e)
}

// Scale returns the number of digits after the decimal point.
// See also method [ExchangeRate.MinScale].
func (r ExchangeRate) Scale() int {
	return r.Decimal().Scale()
}

// MinScale returns the smallest scale that the rate can be rescaled to
// without rounding.
// See also method [ExchangeRate.Trim].
func (r ExchangeRate) MinScale() int {
	return r.Decimal().MinScale()
}

// IsZero returns:
//
//	true  if r == 0
//	false otherwise
func (r ExchangeRate) IsZero() bool {
	return r.Decimal().IsZero()
}

// IsPos returns:
//
//	true  if r > 0
//	false otherwise
func (r ExchangeRate) IsPos() bool {
	return r.Decimal().IsPos()
}

// Sign returns:
//
//	 0 if r = 0
//	+1 if r > 0
func (r ExchangeRate) Sign() int {
	return r.Decimal().Sign()
}

// IsOne returns:
//
//	true  if r == 1
//	false otherwise
func (r ExchangeRate) IsOne() bool {
	return r.Decimal().IsOne()
}

// WithinOne returns:
//
//	true  if 0 <= r < 1
//	false otherwise
func (r ExchangeRate) WithinOne() bool {
	return r.Decimal().WithinOne()
}

// Ceil returns a rate rounded up to the specified number of digits after
// the decimal point using [rounding toward positive infinity].
// See also method [ExchangeRate.Floor].
//
// [rounding toward positive infinity]: https://en.wikipedia.org/wiki/Rounding#Rounding_up
func (r ExchangeRate) Ceil(scale int) (ExchangeRate, error) {
	b, q, d := r.Base(), r.Quote(), r.Decimal()
	d = d.Ceil(scale).Pad(q.Scale())
	return newExchRateSafe(b, q, d)
}

// Floor returns a rate rounded down to the specified number of digits after
// the decimal point using [rounding toward negative infinity].
// See also method [ExchangeRate.Ceil].
//
// Floor returns an error if the result is 0.
//
// [rounding toward negative infinity]: https://en.wikipedia.org/wiki/Rounding#Rounding_down
func (r ExchangeRate) Floor(scale int) (ExchangeRate, error) {
	b, q, d := r.Base(), r.Quote(), r.Decimal()
	d = d.Floor(scale).Pad(q.Scale())
	p, err := newExchRateSafe(b, q, d)
	if err != nil {
		return ExchangeRate{}, fmt.Errorf("flooring %v: %w", r, err)
	}
	return p, nil
}

// Trunc returns a rate truncated to the specified number of digits after
// the decimal point using [rounding toward zero].
//
// Trunc returns an error if the result is 0.
//
// [rounding toward zero]: https://en.wikipedia.org/wiki/Rounding#Rounding_toward_zero
func (r ExchangeRate) Trunc(scale int) (ExchangeRate, error) {
	b, q, d := r.Base(), r.Quote(), r.Decimal()
	d = d.Trunc(scale).Pad(q.Scale())
	p, err := newExchRateSafe(b, q, d)
	if err != nil {
		return ExchangeRate{}, fmt.Errorf("truncating %v: %w", r, err)
	}
	return p, nil
}

// Trim returns a rate with trailing zeros removed up to the given scale.
// If the given scale is less than the scale of the quorte currency, the
// zeros will be removed up to the scale of the quote currency instead.
func (r ExchangeRate) Trim(scale int) ExchangeRate {
	b, q, d := r.Base(), r.Quote(), r.Decimal()
	scale = max(scale, q.Scale())
	d = d.Trim(scale)
	return newExchRateUnsafe(b, q, d)
}

// Round returns a rate rounded to the specified number of digits after
// the decimal point using [rounding half to even] (banker's rounding).
// See also method [ExchangeRate.Rescale].
//
// Round returns an error if the result is 0.
//
// [rounding half to even]: https://en.wikipedia.org/wiki/Rounding#Rounding_half_to_even
func (r ExchangeRate) Round(scale int) (ExchangeRate, error) {
	b, q, d := r.Base(), r.Quote(), r.Decimal()
	d = d.Round(scale).Pad(q.Scale())
	p, err := newExchRateSafe(b, q, d)
	if err != nil {
		return ExchangeRate{}, fmt.Errorf("rounding %v: %w", r, err)
	}
	return p, nil
}

// Quantize returns a rate rescaled to the same scale as rate q.
// The currency and the sign of rate q are ignored.
// See also methods [ExchangeRate.Scale], [ExchangeRate.SameScale], [ExchangeRate.Rescale].
//
// Quantize returns an error if the result is 0.
func (r ExchangeRate) Quantize(q ExchangeRate) (ExchangeRate, error) {
	p, err := r.rescale(q.Scale())
	if err != nil {
		return ExchangeRate{}, fmt.Errorf("rescaling %v to the scale of %v: %w", r, q, err)
	}
	return p, nil
}

// Rescale returns a rate rounded or zero-padded to the given number of digits
// after the decimal point.
// See also method [ExchangeRate.Round].
//
// Rescale returns an error if the result is 0.
func (r ExchangeRate) Rescale(scale int) (ExchangeRate, error) {
	q, err := r.rescale(scale)
	if err != nil {
		return ExchangeRate{}, fmt.Errorf("rescaling %v: %w", r, err)
	}
	return q, nil
}

func (r ExchangeRate) rescale(scale int) (ExchangeRate, error) {
	b, q, d := r.Base(), r.Quote(), r.Decimal()
	d = d.Rescale(scale).Pad(q.Scale())
	return newExchRateSafe(b, q, d)
}

// String method implements the [fmt.Stringer] interface and returns a string
// representation of the exchange rate.
// See also methods [Currency.String] and [Decimal.String].
//
// [fmt.Stringer]: https://pkg.go.dev/fmt#Stringer
// [Decimal.String]: https://pkg.go.dev/github.com/govalues/decimal#Decimal.String
func (r ExchangeRate) String() string {
	var buf [32]byte
	pos := len(buf) - 1
	coef := r.Decimal().Coef()
	scale := r.Decimal().Scale()

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

	// Delimiter
	buf[pos] = ' '
	pos--

	// Quote Currency
	curr := r.Quote().Code()
	for i := len(curr) - 1; i >= 0; i-- {
		buf[pos] = curr[i]
		pos--
	}

	// Deilimiter
	buf[pos] = '/'
	pos--

	// Base Currency
	curr = r.Base().Code()
	for i := len(curr) - 1; i >= 0; i-- {
		buf[pos] = curr[i]
		pos--
	}

	return string(buf[pos+1:])
}

// Format implements the [fmt.Formatter] interface.
// The following [format verbs] are available:
//
//	| Verb   | Example          | Description                   |
//	| ------ | ---------------- | ----------------------------- |
//	| %s, %v | EUR/USD 1.2500   | Currency pair and rate        |
//	| %q     | "EUR/USD 1.2500" | Quoted currency pair and rate |
//	| %f     | 1.2500           | Rate                          |
//	| %b     | EUR              | Base currency                 |
//	| %c     | USD              | Quote currency                |
//
// The '-' format flag can be used with all verbs.
// The '0' format flags can be used with all verbs except %c.
//
// Precision is only supported for the %f verb.
// The default precision is equal to the actual scale of the exchange rate.
//
// [format verbs]: https://pkg.go.dev/fmt#hdr-Printing
// [fmt.Formatter]: https://pkg.go.dev/fmt#Formatter
//
//gocyclo:ignore
func (r ExchangeRate) Format(state fmt.State, verb rune) {
	b, q, d := r.Base(), r.Quote(), r.Decimal()

	// Rescaling
	var tzeros int
	if verb == 'f' || verb == 'F' {
		var scale int
		switch p, ok := state.Precision(); {
		case ok:
			scale = p
		default:
			scale = d.Scale()
		}
		scale = max(scale, q.Scale())
		switch {
		case scale < d.Scale():
			d = d.Round(scale)
		case scale > d.Scale():
			tzeros = scale - d.Scale()
		}
	}

	// Integer and fractional digits
	var intdigs, fracdigs int
	switch rprec := d.Prec(); verb {
	case 'b', 'B', 'c', 'C':
		// skip
	default:
		fracdigs = d.Scale()
		if rprec > fracdigs {
			intdigs = rprec - fracdigs
		}
		if d.WithinOne() {
			intdigs++ // leading 0
		}
	}

	// Decimal point
	var dpoint int
	if fracdigs > 0 || tzeros > 0 {
		dpoint = 1
	}

	// Currency codes and delimiters
	basecode, quocode := "", ""
	var basesyms, quosyms, pairdel, currdel int
	switch verb {
	case 'f', 'F':
		// skip
	case 'b', 'B':
		basecode = b.Code()
		basesyms = len(basecode)
	case 'c', 'C':
		quocode = q.Code()
		quosyms = len(quocode)
	default:
		basecode = b.Code()
		quocode = q.Code()
		basesyms = len(basecode)
		quosyms = len(quocode)
		pairdel = 1
		currdel = 1
	}

	// Opening and closing quotes
	var lquote, tquote int
	if verb == 'q' || verb == 'Q' {
		lquote, tquote = 1, 1
	}

	// Calculating padding
	width := lquote + basesyms + pairdel + quosyms + currdel + intdigs + dpoint + fracdigs + tzeros + tquote
	var lspaces, lzeros, tspaces int
	if w, ok := state.Width(); ok && w > width {
		switch {
		case state.Flag('-'):
			tspaces = w - width
		case state.Flag('0') && verb != 'c' && verb != 'C' && verb != 'b' && verb != 'B':
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

	// Currency delimiter
	if currdel > 0 {
		buf[pos] = ' '
		pos--
	}

	// Quote currency
	for i := quosyms; i > 0; i-- {
		buf[pos] = quocode[i-1]
		pos--
	}

	// Pair delimiter
	if pairdel > 0 {
		buf[pos] = '/'
		pos--
	}

	// Base currency
	for i := basesyms; i > 0; i-- {
		buf[pos] = basecode[i-1]
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
	case 'q', 'Q', 's', 'S', 'v', 'V', 'f', 'F', 'b', 'B', 'c', 'C':
		state.Write(buf)
	default:
		state.Write([]byte("%!"))
		state.Write([]byte{byte(verb)})
		state.Write([]byte("(money.ExchangeRate="))
		state.Write(buf)
		state.Write([]byte(")"))
	}
}
