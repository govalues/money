package money

import (
	"fmt"
	"github.com/govalues/decimal"
	"testing"
	"unsafe"
)

func TestExchangeRate_ZeroValue(t *testing.T) {
	got := ExchangeRate{}
	// The zero value of an ExchangeRate cannot be created using NewExchRate or
	// ParseExchRate, so we check individual properties of the zero value instead.
	if got.Base() != XXX {
		t.Errorf("ExchangeRate{}.Base() = %v, want %v", got.Base(), XXX)
	}
	if got.Quote() != XXX {
		t.Errorf("ExchangeRate{}.Quote() = %v, want %v", got.Quote(), XXX)
	}
	if !got.IsZero() {
		t.Errorf("ExchangeRate{}.value.IsZero() = %v, want %v", got.IsZero(), true)
	}
}

func TestExchangeRate_Sizeof(t *testing.T) {
	r := ExchangeRate{}
	got := unsafe.Sizeof(r)
	want := uintptr(24)
	if got != want {
		t.Errorf("unsafe.Sizeof(%q) = %v, want %v", r, got, want)
	}
}

func TestNewExchRate(t *testing.T) {
	tests := []struct {
		base, quote Currency
		rate        string
		wantOk      bool
	}{
		{USD, EUR, "1.2000", true},
		{USD, EUR, "-1.2000", false},
		{USD, USD, "0.9999", false},
		{USD, USD, "1.0000", true},
		{USD, USD, "1.0001", false},
		{USD, JPY, "100000000000000000", false},
		{USD, EUR, "1000000000000000", false},
		{USD, OMR, "100000000000000", false},
	}
	for _, tt := range tests {
		rate := decimal.MustParse(tt.rate)
		_, err := NewExchRate(tt.base, tt.quote, rate)
		if !tt.wantOk && err == nil {
			t.Errorf("NewExchError(%v, %v, %v) did not fail", tt.base, tt.quote, rate)
		}
		if tt.wantOk && err != nil {
			t.Errorf("NewExchRate(%v, %v, %v) failed: %v", tt.base, tt.quote, rate, err)
		}
	}
}

func TestParseExchRate(t *testing.T) {

	t.Run("valid", func(t *testing.T) {
		tests := []struct {
			base, quote, rate   string
			wantBase, wantQuote Currency
			wantCoef            int64
			wantScale           int
		}{
			{"USD", "JPY", "132", USD, JPY, 132, 0},
			{"USD", "EUR", "1.2", USD, EUR, 12000, 4},
			{"USD", "OMR", "0.38", USD, OMR, 38000, 5},
		}
		for _, tt := range tests {
			rate := decimal.New(tt.wantCoef, tt.wantScale)
			want, err := NewExchRate(tt.wantBase, tt.wantQuote, rate)
			if err != nil {
				t.Errorf("NewExchRate(%v, %v, %v) failed: %v", tt.wantBase, tt.wantQuote, rate, err)
				continue
			}
			got, err := ParseExchRate(tt.base, tt.quote, tt.rate)
			if err != nil {
				t.Errorf("ParseExchRate(%q, %q, %q) failed: %v", tt.base, tt.quote, tt.rate, err)
				continue
			}
			if got != want {
				t.Errorf("ParseExchRate(%q, %q, %q) = %q, want %q", tt.base, tt.quote, tt.rate, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string]struct {
			base, quote, rate string
		}{
			"no data": {"", "", ""},
			"base 1":  {"AAA", "USD", "30000"},
			"quote 1": {"USD", "AAA", "0.00003"},
			"rate 1":  {"USD", "EUR", "x.0000"},
			"rate 2":  {"USD", "USD", "0.9999"},
			"rate 3":  {"USD", "EUR", "0.0"},
			"rate 4":  {"USD", "EUR", "-0.9999"},
		}
		for _, tt := range tests {
			_, err := ParseExchRate(tt.base, tt.quote, tt.rate)
			if err == nil {
				t.Errorf("ParseExchRate(%q, %q, %q) did not fail", tt.base, tt.quote, tt.rate)
			}
		}
	})
}

func TestExchangeRate_Mul(t *testing.T) {

	t.Run("valid", func(t *testing.T) {
		tests := []struct {
			base, quote, rate, factor, want string
		}{
			// There are few test cases here since Decimal.Mul
			// is already tested by the cases in the decimal package.
			{"USD", "EUR", "2", "2", "4"},
			{"USD", "EUR", "2", "3", "6"},
			{"USD", "EUR", "5", "1", "5"},
			{"USD", "EUR", "5", "2", "10"},
			{"USD", "EUR", "1.20", "2", "2.40"},
			{"USD", "EUR", "5.09", "7.1", "36.13900"},
			{"USD", "EUR", "2.5", "4", "10.0"},
			{"USD", "EUR", "2.50", "4", "10.00"},
			{"USD", "EUR", "0.70", "1.05", "0.735000"},
		}
		for _, tt := range tests {
			r := MustParseExchRate(tt.base, tt.quote, tt.rate)
			e := decimal.MustParse(tt.factor)
			want := MustParseExchRate(tt.base, tt.quote, tt.want)
			got := r.Mul(e)
			if got != want {
				t.Errorf("%q.Mul(%q) = %q, want %q", r, e, got, want)
			}
		}
	})

	t.Run("panic", func(t *testing.T) {
		tests := map[string]struct {
			base, quote, rate, factor string
		}{
			"overflow 1": {"USD", "EUR", "0.9", "10000000000000000"},
			"overflow 2": {"USD", "EUR", "100000000000000", "10"},
			"factor 1":   {"USD", "EUR", "0.9", "0.0"},
			"factor 2":   {"USD", "EUR", "0.9", "-0.1"},
		}
		for name, tt := range tests {
			a := MustParseExchRate(tt.base, tt.quote, tt.rate)
			e := decimal.MustParse(tt.factor)
			t.Run(name, func(t *testing.T) {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("%q.Mul(%q) did not panic", a, e)
					}
				}()
				a.Mul(e)
			})
		}
	})
}

