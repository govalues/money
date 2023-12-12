package money

import (
	"database/sql/driver"
	"errors"
	"fmt"
)

//go:generate go run scripts/currency/codegen.go

// Currency type represents a currency in the global financial system.
// The zero value is "XXX", which indicates an unknown currency.
//
// Currency is implemented as an integer index into an in-memory array that
// stores information such as code and scale.
// This design ensures safe concurrency for multiple goroutines accessing
// the same Currency value.
//
// When persisting a currency value, use the alphabetic code returned by
// the [Currency.Code] method, rather than the integer index, as mapping between
// index and a particular currency may change in future versions.
type Currency uint8

var errUnknownCurrency = errors.New("unknown currency")

// ParseCurr converts a string to currency.
// The input string must be in one of the following formats:
//
//	USD
//	usd
//	840
//
// ParseCurr returns an error if the string does not represent a valid currency code.
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

// Scale returns the number of digits after the decimal point required for
// the minor unit of the currency.
// This represents the [ratio] of the minor unit to the major unit.
// A scale of 0 means that there is no minor unit for the currency, whereas
// scales of 1, 2, and 3 signify ratios of 10:1, 100:1, and 1000:1, respectively.
//
// [ratio]: https://en.wikipedia.org/wiki/ISO_4217#Minor_unit_fractions
func (c Currency) Scale() int {
	return int(scaleLookup[c])
}

// Num returns the [3-digit code] assigned to the currency by the ISO 4217 standard.
// If the currency does not have such a [code], the method will return an empty string.
//
// [3-digit code]: https://en.wikipedia.org/wiki/ISO_4217#Numeric_codes
// [code]: https://en.wikipedia.org/wiki/ISO_4217#X_currencies_(funds,_precious_metals,_supranationals,_other)
func (c Currency) Num() string {
	return numLookup[c]
}

// Code returns the [3-letter code] assigned to the currency by the ISO 4217 standard.
// This code is a unique identifier of the currency and is used in
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

// UnmarshalText implements [encoding.TextUnmarshaler] interface.
// Also see method [ParseCurr].
//
// [encoding.TextUnmarshaler]: https://pkg.go.dev/encoding#TextUnmarshaler
func (c *Currency) UnmarshalText(text []byte) error {
	var err error
	*c, err = ParseCurr(string(text))
	return err
}

// MarshalText implements [encoding.TextMarshaler] interface.
// Also see method [Currency.String].
//
// [encoding.TextMarshaler]: https://pkg.go.dev/encoding#TextMarshaler
func (c Currency) MarshalText() ([]byte, error) {
	return []byte(c.String()), nil
}

// Scan implements the [sql.Scanner] interface.
// See also method [ParseCurr].
//
// [sql.Scanner]: https://pkg.go.dev/database/sql#Scanner
func (c *Currency) Scan(v any) error {
	var err error
	switch v := v.(type) {
	case string:
		*c, err = ParseCurr(v)
	default:
		err = fmt.Errorf("failed to convert from %T to %T", v, XXX)
	}
	return err
}

// Value implements the [driver.Valuer] interface.
// See also method [Currency.String].
//
// [driver.Valuer]: https://pkg.go.dev/database/sql/driver#Valuer
func (c Currency) Value() (driver.Value, error) {
	return c.String(), nil
}

// Format implements the [fmt.Formatter] interface.
// The following [format verbs] are available:
//
//	| Verb       | Example | Description     |
//	| ---------- | ------- | --------------- |
//	| %c, %s, %v | USD     | Currency        |
//	| %q         | "USD"   | Quoted currency |
//
// The '-' format flag can be used with all verbs.
//
// [format verbs]: https://pkg.go.dev/fmt#hdr-Printing
// [fmt.Formatter]: https://pkg.go.dev/fmt#Formatter
func (c Currency) Format(state fmt.State, verb rune) {
	// Currency symbols
	curr := c.Code()
	currlen := len(curr)

	// Opening and closing quotes
	lquote, tquote := 0, 0
	if verb == 'q' || verb == 'Q' {
		lquote, tquote = 1, 1
	}

	// Calculating padding
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

	// Currency symbols
	for i := currlen; i > 0; i-- {
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

// NullCurrency represents a currency that can be null.
// Its zero value is null.
// NullCurrency is not thread-safe.
type NullCurrency struct {
	Currency Currency
	Valid    bool
}

// Scan implements the [sql.Scanner] interface.
// See also method [ParseCurr].
//
// [sql.Scanner]: https://pkg.go.dev/database/sql#Scanner
func (n *NullCurrency) Scan(value any) error {
	if value == nil {
		n.Currency = XXX
		n.Valid = false
		return nil
	}
	err := n.Currency.Scan(value)
	if err != nil {
		n.Currency = XXX
		n.Valid = false
		return err
	}
	n.Valid = true
	return nil
}

// Value implements the [driver.Valuer] interface.
// See also method [Currency.String].
//
// [driver.Valuer]: https://pkg.go.dev/database/sql/driver#Valuer
func (n NullCurrency) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.Currency.Value()
}
