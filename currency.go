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

var errInvalidCurrency = errors.New("invalid currency")

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
		return XXX, errInvalidCurrency
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

// String method implements the [fmt.Stringer] interface and returns
// a string representation of the Currency value.
// See also method [Currency.Format].
//
// [fmt.Stringer]: https://pkg.go.dev/fmt#Stringer
func (c Currency) String() string {
	return c.Code()
}

// UnmarshalJSON implements the [json.Unmarshaler] interface.
// See also constructor [ParseCurr].
//
// [json.Unmarshaler]: https://pkg.go.dev/encoding/json#Unmarshaler
func (c *Currency) UnmarshalJSON(text []byte) error {
	if string(text) == "null" {
		return nil
	}
	if len(text) >= 2 && text[0] == '"' && text[len(text)-1] == '"' {
		text = text[1 : len(text)-1]
	}
	var err error
	*c, err = ParseCurr(string(text))
	if err != nil {
		return fmt.Errorf("unmarshaling %T: %w", XXX, err)
	}
	return nil
}

// MarshalJSON implements the [json.Marshaler] interface.
// MarshalJSON always returns a 3-letter code.
// See also method [Currency.Code].
//
// [json.Marshaler]: https://pkg.go.dev/encoding/json#Marshaler
func (c Currency) MarshalJSON() ([]byte, error) {
	text := make([]byte, 0, 5)
	text = append(text, '"')
	text = append(text, c.Code()...)
	text = append(text, '"')
	return text, nil
}

// UnmarshalText implements [encoding.TextUnmarshaler] interface.
// See also constructor [ParseCurr].
//
// [encoding.TextUnmarshaler]: https://pkg.go.dev/encoding#TextUnmarshaler
func (c *Currency) UnmarshalText(text []byte) error {
	var err error
	*c, err = ParseCurr(string(text))
	if err != nil {
		return fmt.Errorf("unmarshaling %T: %w", XXX, err)
	}
	return err
}

// AppendText implements the [encoding.TextAppender] interface.
// AppendText always appends a 3-letter code.
// See also method [Currency.Code].
//
// [encoding.TextAppender]: https://pkg.go.dev/encoding#TextAppender
func (c Currency) AppendText(text []byte) ([]byte, error) {
	return append(text, c.Code()...), nil
}

// MarshalText implements [encoding.TextMarshaler] interface.
// MarshalText always returns a 3-letter code.
// See also method [Currency.Code].
//
// [encoding.TextMarshaler]: https://pkg.go.dev/encoding#TextMarshaler
func (c Currency) MarshalText() ([]byte, error) {
	return []byte(c.Code()), nil
}

// UnmarshalBinary implements the [encoding.BinaryUnmarshaler] interface.
// See also constructor [ParseCurr].
//
// [encoding.BinaryUnmarshaler]: https://pkg.go.dev/encoding#BinaryUnmarshaler
func (c *Currency) UnmarshalBinary(text []byte) error {
	var err error
	*c, err = ParseCurr(string(text))
	if err != nil {
		return fmt.Errorf("unmarshaling %T: %w", XXX, err)
	}
	return err
}

// AppendBinary implements the [encoding.BinaryAppender] interface.
// AppendBinary always appends a 3-letter code.
// See also method [Currency.Code].
//
// [encoding.BinaryAppender]: https://pkg.go.dev/encoding#BinaryAppender
func (c Currency) AppendBinary(data []byte) ([]byte, error) {
	return append(data, c.Code()...), nil
}

// MarshalBinary implements the [encoding.BinaryMarshaler] interface.
// MarshalBinary always returns a 3-letter code.
// See also method [Currency.Code].
//
// [encoding.BinaryMarshaler]: https://pkg.go.dev/encoding#BinaryMarshaler
func (c Currency) MarshalBinary() ([]byte, error) {
	return []byte(c.Code()), nil
}

// UnmarshalBSONValue implements the [v2/bson.ValueUnmarshaler] interface.
// See also constructor [ParseCurr].
//
// [v2/bson.ValueUnmarshaler]: https://pkg.go.dev/go.mongodb.org/mongo-driver/v2/bson#ValueUnmarshaler
func (c *Currency) UnmarshalBSONValue(typ byte, data []byte) error {
	// constants are from https://bsonspec.org/spec.html
	var err error
	switch typ {
	case 2:
		*c, err = parseBSONString(data)
	case 10:
		// null, do nothing
	default:
		err = fmt.Errorf("BSON type %d is not supported", typ)
	}
	if err != nil {
		err = fmt.Errorf("converting from BSON type %d to %T: %w", typ, XXX, err)
	}
	return err
}

// MarshalBSONValue implements the [v2/bson.ValueMarshaler] interface.
// MarshalBSONValue always returns a 3-letter code.
// See also method [Currency.Code].
//
// [v2/bson.ValueMarshaler]: https://pkg.go.dev/go.mongodb.org/mongo-driver/v2/bson#ValueMarshaler
func (c Currency) MarshalBSONValue() (typ byte, data []byte, err error) {
	return 2, c.bsonString(), nil
}

