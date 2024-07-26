package money

import (
	"database/sql/driver"
	"errors"
	"fmt"
)

//go:generate go run scripts/currency/codegen.go

// Currency type represents a currency in the global financial system.
// The zero value is [XXX], which indicates an unknown currency.
//
// Currency is implemented as an integer index into an in-memory array that
// stores properties defined by [ISO 4217], such as code and scale.
// This design ensures safe concurrency for multiple goroutines accessing
// the same Currency value.
//
// When persisting a currency value, use the alphabetic code returned by
// the [Currency.Code] method, rather than the integer index, as mapping between
// index and a particular currency may change in future versions.
//
// [ISO 4217]: https://en.wikipedia.org/wiki/ISO_4217
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
		panic(fmt.Sprintf("ParseCurr(%q) failed: %v", curr, err))
	}
	return c
}

// Scale returns the number of digits after the decimal point required for
// representing the minor unit of a currency.
// The currently supported currencies use scales of 0, 2, or 3:
//   - A scale of 0 indicates currencies without minor units.
//     For example, the [Japanese Yen] does not have minor units.
//   - A scale of 2 indicates currencies that use 2 digits to represent their minor units.
//     For example, the [US Dollar] represents its minor unit, 1 cent, as 0.01 dollars.
//   - A scale of 3 indicates currencies with 3 digits in their minor units.
//     For instance, the minor unit of the [Omani Rial], 1 baisa, is represented as 0.001 rials.
//
// [Japanese Yen]: https://en.wikipedia.org/wiki/Japanese_yen
// [US Dollar]: https://en.wikipedia.org/wiki/United_States_dollar
// [Omani Rial]: https://en.wikipedia.org/wiki/Omani_rial
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
// See also method [Currency.Format].
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
func (c *Currency) Scan(value any) error {
	var err error
	switch value := value.(type) {
	case string:
		*c, err = ParseCurr(value)
	case []byte:
		*c, err = ParseCurr(string(value))
	case nil:
		err = fmt.Errorf("converting to %T: nil is not supported", c)
	default:
		err = fmt.Errorf("converting from %T to %T: type %T is not supported", value, c, value)
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
	//nolint:errcheck
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