func TestExchangeRate_Inv(t *testing.T) {

	t.Run("valid", func(t *testing.T) {
		tests := []struct {
			base, quote, rate, want string
		}{
			{"USD", "EUR", "0.5", "2"},
		}
		for _, tt := range tests {
			r := MustParseExchRate(tt.base, tt.quote, tt.rate)
			want := MustParseExchRate(tt.quote, tt.base, tt.want)
			got := r.Inv()
			if got != want {
				t.Errorf("%q.Inv() = %q, want %q", r, got, want)
			}
		}
	})

	t.Run("panic", func(t *testing.T) {
		q := ExchangeRate{}
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("%q.Inv() did not panic", q)
			}
		}()
		q.Inv()
	})
}

func TestExchangeRate_SameScaleAsCurr(t *testing.T) {
	tests := []struct {
		base, quote, rate string
		want              bool
	}{
		{"USD", "EUR", "1", true},
		{"USD", "EUR", "1.0", true},
		{"USD", "EUR", "1.00", true},
		{"USD", "EUR", "1.000", true},
		{"USD", "EUR", "1.0000", true},
		{"USD", "EUR", "1.00000", false},
		{"USD", "EUR", "1.000000", false},
	}
	for _, tt := range tests {
		r := MustParseExchRate(tt.base, tt.quote, tt.rate)
		got := r.SameScaleAsCurr()
		if got != tt.want {
			t.Errorf("%q.SameScaleAsCurr() = %v, want %v", r, got, tt.want)
		}
	}
}

func TestExchangeRate_Conv(t *testing.T) {

	t.Run("valid", func(t *testing.T) {
		tests := []struct {
			base, quote, rate, amount, want string
		}{
			{"JPY", "USD", "0.0075", "100", "0.7500"},
			{"EUR", "USD", "1.0995", "100.00", "109.950000"},
			{"OMR", "USD", "2.59765", "100.000", "259.76500000"},
		}
		for _, tt := range tests {
			r := MustParseExchRate(tt.base, tt.quote, tt.rate)
			a := MustParseAmount(tt.base, tt.amount)
			want := MustParseAmount(tt.quote, tt.want)
			got := r.Conv(a)
			if got != want {
				t.Errorf("%q.Conv(%q) = %q, want %q", r, a, got, want)
			}
		}
	})

	t.Run("panic", func(t *testing.T) {
		tests := map[string]struct {
			base, quote, rate, curr, amount string
		}{
			"currency 1": {"USD", "EUR", "1.2000", "JPY", "100"},
			"currency 2": {"XXX", "EUR", "1.2000", "XXX", "100"},
			"overflow 1": {"USD", "JPY", "1000.00", "USD", "10000000000000000.00"},
			"overflow 2": {"USD", "EUR", "10.0000", "USD", "10000000000000000.00"},
			"overflow 3": {"USD", "OMR", "10.00000", "USD", "1000000000000000.00"},
		}
		for name, tt := range tests {
			r := MustParseExchRate(tt.base, tt.quote, tt.rate)
			a := MustParseAmount(tt.curr, tt.amount)
			t.Run(name, func(t *testing.T) {
				defer func() {
					if e := recover(); e == nil {
						t.Errorf("%q.Conv(%q) did not panic", r, a)
					}
				}()
				r.Conv(a)
			})
		}
	})
}