// parseBSONString parses a BSON string to currency.
// The byte order of the input data must be little-endian.
func parseBSONString(data []byte) (Currency, error) {
	if len(data) < 4 {
		return XXX, fmt.Errorf("%w: invalid data length %v", errInvalidCurrency, len(data))
	}
	u := uint32(data[0])
	u |= uint32(data[1]) << 8
	u |= uint32(data[2]) << 16
	u |= uint32(data[3]) << 24
	l := int(int32(u)) //nolint:gosec
	if l < 1 || len(data) < l+4 {
		return XXX, fmt.Errorf("%w: invalid string length %v", errInvalidCurrency, l)
	}
	if data[l+4-1] != 0 {
		return XXX, fmt.Errorf("%w: invalid null terminator %v", errInvalidCurrency, data[l+4-1])
	}
	s := string(data[4 : l+4-1])
	return ParseCurr(s)
}

// bosnString returns the BSON string representation of the currency.
// The byte order of the result is little-endian.
func (c Currency) bsonString() []byte {
	s := c.String()
	l := len(s) + 1
	data := make([]byte, 4+l)
	data[0] = byte(l)
	data[1] = byte(l >> 8)
	data[2] = byte(l >> 16)
	data[3] = byte(l >> 24)
	copy(data[4:], s)
	data[4+l-1] = 0
	return data
}

// Scan implements the [sql.Scanner] interface.
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
		err = fmt.Errorf("%T does not support null values, use %T or *%T", XXX, NullCurrency{}, XXX)
	default:
		err = fmt.Errorf("type %T is not supported", value)
	}
	if err != nil {
		err = fmt.Errorf("converting from %T to %T: %w", value, XXX, err)
	}
	return err
}

// Value implements the [driver.Valuer] interface.
//
// [driver.Valuer]: https://pkg.go.dev/database/sql/driver#Valuer
func (c Currency) Value() (driver.Value, error) {
	return c.Code(), nil
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
	for range tspaces {
		buf[pos] = ' '
		pos--
	}

	// Closing quote
	for range tquote {
		buf[pos] = '"'
		pos--
	}

	// Currency symbols
	for i := range currlen {
		buf[pos] = curr[currlen-i-1]
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

// NullCurrency represents a currency that can be null.
// Its zero value is null.
// NullCurrency is not thread-safe.
type NullCurrency struct {
	Currency Currency
	Valid    bool
}

// Scan implements the [sql.Scanner] interface.
// See also method [Currency.Scan].
//
// [sql.Scanner]: https://pkg.go.dev/database/sql#Scanner
func (n *NullCurrency) Scan(value any) error {
	if value == nil {
		n.Currency = XXX
		n.Valid = false
		return nil
	}
	n.Valid = true
	return n.Currency.Scan(value)
}

// Value implements the [driver.Valuer] interface.
// See also method [Currency.Value].
//
// [driver.Valuer]: https://pkg.go.dev/database/sql/driver#Valuer
func (n NullCurrency) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.Currency.Value()
}

// UnmarshalJSON implements the [json.Unmarshaler] interface.
// See also method [Currency.UnmarshalJSON].
//
// [json.Unmarshaler]: https://pkg.go.dev/encoding/json#Unmarshaler
func (n *NullCurrency) UnmarshalJSON(text []byte) error {
	if string(text) == "null" {
		n.Currency = XXX
		n.Valid = false
		return nil
	}
	n.Valid = true
	return n.Currency.UnmarshalJSON(text)
}

// MarshalJSON implements the [json.Marshaler] interface.
// See also method [Currency.MarshalJSON].
//
// [json.Marshaler]: https://pkg.go.dev/encoding/json#Marshaler
func (n NullCurrency) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return n.Currency.MarshalJSON()
}

// UnmarshalBSONValue implements the [v2/bson.ValueUnmarshaler] interface.
// See also method [Currency.UnmarshalBSONValue].
//
// [v2/bson.ValueUnmarshaler]: https://pkg.go.dev/go.mongodb.org/mongo-driver/v2/bson#ValueUnmarshaler
func (n *NullCurrency) UnmarshalBSONValue(typ byte, data []byte) error {
	if typ == 10 {
		n.Currency = XXX
		n.Valid = false
		return nil
	}
	n.Valid = true
	return n.Currency.UnmarshalBSONValue(typ, data)
}

// MarshalBSONValue implements the [v2/bson.ValueMarshaler] interface.
// See also method [Currency.MarshalBSONValue].
//
// [v2/bson.ValueMarshaler]: https://pkg.go.dev/go.mongodb.org/mongo-driver/v2/bson#ValueMarshaler
func (n NullCurrency) MarshalBSONValue() (typ byte, data []byte, err error) {
	if !n.Valid {
		return 10, nil, nil
	}
	return n.Currency.MarshalBSONValue()
}
