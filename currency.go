package money

import (
	"errors"
	"fmt"
)

//go:generate go run scripts/currency/gen.go

// Currency type represents a currency in the global financial system.
// The zero value is "XXX", which indicates an unknown currency.
//
// Currency is implemented as an integer index into an in-memory array that
// stores information such as code, symbol, and scale.
// This design ensures safe concurrency for multiple goroutines accessing
// the same Currency value.
//
// When persisting a currency value, use the alphabetic code returned by
// the [Currency.Code] method, rather than the integer index, as mapping between
// index and a particular currency may change in future versions.
type Currency uint8

var (
	errUnknownCurrency = errors.New("unknown currency")
)

// ParseCurr converts a string to a currency.
// The input string must be in one of the following formats:
//
//	USD
//	usd
//	840
//
// ParseCurr returns error if string does not represent a valid currency code.
func ParseCurr(curr string) (Currency, error) {
	c, ok := currLookup[curr]
	if !ok {
		return XXX, errUnknownCurrency
	}
	return c, nil
}

// MustParseCurr is like [ParseCurr] but panics if the string cannot be parsed.
// It simplifies safe initialization of global variables holding currencies.
func MustParseCurr(curr string) Currency {
	c, err := ParseCurr(curr)
	if err != nil {
		panic(fmt.Sprintf("MustParseCurr(%q) failed: %v", curr, err))
	}
	return c
}

// UnmarshalText implements [encoding.TextUnmarshaler] interface.
// Also see method [ParseCurr].
//
// [encoding.TextUnmarshaler]: https://pkg.go.dev/encoding#TextUnmarshaler
func (c *Currency) UnmarshalText(text []byte) error {
	var err error
	*c, err = ParseCurr(string(text))
	return err
}

// Scale returns the number of digits after the decimal point required for
// the minor unit of the currency.
// This represents the [ratio] of the minor unit to the major unit.
// A scale of 0 means that there is no minor unit for the currency, whereas
// scales of 1, 2, and 3 signify ratios of 10:1, 100:1, and 1000:1, respectively.
//
// [ratio]: https://en.wikipedia.org/wiki/ISO_4217#Minor_unit_fractions
func (c Currency) Scale() int {
	return scaleLookup[c]
}

// Code returns the [3-digit code] assigned to the currency.
// If the currency does not have [such code], the method will return an empty string.
//
// [3-digit code]: https://en.wikipedia.org/wiki/ISO_4217#Numeric_codes
// [such code]: https://en.wikipedia.org/wiki/ISO_4217#X_currencies_(funds,_precious_metals,_supranationals,_other)
func (c Currency) Num() string {
	return numLookup[c]
}

// Code returns the [3-letter code] assigned to the currency.
// This code is a unique identifier of the currency, and is used in
// international finance and commerce.
// This method always returns a valid code.
//
// [3-letter code]: https://en.wikipedia.org/wiki/ISO_4217#National_currencies
func (c Currency) Code() string {
	return codeLookup[c]
}

// String method implements the [fmt.Stringer] interface and returns
// a string representation of the Currency value.
func (c Currency) String() string {
	return c.Code()
}

// MarshalText implements [encoding.TextMarshaler] interface.
// Also see method [Currency.String].
//
// [encoding.TextMarshaler]: https://pkg.go.dev/encoding#TextMarshaler
func (c Currency) MarshalText() ([]byte, error) {
	return []byte(c.String()), nil
}

// Format implements [fmt.Formatter] interface.
// The following [verbs] are available:
//
//	%s, %v: USD
//	%q:    "USD"
//	%c:     USD
//
// The '-' format flag can be used with all verbs.
//
// [verbs]: https://pkg.go.dev/fmt#hdr-Printing
// [fmt.Formatter]: https://pkg.go.dev/fmt#Formatter
func (c Currency) Format(state fmt.State, verb rune) {

	// Currency symbols
	curr := c.Code()
	currlen := len(curr)

	// Quotes
	lquote, tquote := 0, 0
	if verb == 'q' || verb == 'Q' {
		lquote, tquote = 1, 1
	}

	// Padding
	width := lquote + currlen + tquote
	lspaces, tspaces := 0, 0
	if w, ok := state.Width(); ok && w > width {
		switch {
		case state.Flag('-'):
			tspaces = w - width
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
	case 'q', 'Q', 's', 'S', 'v', 'V', 'c', 'C':
		state.Write(buf)
	default:
		state.Write([]byte("%!"))
		state.Write([]byte{byte(verb)})
		state.Write([]byte("(money.Currency="))
		state.Write(buf)
		state.Write([]byte(")"))
	}
}
