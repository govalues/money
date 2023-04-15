package money

import (
	"fmt"
	"github.com/govalues/decimal"
)

// ExchangeRate represents a unidirectional exchange rate between two currencies.
// The zero value corresponds to an exchange rate of "XXX/XXX 0", where XXX indicates
// an unknown currency.
// This type is designed to be safe for concurrent use by multiple goroutines.
type ExchangeRate struct {
	base  Currency        // currency being exchanged
	quote Currency        // currency being obtained in exchange for the base currency
	value decimal.Decimal // how many units of quote currency are needed to exchange for 1 unit of the base currency
}

// NewExchRate returns a new exchange rate between the base and quote currencies.
func NewExchRate(base, quote Currency, rate decimal.Decimal) (ExchangeRate, error) {
	if !rate.IsPos() {
		return ExchangeRate{}, fmt.Errorf("exchange rate must be positive")
	}
	if base == quote && !rate.IsOne() {
		return ExchangeRate{}, fmt.Errorf("exchange rate must be equal to 1")
	}
	if rate.Scale() < base.Scale()+quote.Scale() {
		if rate.Prec()-rate.Scale() > decimal.MaxPrec-base.Scale()-quote.Scale() {
			return ExchangeRate{}, fmt.Errorf("with a pair of currencies %v/%v, the integer part of a %T can have at most %v digit(s), but it has %v digit(s)", base, quote, ExchangeRate{}, decimal.MaxPrec-base.Scale()-quote.Scale(), rate.Prec()-rate.Scale())
		}
		rate = rate.Round(base.Scale() + quote.Scale())
	}
	return ExchangeRate{base: base, quote: quote, value: rate}, nil
}

func mustNewExchRate(base, quote Currency, rate decimal.Decimal) ExchangeRate {
	r, err := NewExchRate(base, quote, rate)
	if err != nil {
		panic(fmt.Sprintf("NewExchRate(%v, %v, %v) failed: %v", base, quote, rate, err))
	}
	return r
}