func TestExchangeRate_Format(t *testing.T) {
	tests := []struct {
		base, quote, rate, format, want string
	}{
		// %T verb
		{"USD", "EUR", "100.00", "%T", "money.ExchangeRate"},
		// %q verb
		{"USD", "EUR", "100.00", "%q", "\"USD/EUR 100.0000\""},
		{"USD", "EUR", "100.00", "%+q", "\"USD/EUR 100.0000\""},  // '+' is ignored
		{"USD", "EUR", "100.00", "% q", "\"USD/EUR 100.0000\""},  // ' ' is ignored
		{"USD", "EUR", "100.00", "%.6q", "\"USD/EUR 100.0000\""}, // precision is ignored
		{"USD", "EUR", "100.00", "%16q", "\"USD/EUR 100.0000\""},
		{"USD", "EUR", "100.00", "%17q", "\"USD/EUR 100.0000\""},
		{"USD", "EUR", "100.00", "%18q", "\"USD/EUR 100.0000\""},
		{"USD", "EUR", "100.00", "%19q", " \"USD/EUR 100.0000\""},
		{"USD", "EUR", "100.00", "%019q", "\"USD/EUR 0100.0000\""},
		{"USD", "EUR", "100.00", "%+19q", " \"USD/EUR 100.0000\""}, // '+' is ignored
		{"USD", "EUR", "100.00", "%-19q", "\"USD/EUR 100.0000\" "},
		{"USD", "EUR", "100.00", "%+-021q", "\"USD/EUR 100.0000\"   "}, // '+' and '0' are ignored
		// %s verb
		{"USD", "EUR", "100.00", "%s", "USD/EUR 100.0000"},
		{"USD", "EUR", "100.00", "%+s", "USD/EUR 100.0000"},  // '+' is ignored
		{"USD", "EUR", "100.00", "% s", "USD/EUR 100.0000"},  // ' ' is ignored
		{"USD", "EUR", "100.00", "%.6s", "USD/EUR 100.0000"}, // precision is ignored
		{"USD", "EUR", "100.00", "%16s", "USD/EUR 100.0000"},
		{"USD", "EUR", "100.00", "%17s", " USD/EUR 100.0000"},
		{"USD", "EUR", "100.00", "%18s", "  USD/EUR 100.0000"},
		{"USD", "EUR", "100.00", "%19s", "   USD/EUR 100.0000"},
		{"USD", "EUR", "100.00", "%019s", "USD/EUR 000100.0000"},
		{"USD", "EUR", "100.00", "%+19s", "   USD/EUR 100.0000"}, // '+' is ignored
		{"USD", "EUR", "100.00", "%-19s", "USD/EUR 100.0000   "},
		{"USD", "EUR", "100.00", "%+-021s", "USD/EUR 100.0000     "}, // '+' and '0' are ignored
		// %v verb
		{"USD", "EUR", "100.00", "%v", "USD/EUR 100.0000"},
		{"USD", "EUR", "100.00", "%+v", "USD/EUR 100.0000"},  // '+' is ignored
		{"USD", "EUR", "100.00", "% v", "USD/EUR 100.0000"},  // ' ' is ignored
		{"USD", "EUR", "100.00", "%.6v", "USD/EUR 100.0000"}, // precision is ignored
		{"USD", "EUR", "100.00", "%16v", "USD/EUR 100.0000"},
		{"USD", "EUR", "100.00", "%17v", " USD/EUR 100.0000"},
		{"USD", "EUR", "100.00", "%18v", "  USD/EUR 100.0000"},
		{"USD", "EUR", "100.00", "%19v", "   USD/EUR 100.0000"},
		{"USD", "EUR", "100.00", "%019v", "USD/EUR 000100.0000"},
		{"USD", "EUR", "100.00", "%+19v", "   USD/EUR 100.0000"}, // '+' is ignored
		{"USD", "EUR", "100.00", "%-19v", "USD/EUR 100.0000   "},
		{"USD", "EUR", "100.00", "%+-021v", "USD/EUR 100.0000     "}, // '+' and '0' are ignored
		// %f verb
		{"JPY", "EUR", "0.01", "%f", "0.01"},
		{"JPY", "EUR", "100.00", "%f", "100.00"},
		{"OMR", "EUR", "0.01", "%f", "0.01000"},
		{"OMR", "EUR", "100.00", "%f", "100.00000"},
		{"USD", "EUR", "0.01", "%f", "0.0100"},
		{"USD", "EUR", "100.00", "%f", "100.0000"},
		{"USD", "EUR", "9.996208266660", "%f", "9.9962"},
		{"USD", "EUR", "0.9996208266660", "%f", "0.9996"},
		{"USD", "EUR", "0.09996208266660", "%f", "0.1000"},
		{"USD", "EUR", "0.009996208266660", "%f", "0.0100"},
		{"USD", "EUR", "100.00", "%+f", "100.0000"},  // '+' is ignored
		{"USD", "EUR", "100.00", "% f", "100.0000"},  // ' ' is ignored
		{"USD", "EUR", "100.00", "%.3f", "100.0000"}, // precision cannot be smaller than curr scale
		{"USD", "EUR", "100.00", "%.4f", "100.0000"},
		{"USD", "EUR", "100.00", "%.5f", "100.00000"},
		{"USD", "EUR", "100.00", "%.6f", "100.000000"},
		{"USD", "EUR", "100.00", "%.7f", "100.0000000"},
		{"USD", "EUR", "100.00", "%.8f", "100.00000000"},
		{"USD", "EUR", "100.00", "%9f", " 100.0000"},
		{"USD", "EUR", "100.00", "%10f", "  100.0000"},
		{"USD", "EUR", "100.00", "%11f", "   100.0000"},
		{"USD", "EUR", "100.00", "%12f", "    100.0000"},
		{"USD", "EUR", "100.00", "%013f", "00000100.0000"},
		{"USD", "EUR", "100.00", "%+13f", "     100.0000"}, // '+' is ignored
		{"USD", "EUR", "100.00", "%-13f", "100.0000     "},
		// %c verb
		{"USD", "EUR", "100.00", "%c", "USD/EUR"},
		{"USD", "EUR", "100.00", "%+c", "USD/EUR"}, // '+' is ignored
		{"USD", "EUR", "100.00", "% c", "USD/EUR"}, // ' ' is ignored
		{"USD", "EUR", "100.00", "%#c", "USD/EUR"}, // '#' is ignored
		{"USD", "EUR", "100.00", "%9c", "  USD/EUR"},
		{"USD", "EUR", "100.00", "%09c", "  USD/EUR"}, // '0' is ignored
		{"USD", "EUR", "100.00", "%#9c", "  USD/EUR"}, // '#' is ignored
		{"USD", "EUR", "100.00", "%-9c", "USD/EUR  "},
		{"USD", "EUR", "100.00", "%-#9c", "USD/EUR  "}, // '#' is ignored
		// wrong verbs
		{"USD", "EUR", "12.34", "%b", "%!b(money.ExchangeRate=USD/EUR 12.3400)"},
		{"USD", "EUR", "12.34", "%d", "%!d(money.ExchangeRate=USD/EUR 12.3400)"},
		{"USD", "EUR", "12.34", "%e", "%!e(money.ExchangeRate=USD/EUR 12.3400)"},
		{"USD", "EUR", "12.34", "%E", "%!E(money.ExchangeRate=USD/EUR 12.3400)"},
		{"USD", "EUR", "12.34", "%g", "%!g(money.ExchangeRate=USD/EUR 12.3400)"},
		{"USD", "EUR", "12.34", "%G", "%!G(money.ExchangeRate=USD/EUR 12.3400)"},
		{"USD", "EUR", "12.34", "%x", "%!x(money.ExchangeRate=USD/EUR 12.3400)"},
		{"USD", "EUR", "12.34", "%X", "%!X(money.ExchangeRate=USD/EUR 12.3400)"},
	}
	for _, tt := range tests {
		r := MustParseExchRate(tt.base, tt.quote, tt.rate)
		got := fmt.Sprintf(tt.format, r)
		if got != tt.want {
			t.Errorf("fmt.Sprintf(%q, %q) = %q, want %q", tt.format, r, got, tt.want)
		}
	}
}
