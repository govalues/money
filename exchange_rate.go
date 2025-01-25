package money

import (
	"fmt"
	"math"
	"strconv"

	"github.com/govalues/decimal"
)

var errRateOverflow = fmt.Errorf("rate overflow")

// ExchangeRate represents an exchange rate between two currencies.
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
func newExchRateUnsafe(m, n Currency, d decimal.Decimal) ExchangeRate {
	return ExchangeRate{base: m, quote: n, value: d}
}

// newExchRateSafe creates a new rate and checks the sign and the scale.
func newExchRateSafe(m, n Currency, d decimal.Decimal) (ExchangeRate, error) {
	if d.IsZero() {
		return ExchangeRate{}, fmt.Errorf("exchange rate cannot be 0")
	}
	if d.IsNeg() {
		return ExchangeRate{}, fmt.Errorf("exchange rate must be positive")
	}
	if m == n && !d.IsOne() {
		return ExchangeRate{}, fmt.Errorf("exchange rate between identical currencies must be equal to 1")
	}
	if d.Scale() < n.Scale() {
		d = d.Pad(n.Scale())
		if d.Scale() < n.Scale() {
			return ExchangeRate{}, fmt.Errorf("padding exchange rate: %w", errRateOverflow)
		}
	}
	return newExchRateUnsafe(m, n, d), nil
}

// MustNewExchRate is like [NewExchRate] but panics if the rate cannot be constructed.
// It simplifies safe initialization of global variables holding rates.
func MustNewExchRate(base, quote string, value int64, scale int) ExchangeRate {
	r, err := NewExchRate(base, quote, value, scale)
	if err != nil {
		panic(fmt.Sprintf("NewExchRate(%q, %q, %v, %v) failed: %v", base, quote, value, scale, err))
	}
	return r
}

// NewExchRate returns a rate equal to value / 10^scale.
// If the scale of the rate is less than the scale of the quote currency, the
// result will be zero-padded to the right.
//
// NewExchRate returns an error if:
//   - the rate is 0 or negative
//   - the currency codes are not valid;
//   - the scale is negative or greater than [decimal.MaxScale];
//   - the integer part of the result has more than
//     ([decimal.MaxPrec] - [Currency.Scale]) digits.
//     For example, when the quote currency is US Dollars, NewAmount will return an error
//     if the integer part of the result has more than 17 digits (19 - 2 = 17).
func NewExchRate(base, quote string, value int64, scale int) (ExchangeRate, error) {
	// Currency
	m, err := ParseCurr(base)
	if err != nil {
		return ExchangeRate{}, fmt.Errorf("parsing currency: %w", err)
	}
	n, err := ParseCurr(quote)
	if err != nil {
		return ExchangeRate{}, fmt.Errorf("parsing currency: %w", err)
	}
	// Decimal
	d, err := decimal.New(value, scale)
	if err != nil {
		return ExchangeRate{}, fmt.Errorf("converting coefficient: %w", err)
	}
	// Rate
	r, err := newExchRateSafe(m, n, d)
	if err != nil {
		return ExchangeRate{}, fmt.Errorf("converting coefficient: %w", err)
	}
	return r, nil
}

// NewExchRateFromDecimal returns a rate with the specified currencies and value.
// If the scale of the rate is less than the scale of quote currency, the result
// will be zero-padded to the right. See also method [ExchangeRate.Decimal].
//
// NewExchRateFromDecimal returns an error if
//   - the rate is 0 or negative;
//   - the integer part of the result has more than
//     ([decimal.MaxPrec] - [Currency.Scale]) digits.
//     For example, when the quote currency is US Dollars, NewExchRateFromDecimal will return
//     an error if the integer part of the result has more than 17 digits (19 - 2 = 17).
func NewExchRateFromDecimal(base, quote Currency, rate decimal.Decimal) (ExchangeRate, error) {
	return newExchRateSafe(base, quote, rate)
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

// NewExchRateFromInt64 converts a pair of integers, representing the whole and
// fractional parts, to a (possibly rounded) rate equal to whole + frac / 10^scale.
// NewExchRateFromInt64 deletes trailing zeros up to the scale of the quote currency.
// This method is useful for converting rates from [protobuf] format.
// See also method [ExchangeRate.Int64].
//
// NewExchRateFromInt64 returns an error if:
//   - the rate is 0 or negative;
//   - the currency codes are not valid;
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
	m, err := ParseCurr(base)
	if err != nil {
		return ExchangeRate{}, fmt.Errorf("parsing currency: %w", err)
	}
	n, err := ParseCurr(quote)
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
		f = f.Trim(n.Scale())
		d, err = d.AddExact(f, n.Scale())
		if err != nil {
			return ExchangeRate{}, fmt.Errorf("converting integers: %w", err)
		}
	}
	// Amount
	return newExchRateSafe(m, n, d)
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