// ParseExchRate converts currency and decimal strings to a (possibly rounded) exchange rate.
// See also methods [ParseCurr] and [decimal.Parse].
func ParseExchRate(base, quote, rate string) (ExchangeRate, error) {
	b, err := ParseCurr(base)
	if err != nil {
		return ExchangeRate{}, fmt.Errorf("base currency parsing: %w", err)
	}
	q, err := ParseCurr(quote)
	if err != nil {
		return ExchangeRate{}, fmt.Errorf("quote currency parsing: %w", err)
	}
	d, err := decimal.ParseExact(rate, b.Scale()+q.Scale())
	if err != nil {
		return ExchangeRate{}, fmt.Errorf("rate parsing: %w", err)
	}
	r, err := NewExchRate(b, q, d)
	if err != nil {
		return ExchangeRate{}, fmt.Errorf("rate construction: %w", err)
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

// Mul returns an exchange rate with the same base and quote currencies,
// but with the rate multiplied by a positive factor e.
//
// Mul panics if factor e is not positive.
// To avoid this panic, use the [decimal.Decimal.IsPos] to verify that decimal
// is positive before calling Mul.
func (r ExchangeRate) Mul(e decimal.Decimal) ExchangeRate {
	if !e.IsPos() {
		panic(fmt.Sprintf("%q.Mul(%q) failed: factor must be postitve", r, e))
	}
	d := r.value
	f := d.MulExact(e, r.Base().Scale()+r.Quote().Scale())
	return mustNewExchRate(r.Base(), r.Quote(), f)
}

// CanConv returns true if [ExchangeRate.Conv] can be used to convert the given amount.
func (r ExchangeRate) CanConv(b Amount) bool {
	return b.Curr() == r.Base() &&
		r.Base() != XXX &&
		r.Quote() != XXX &&
		r.value.IsPos()
}

// Conv returns the amount converted from the base currency to the quote currency.
//
// Conv panics if the base currency of the exchange rate does not match
// the currency of the given amount.
// To avoid this panic, use the [ExchangeRate.CanConv] method to ensure
// the currencies are compatible before calling Conv.
func (r ExchangeRate) Conv(b Amount) Amount {
	if !r.CanConv(b) {
		panic(fmt.Sprintf("%q.Conv(%q) failed: %v", r, b, errCurrencyMismatch))
	}
	d := r.value
	e := b.value
	f := d.MulExact(e, r.Quote().Scale())
	return mustNewAmount(r.Quote(), f)
}

// Inv returns the inverse of the exchange rate.
func (r ExchangeRate) Inv() ExchangeRate {
	d := r.value
	if d.IsZero() {
		panic(fmt.Sprintf("%q.Inv() failed: zero rate does not have an inverse: %v", r, errDivisionByZero))
	}
	one := d.One()
	return mustNewExchRate(r.Quote(), r.Base(), one.Quo(d))
}

// SameCurr returns true if exchange rates are denominated in the same base
// and quote currencies.
// See also methods [ExchangeRate.Base] and [ExchangeRate.Quote].
func (r ExchangeRate) SameCurr(q ExchangeRate) bool {
	return q.Base() == r.Base() && q.Quote() == r.Quote()
}

// SameScale returns true if exchange rates have the same scale.
// See also method [ExchangeRate.Scale].
func (r ExchangeRate) SameScale(q ExchangeRate) bool {
	return q.Scale() == r.Scale()
}

// SameScaleAsCurr returns true if the scale of the exchange rate is equal to
// the sum of the scales of its base and quote currencies.
// See also method [ExchangeRate.RoundToCurr].
func (r ExchangeRate) SameScaleAsCurr() bool {
	return r.Scale() == r.Base().Scale()+r.Quote().Scale()
}

// Prec returns the number of digits in the coefficient.
func (r ExchangeRate) Prec() int {
	return r.value.Prec()
}

// Scale returns the number of digits after the decimal point.
func (r ExchangeRate) Scale() int {
	return r.value.Scale()
}

// IsZero returns:
//
//	true  if r == 0
//	false otherwise
func (r ExchangeRate) IsZero() bool {
	return r.value.IsZero()
}

// IsOne returns:
//
//	true  if r == 1
//	false otherwise
func (r ExchangeRate) IsOne() bool {
	return r.value.IsOne()
}

// WithinOne returns:
//
//	true  if 0 <= r < 1
//	false otherwise
func (r ExchangeRate) WithinOne() bool {
	return r.value.WithinOne()
}

// Round returns an exchange rate that is rounded to the specified number of digits after
// the decimal point.
// If the scale of the exchange rate is less than the specified scale, the result will be
// zero-padded to the right.
// If the specified scale is less than the sum of the scales of the base and quote
// currency then the exchange rate will be rounded to the sum of scales instead.
// See also method [ExchangeRate.RoundToCurr].
//
// Round panics if the integer part of the result exceeds the maximum precision.
// This limit is calculated as ([decimal.MaxPrec] - scale).
func (r ExchangeRate) Round(scale int) ExchangeRate {
	if scale < r.Base().Scale()+r.Quote().Scale() {
		scale = r.Base().Scale() + r.Quote().Scale()
	}
	d := r.value
	return mustNewExchRate(r.Base(), r.Quote(), d.Round(scale))
}

// RoundToCurr returns an exchange rate that is rounded to the sum of the scales of its base
// and quote currencies.
// See also method [ExchangeRate.SameScaleAsCurr].
func (r ExchangeRate) RoundToCurr() ExchangeRate {
	return r.Round(r.Base().Scale() + r.Quote().Scale())
}

// String method implements the [fmt.Stringer] interface and returns a string
// representation of the exchange rate.
// See also methods [Currency.String] and [Decimal.String].
//
// [fmt.Stringer]: https://pkg.go.dev/fmt#Stringer
// [Decimal.String]: https://pkg.go.dev/github.com/govalues/decimal#Decimal.String
func (r ExchangeRate) String() string {
	return r.Base().String() + "/" + r.Quote().String() + " " + r.value.String()
}

// Format implements [fmt.Formatter] interface.
// The following [format verbs] are available:
//
//	%s, %v: USD/EUR 1.2345
//	%q:    "USD/EUR 1.2345"
//	%f:     1.2345
//	%c:     USD/EUR
//
// The '-' format flag can be used with all verbs.
// The '0' format flags can be used with all verbs except %c.
//
// Precision is only supported for the %f verb.
// The default precision is equal to the sum of the scales of its base and quote currencies.
//
// [format verbs]: https://pkg.go.dev/fmt#hdr-Printing
// [fmt.Formatter]: https://pkg.go.dev/fmt#Formatter
func (r ExchangeRate) Format(state fmt.State, verb rune) {
	// Rescaling
	tzeroes := 0
	if verb == 'f' || verb == 'F' {
		scale := 0
		switch p, ok := state.Precision(); {
		case ok:
			scale = p
		case verb == 'f' || verb == 'F':
			scale = r.Base().Scale() + r.Quote().Scale()
		}
		switch {
		case scale < 0:
			scale = 0
		case scale > r.Scale():
			tzeroes = scale - r.Scale()
			scale = r.Scale()
		}
		r = r.Round(scale)
	}

	// Integer and fractional digits
	intdigs, fracdigs := 0, 0

	switch rprec := r.Prec(); verb {
	case 'c', 'C':
		// skip
	default:
		fracdigs = r.Scale()
		if rprec > fracdigs {
			intdigs = rprec - fracdigs
		}
		if r.WithinOne() {
			intdigs++ // leading 0
		}
	}

	// Decimal point
	dpoint := 0
	if fracdigs > 0 || tzeroes > 0 {
		dpoint = 1
	}

	// Currency symbols
	curr := ""
	switch verb {
	case 'f', 'F':
		// skip
	case 'c', 'C':
		curr = r.Base().String() + "/" + r.Quote().String()
	default:
		curr = r.Base().String() + "/" + r.Quote().String() + " "
	}
	currlen := len(curr)

	// Quotes
	lquote, tquote := 0, 0
	if verb == 'q' || verb == 'Q' {
		lquote, tquote = 1, 1
	}

	// Padding
	width := lquote + intdigs + dpoint + fracdigs + tzeroes + currlen + tquote
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
	coef := r.value.Coef()
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
	case 'q', 'Q', 's', 'S', 'v', 'V', 'f', 'F', 'c', 'C':
		state.Write(buf)
	default:
		state.Write([]byte("%!"))
		state.Write([]byte{byte(verb)})
		state.Write([]byte("(money.ExchangeRate="))
		state.Write(buf)
		state.Write([]byte(")"))
	}
}