// NewExchRateFromFloat64 converts a float to a (possibly rounded) rate.
// See also method [ExchangeRate.Float64].
//
// NewExchRateFromFloat64 returns an error if:
//   - the rate is 0 or negative;
//   - the currency codes are not valid;
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

// Float64 returns the nearest binary floating-point number rounded
// using [rounding half to even] (banker's rounding).
// See also constructor [NewExchRateFromFloat64].
//
// This conversion may lose data, as float64 has a smaller precision
// than the underlying decimal type.
//
// [rounding half to even]: https://en.wikipedia.org/wiki/Rounding#Rounding_half_to_even
func (r ExchangeRate) Float64() (f float64, ok bool) {
	return r.Decimal().Float64()
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

// ParseExchRate converts currency and decimal strings to a (possibly rounded) rate.
// If the scale of the rate is less than the scale of the quote currency,
// the result will be zero-padded to the right.
// See also constructors [ParseCurr] and [decimal.Parse].
func ParseExchRate(base, quote, rate string) (ExchangeRate, error) {
	// Currency
	m, err := ParseCurr(base)
	if err != nil {
		return ExchangeRate{}, fmt.Errorf("parsing base currency: %w", err)
	}
	n, err := ParseCurr(quote)
	if err != nil {
		return ExchangeRate{}, fmt.Errorf("parsing quote currency: %w", err)
	}
	// Decimal
	d, err := decimal.ParseExact(rate, n.Scale())
	if err != nil {
		return ExchangeRate{}, fmt.Errorf("parsing exchange rate: %w", err)
	}
	// Rate
	r, err := newExchRateSafe(m, n, d)
	if err != nil {
		return ExchangeRate{}, fmt.Errorf("parsing exchange rate: %w", err)
	}
	return r, nil
}

// String method implements the [fmt.Stringer] interface and returns a string
// representation of the exchange rate.
// See also methods [Currency.String] and [Decimal.String].
//
// [fmt.Stringer]: https://pkg.go.dev/fmt#Stringer
// [Decimal.String]: https://pkg.go.dev/github.com/govalues/decimal#Decimal.String
func (r ExchangeRate) String() string {
	return string(r.bytes())
}

// bytes returns a string representation of the exchange rate as a byte slice.
func (r ExchangeRate) bytes() []byte {
	text := make([]byte, 0, 32)
	return r.append(text)
}

// append appends a string representation of the exchange rate to the given byte slice.
func (r ExchangeRate) append(text []byte) []byte {
	text, _ = r.Base().AppendText(text) // Currency.AppendText is always successful
	text = append(text, '/')
	text, _ = r.Quote().AppendText(text) // Currency.AppendText is always successful
	text = append(text, ' ')
	text, _ = r.Decimal().AppendText(text) // Decimal.AppendText is always successful
	return text
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
//nolint:gocyclo
func (r ExchangeRate) Format(state fmt.State, verb rune) {
	m, n, d := r.Base(), r.Quote(), r.Decimal()

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
		scale = max(scale, n.Scale())
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
	var basecode, quocode string
	var basesyms, quosyms, pairdel, currdel int
	switch verb {
	case 'f', 'F':
		// skip
	case 'b', 'B':
		basecode = m.Code()
		basesyms = len(basecode)
	case 'c', 'C':
		quocode = n.Code()
		quosyms = len(quocode)
	default:
		basecode = m.Code()
		quocode = n.Code()
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
	for range tspaces {
		buf[pos] = ' '
		pos--
	}

	// Closing quote
	for range tquote {
		buf[pos] = '"'
		pos--
	}

	// Trailing zeros
	for range tzeros {
		buf[pos] = '0'
		pos--
	}

	// Fractional digits
	coef := d.Coef()
	for range fracdigs {
		buf[pos] = byte(coef%10) + '0'
		pos--
		coef /= 10
	}

	// Decimal point
	for range dpoint {
		buf[pos] = '.'
		pos--
	}

	// Integer digits
	for range intdigs {
		buf[pos] = byte(coef%10) + '0'
		pos--
		coef /= 10
	}

	// Leading zeros
	for range lzeros {
		buf[pos] = '0'
		pos--
	}

	// Currency delimiter
	for range currdel {
		buf[pos] = ' '
		pos--
	}

	// Quote currency
	for i := range quosyms {
		buf[pos] = quocode[quosyms-i-1]
		pos--
	}

	// Pair delimiter
	for range pairdel {
		buf[pos] = '/'
		pos--
	}

	// Base currency
	for i := range basesyms {
		buf[pos] = basecode[basesyms-i-1]
		pos--
	}

	// Opening quote
	for range lquote {
		buf[pos] = '"'
		pos--
	}

	// Leading spaces
	for range lspaces {
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

// Sign returns:
//
//	 0 if r = 0
//	+1 if r > 0
func (r ExchangeRate) Sign() int {
	return r.Decimal().Sign()
}

// IsPos returns:
//
//	true  if r > 0
//	false otherwise
func (r ExchangeRate) IsPos() bool {
	return r.Decimal().IsPos()
}

// IsZero returns:
//
//	true  if r == 0
//	false otherwise
func (r ExchangeRate) IsZero() bool {
	return r.Decimal().IsZero()
}

// CanConv returns true if [ExchangeRate.Conv] can be used to convert the given amount.
func (r ExchangeRate) CanConv(b Amount) bool {
	return r.IsPos() &&
		!(r.Base() == XXX || r.Quote() == XXX) &&
		(r.Base() == b.Curr() || r.Quote() == b.Curr())
}

// Conv returns a (possibly rounded) amount converted between the base and quote
// currencies, automatically calculating the correct direction.
// See also method [ExchangeRate.CanConv].
//
// Conv returns an error if:
//   - the currency of the amount does not match either the base or
//     the quote currency of the exchange rate;
//   - the integer part of the result has more than
//     ([decimal.MaxPrec] - [Currency.Scale]) digits.
//     For example, when converting to US Dollars, Conv will return an error
//     if the integer part of the result has more than 17 digits (19 - 2 = 17).
func (r ExchangeRate) Conv(b Amount) (Amount, error) {
	q, err := r.conv(b)
	if err != nil {
		return Amount{}, fmt.Errorf("converting [%v]: %w", b, err)
	}
	return q, nil
}

func (r ExchangeRate) conv(b Amount) (Amount, error) {
	if !r.CanConv(b) {
		return Amount{}, errCurrencyMismatch
	}
	m, n, d, e := r.Base(), r.Quote(), r.Decimal(), b.Decimal()
	if m == b.Curr() {
		// Direct conversion
		e, err := e.MulExact(d, n.Scale())
		if err != nil {
			return Amount{}, fmt.Errorf("[%v -> %v]: %w", m, n, err)
		}
		return newAmountSafe(n, e)
	}
	// Reverse conversion
	e, err := e.QuoExact(d, m.Scale())
	if err != nil {
		return Amount{}, fmt.Errorf("[%v <- %v]: %w", m, n, err)
	}
	return newAmountSafe(m, e)
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
	m, n, d := r.Base(), r.Quote(), r.Decimal()
	d, err := d.MulExact(e, n.Scale())
	if err != nil {
		return ExchangeRate{}, err
	}
	return newExchRateSafe(m, n, d)
}

// Deprecated: since [ExchangeRate.Conv] is now bidirectional, this method is no longer needed.
func (r ExchangeRate) Inv() (ExchangeRate, error) {
	q, err := r.inv()
	if err != nil {
		return ExchangeRate{}, fmt.Errorf("inverting %v: %w", r, err)
	}
	return q, nil
}

func (r ExchangeRate) inv() (ExchangeRate, error) {
	m, n, d, e := r.Base(), r.Quote(), r.Decimal(), decimal.One
	d, err := e.QuoExact(d, m.Scale())
	if err != nil {
		return ExchangeRate{}, err
	}
	return newExchRateSafe(n, m, d)
}

// Ceil returns a rate rounded up to the specified number of digits after
// the decimal point using [rounding toward positive infinity].
// See also method [ExchangeRate.Floor].
// If the given scale is negative, it is redefined to zero.
//
// [rounding toward positive infinity]: https://en.wikipedia.org/wiki/Rounding#Rounding_up
func (r ExchangeRate) Ceil(scale int) (ExchangeRate, error) {
	m, n, d := r.Base(), r.Quote(), r.Decimal()
	d = d.Ceil(scale).Pad(n.Scale())
	return newExchRateSafe(m, n, d)
}

// Floor returns a rate rounded down to the specified number of digits after
// the decimal point using [rounding toward negative infinity].
// If the given scale is negative, it is redefined to zero.
// See also method [ExchangeRate.Ceil].
//
// Floor returns an error if the result is 0.
//
// [rounding toward negative infinity]: https://en.wikipedia.org/wiki/Rounding#Rounding_down
func (r ExchangeRate) Floor(scale int) (ExchangeRate, error) {
	m, n, d := r.Base(), r.Quote(), r.Decimal()
	d = d.Floor(scale).Pad(n.Scale())
	f, err := newExchRateSafe(m, n, d)
	if err != nil {
		return ExchangeRate{}, fmt.Errorf("flooring %v: %w", r, err)
	}
	return f, nil
}

// Trunc returns a rate truncated to the specified number of digits after
// the decimal point using [rounding toward zero].
// If the given scale is negative, it is redefined to zero.
//
// Trunc returns an error if the result is 0.
//
// [rounding toward zero]: https://en.wikipedia.org/wiki/Rounding#Rounding_toward_zero
func (r ExchangeRate) Trunc(scale int) (ExchangeRate, error) {
	m, n, d := r.Base(), r.Quote(), r.Decimal()
	d = d.Trunc(scale).Pad(n.Scale())
	f, err := newExchRateSafe(m, n, d)
	if err != nil {
		return ExchangeRate{}, fmt.Errorf("truncating %v: %w", r, err)
	}
	return f, nil
}

// Trim returns a rate with trailing zeros removed up to the given scale.
// If the given scale is less than the scale of the quorte currency, the
// zeros will be removed up to the scale of the quote currency instead.
func (r ExchangeRate) Trim(scale int) ExchangeRate {
	m, n, d := r.Base(), r.Quote(), r.Decimal()
	scale = max(scale, n.Scale())
	d = d.Trim(scale)
	return newExchRateUnsafe(m, n, d)
}

// Round returns a rate rounded to the specified number of digits after
// the decimal point using [rounding half to even] (banker's rounding).
// If the given scale is negative, it is redefined to zero.
// See also method [ExchangeRate.Rescale].
//
// Round returns an error if the result is 0.
//
// [rounding half to even]: https://en.wikipedia.org/wiki/Rounding#Rounding_half_to_even
func (r ExchangeRate) Round(scale int) (ExchangeRate, error) {
	m, n, d := r.Base(), r.Quote(), r.Decimal()
	d = d.Round(scale).Pad(n.Scale())
	f, err := newExchRateSafe(m, n, d)
	if err != nil {
		return ExchangeRate{}, fmt.Errorf("rounding %v: %w", r, err)
	}
	return f, nil
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
// If the given scale is negative, it is redefined to zero.
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
	m, n, d := r.Base(), r.Quote(), r.Decimal()
	d = d.Rescale(scale).Pad(n.Scale())
	return newExchRateSafe(m, n, d)
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
