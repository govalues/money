package money

import (
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/govalues/decimal"
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
		t.Errorf("ExchangeRate{}.IsZero() = %v, want %v", got.IsZero(), true)
	}
}

func TestExchangeRate_Size(t *testing.T) {
	r := ExchangeRate{}
	got := unsafe.Sizeof(r)
	want := uintptr(24)
	if got != want {
		t.Errorf("unsafe.Sizeof(%q) = %v, want %v", r, got, want)
	}
}

func TestExchangeRate_Interfaces(t *testing.T) {
	var i any = ExchangeRate{}
	_, ok := i.(fmt.Stringer)
	if !ok {
		t.Errorf("%T does not implement fmt.Stringer", i)
	}
	_, ok = i.(fmt.Formatter)
	if !ok {
		t.Errorf("%T does not implement fmt.Formatter", i)
	}
}

func TestMustNewExchRate(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("MustNewExchRate(\"EUR\", \"USD\", 0, -1) did not panic")
			}
		}()
		MustNewExchRate("EUR", "USD", 0, -1)
	})
}

func TestMustParseExchRate(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("MustParseExchRate(\"EUR\", \"USD\", \".\") did not panic")
			}
		}()
		MustParseExchRate("EUR", "USD", ".")
	})
}

func TestNewExchRate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			base, quote string
			coef        int64
			scale       int
			want        string
		}{
			{"EUR", "JPY", math.MaxInt64, 0, "9223372036854775807"},
			{"EUR", "JPY", math.MaxInt64, 19, "0.9223372036854775807"},
			{"EUR", "USD", math.MaxInt64, 2, "92233720368547758.07"},
			{"EUR", "USD", math.MaxInt64, 19, "0.9223372036854775807"},
			{"EUR", "OMR", math.MaxInt64, 3, "9223372036854775.807"},
			{"EUR", "OMR", math.MaxInt64, 19, "0.9223372036854775807"},
		}
		for _, tt := range tests {
			got, err := NewExchRate(tt.base, tt.quote, tt.coef, tt.scale)
			if err != nil {
				t.Errorf("NewExchRate(%q, %q, %v, %v) failed: %v", tt.base, tt.quote, tt.coef, tt.scale, err)
				continue
			}
			want := MustParseExchRate(tt.base, tt.quote, tt.want)
			if got != want {
				t.Errorf("NewExchRate(%q, %q, %v, %v) = %q, want %q", tt.base, tt.quote, tt.coef, tt.scale, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string]struct {
			base, quote string
			coef        int64
			scale       int
		}{
			"base currency 1":  {"EEE", "USD", 1, 0},
			"quote currency 1": {"EUR", "UUU", 1, 0},
			"scale range 1":    {"EUR", "USD", 1, -1},
			"scale range 2":    {"EUR", "USD", 1, 20},
			"coefficient 1":    {"EUR", "USD", 0, 0},
			"coefficient 2":    {"EUR", "USD", -1, 0},
			"coefficient 3":    {"EUR", "EUR", 2, 0},
			"overflow 1":       {"EUR", "USD", math.MaxInt64, 0},
			"overflow 2":       {"EUR", "USD", math.MaxInt64, 1},
			"overflow 3":       {"EUR", "USD", math.MinInt64, 0},
			"overflow 4":       {"EUR", "USD", math.MinInt64, 1},
			"overflow 5":       {"EUR", "OMR", math.MaxInt64, 0},
			"overflow 6":       {"EUR", "OMR", math.MaxInt64, 1},
			"overflow 7":       {"EUR", "OMR", math.MaxInt64, 2},
			"overflow 8":       {"EUR", "OMR", math.MinInt64, 0},
			"overflow 9":       {"EUR", "OMR", math.MinInt64, 1},
			"overflow 10":      {"EUR", "OMR", math.MinInt64, 2},
		}
		for _, tt := range tests {
			_, err := NewExchRate(tt.base, tt.quote, tt.coef, tt.scale)
			if err == nil {
				t.Errorf("NewExchRate(%q, %q, %v, %v) did not fail", tt.base, tt.quote, tt.coef, tt.scale)
			}
		}
	})
}

func TestNewExchRateFromInt64(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			base, quote string
			whole, frac int64
			scale       int
			want        string
		}{
			{"EUR", "USD", 1, 1, 1, "1.10"},
			{"EUR", "USD", 1, 1, 2, "1.01"},
			{"EUR", "USD", 1, 1, 3, "1.001"},
			{"EUR", "USD", 1, 100000000, 9, "1.10"},
			{"EUR", "USD", 1, 1, 18, "1.000000000000000001"},
			{"EUR", "USD", 1, 1, 19, "1.000000000000000000"},
			{"EUR", "JPY", math.MaxInt64, math.MaxInt32, 10, "9223372036854775807"},
			{"EUR", "JPY", math.MaxInt64, math.MaxInt64, 19, "9223372036854775808"},
		}
		for _, tt := range tests {
			got, err := NewExchRateFromInt64(tt.base, tt.quote, tt.whole, tt.frac, tt.scale)
			if err != nil {
				t.Errorf("NewExchRateFromInt64(%q, %q, %v, %v, %v) failed: %v", tt.base, tt.quote, tt.whole, tt.frac, tt.scale, err)
				continue
			}
			want := MustParseExchRate(tt.base, tt.quote, tt.want)
			if got != want {
				t.Errorf("NewExchRateFromInt64(%q, %q, %v, %v, %v) = %q, want %q", tt.base, tt.quote, tt.whole, tt.frac, tt.scale, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string]struct {
			base, quote string
			whole, frac int64
			scale       int
		}{
			"quote currency 1":  {"EUR", "UUU", 1, 0, 0},
			"base currency 1":   {"EEE", "USD", 1, 0, 0},
			"different signs 1": {"EUR", "USD", -1, 1, 1},
			"different signs 2": {"EUR", "USD", 1, -1, 1},
			"fraction range 1":  {"EUR", "USD", 1, 1, 0},
			"scale range 1":     {"EUR", "USD", 1, 1, -1},
			"scale range 2":     {"EUR", "USD", 1, 0, -1},
			"scale range 3":     {"EUR", "USD", 1, 1, 20},
			"scale range 4":     {"EUR", "USD", 1, 0, 20},
			"overflow 1":        {"EUR", "USD", 100000000000000000, 100000000000000000, 18},
			"overflow 2":        {"EUR", "USD", 999999999999999999, 9, 1},
			"overflow 3":        {"EUR", "USD", 999999999999999999, 99, 2},
			"overflow 4":        {"EUR", "USD", math.MaxInt64, math.MaxInt32, 10},
			"overflow 5":        {"EUR", "OMR", math.MaxInt64, math.MaxInt32, 10},
		}
		for name, tt := range tests {
			t.Run(name, func(t *testing.T) {
				_, err := NewExchRateFromInt64(tt.base, tt.quote, tt.whole, tt.frac, tt.scale)
				if err == nil {
					t.Errorf("NewExchRateFromInt64(%q, %q, %v, %v, %v) did not fail", tt.base, tt.quote, tt.whole, tt.frac, tt.scale)
				}
			})
		}
	})
}

func TestNewExchRateFromFloat64(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			base, quote string
			f           float64
			want        string
		}{
			{"EUR", "USD", 1e-19, "0.0000000000000000001"},
			{"EUR", "USD", 1e-5, "0.00001"},
			{"EUR", "USD", 1e-4, "0.0001"},
			{"EUR", "USD", 1e-3, "0.001"},
			{"EUR", "USD", 1e-2, "0.01"},
			{"EUR", "USD", 1e-1, "0.1"},
			{"EUR", "USD", 1e0, "1"},
			{"EUR", "USD", 1e1, "10"},
			{"EUR", "USD", 1e2, "100"},
			{"EUR", "USD", 1e3, "1000"},
			{"EUR", "USD", 1e4, "10000"},
			{"EUR", "USD", 1e5, "100000"},
			{"EUR", "JPY", 1e18, "1000000000000000000"},
			{"EUR", "USD", 1e16, "10000000000000000"},
			{"EUR", "OMR", 1e15, "1000000000000000"},
		}
		for _, tt := range tests {
			got, err := NewExchRateFromFloat64(tt.base, tt.quote, tt.f)
			if err != nil {
				t.Errorf("NewExchRateFromFloat64(%q, %q, %v) failed: %v", tt.base, tt.quote, tt.f, err)
				continue
			}
			want := MustParseExchRate(tt.base, tt.quote, tt.want)
			if got != want {
				t.Errorf("NewExchRateFromFloat64(%q, %q, %v) = %q, want %q", tt.base, tt.quote, tt.f, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string]struct {
			base, quote string
			f           float64
		}{
			"base currency 1":  {"EEE", "USD", 1},
			"quote currency 1": {"EUR", "ZZZ", 1},
			"overflow 1":       {"EUR", "JPY", 1e19},
			"overflow 2":       {"EUR", "USD", 1e17},
			"overflow 3":       {"EUR", "OMR", 1e16},
			"special value 1":  {"EUR", "USD", math.NaN()},
			"special value 2":  {"EUR", "USD", math.Inf(1)},
			"special value 3":  {"EUR", "USD", math.Inf(-1)},
		}
		for name, tt := range tests {
			t.Run(name, func(t *testing.T) {
				_, err := NewExchRateFromFloat64(tt.base, tt.quote, tt.f)
				if err == nil {
					t.Errorf("NewExchRateFromFloat64(%q, %q, %v) did not fail", tt.base, tt.quote, tt.f)
				}
			})
		}
	})
}

func TestNewExchRateFromDecimal(t *testing.T) {
	tests := []struct {
		b, q   Currency
		r      string
		wantOk bool
	}{
		{USD, EUR, "1.2000", true},
		{USD, EUR, "-1.2000", false},
		{USD, USD, "0.9999", false},
		{USD, USD, "1.0000", true},
		{USD, USD, "1.0001", false},
		{USD, JPY, "1000000000000000000", true},
		{USD, EUR, "100000000000000000", false},
		{USD, OMR, "10000000000000000", false},
	}
	for _, tt := range tests {
		rate := decimal.MustParse(tt.r)
		_, err := NewExchRateFromDecimal(tt.b, tt.q, rate)
		if !tt.wantOk && err == nil {
			t.Errorf("NewExchRateFromDecimal(%v, %v, %v) did not fail", tt.b, tt.q, rate)
		}
		if tt.wantOk && err != nil {
			t.Errorf("NewExchRateFromDecimal(%v, %v, %v) failed: %v", tt.b, tt.q, rate, err)
		}
	}
}

func TestParseExchRate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			m, n, d             string
			wantBase, wantQuote Currency
			wantCoef            int64
			wantScale           int
		}{
			{"USD", "JPY", "132", USD, JPY, 132, 0},
			{"USD", "EUR", "1.25", USD, EUR, 125, 2},
			{"USD", "OMR", "0.38", USD, OMR, 380, 3},
		}
		for _, tt := range tests {
			got, err := ParseExchRate(tt.m, tt.n, tt.d)
			if err != nil {
				t.Errorf("ParseExchRate(%q, %q, %q) failed: %v", tt.m, tt.n, tt.d, err)
				continue
			}
			wantRate, err := decimal.New(tt.wantCoef, tt.wantScale)
			if err != nil {
				t.Errorf("decimal.New(%v, %v) failed: %v", tt.wantCoef, tt.wantScale, err)
				continue
			}
			want, err := NewExchRateFromDecimal(tt.wantBase, tt.wantQuote, wantRate)
			if err != nil {
				t.Errorf("NewExchRate(%v, %v, %v) failed: %v", tt.wantBase, tt.wantQuote, wantRate, err)
				continue
			}
			if got != want {
				t.Errorf("ParseExchRate(%q, %q, %q) = %q, want %q", tt.m, tt.n, tt.d, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string]struct {
			m, n, d string
		}{
			"no data":          {"", "", ""},
			"base currency 1":  {"AAA", "USD", "30000"},
			"quote currency 1": {"USD", "AAA", "0.00003"},
			"invalid rate 1":   {"USD", "EUR", "x.0000"},
			"invalid rate 2":   {"USD", "USD", "0.9999"},
			"invalid rate 3":   {"USD", "EUR", "0.0"},
			"invalid rate 4":   {"USD", "EUR", "-0.9999"},
		}
		for _, tt := range tests {
			_, err := ParseExchRate(tt.m, tt.n, tt.d)
			if err == nil {
				t.Errorf("ParseExchRate(%q, %q, %q) did not fail", tt.m, tt.n, tt.d)
			}
		}
	})
}

func TestExchangeRate_Mul(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			m, n, d, e, want string
		}{
			// There are few test cases here since Decimal.Mul
			// is already tested by the cases in the decimal package.
			{"USD", "EUR", "2", "2", "4"},
			{"USD", "EUR", "2", "3", "6"},
			{"USD", "EUR", "5", "1", "5"},
			{"USD", "EUR", "5", "2", "10"},
			{"USD", "EUR", "1.20", "2", "2.40"},
			{"USD", "EUR", "5.09", "7.1", "36.139"},
			{"USD", "EUR", "2.5", "4", "10.0"},
			{"USD", "EUR", "2.50", "4", "10.00"},
			{"USD", "EUR", "0.70", "1.05", "0.7350"},
		}
		for _, tt := range tests {
			r := MustParseExchRate(tt.m, tt.n, tt.d)
			e := decimal.MustParse(tt.e)
			got, err := r.Mul(e)
			if err != nil {
				t.Errorf("%q.Mul(%q) failed: %v", r, e, err)
				continue
			}
			want := MustParseExchRate(tt.m, tt.n, tt.want)
			if got != want {
				t.Errorf("%q.Mul(%q) = %q, want %q", r, e, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string]struct {
			m, n, d, e string
		}{
			"overflow 1": {"USD", "EUR", "0.9", "1000000000000000000"},
			"overflow 2": {"USD", "EUR", "10000000000000000", "10"},
			"factor 1":   {"USD", "EUR", "0.9", "0.0"},
			"factor 2":   {"USD", "EUR", "0.9", "-0.1"},
		}
		for _, tt := range tests {
			a := MustParseExchRate(tt.m, tt.n, tt.d)
			e := decimal.MustParse(tt.e)
			_, err := a.Mul(e)
			if err == nil {
				t.Errorf("%q.Mul(%q) did not fail", a, e)
			}
		}
	})
}

func TestExchangeRate_Inv(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			m, n, d, want string
		}{
			{"USD", "EUR", "0.5", "2"},
		}
		for _, tt := range tests {
			r := MustParseExchRate(tt.m, tt.n, tt.d)
			got, err := r.Inv()
			if err != nil {
				t.Errorf("%q.Inv() failed: %v", r, err)
				continue
			}
			want := MustParseExchRate(tt.n, tt.m, tt.want)
			if got != want {
				t.Errorf("%q.Inv() = %q, want %q", r, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		r := ExchangeRate{}
		_, err := r.Inv()
		if err == nil {
			t.Errorf("%q.Inv() did not fail", r)
		}
	})
}

func TestExchangeRate_Conv(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			m, n, r, d, wantDir, wantRev string
		}{
			{"JPY", "USD", "0.0075", "100", "0.7500", "13333.33333333333333"},
			{"EUR", "USD", "1.0995", "100.00", "109.950000", "90.95043201455206912"},
			{"OMR", "USD", "2.59765", "100.000", "259.76500000", "38.4963332242603892"},
		}
		for _, tt := range tests {
			r := MustParseExchRate(tt.m, tt.n, tt.r)

			// Direct conversion
			a := MustParseAmount(tt.m, tt.d)
			got, err := r.Conv(a)
			if err != nil {
				t.Errorf("%q.Conv(%q) failed: %v", r, a, err)
				continue
			}
			want := MustParseAmount(tt.n, tt.wantDir)
			if got != want {
				t.Errorf("%q.Conv(%q) = %q, want %q", r, a, got, want)
			}

			// Reverse conversion
			a = MustParseAmount(tt.n, tt.d)
			got, err = r.Conv(a)
			if err != nil {
				t.Errorf("%q.Conv(%q) failed: %v", r, a, err)
				continue
			}
			want = MustParseAmount(tt.m, tt.wantRev)
			if got != want {
				t.Errorf("%q.Conv(%q) = %q, want %q", r, a, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string]struct {
			m, n, r, c, d string
		}{
			"currency 1": {"USD", "EUR", "1.2000", "JPY", "100"},
			"currency 2": {"XXX", "EUR", "1.2000", "XXX", "100"},
			"overflow 1": {"USD", "JPY", "1000.00", "USD", "10000000000000000.00"},
			"overflow 2": {"USD", "EUR", "10.0000", "USD", "10000000000000000.00"},
			"overflow 3": {"USD", "OMR", "10.00000", "USD", "1000000000000000.00"},
		}
		for _, tt := range tests {
			r := MustParseExchRate(tt.m, tt.n, tt.r)
			a := MustParseAmount(tt.c, tt.d)
			_, err := r.Conv(a)
			if err == nil {
				t.Errorf("%q.Conv(%q) did not fail", r, a)
			}
		}
	})
}

func TestExchangeRate_Format(t *testing.T) {
	tests := []struct {
		m, n, d, format, want string
	}{
		// %T verb
		{"USD", "EUR", "100.0000", "%T", "money.ExchangeRate"},
		// %q verb
		{"USD", "EUR", "100.0000", "%q", "\"USD/EUR 100.0000\""},
		{"USD", "EUR", "100.0000", "%+q", "\"USD/EUR 100.0000\""},  // '+' is ignored
		{"USD", "EUR", "100.0000", "% q", "\"USD/EUR 100.0000\""},  // ' ' is ignored
		{"USD", "EUR", "100.0000", "%.6q", "\"USD/EUR 100.0000\""}, // precision is ignored
		{"USD", "EUR", "100.0000", "%16q", "\"USD/EUR 100.0000\""},
		{"USD", "EUR", "100.0000", "%17q", "\"USD/EUR 100.0000\""},
		{"USD", "EUR", "100.0000", "%18q", "\"USD/EUR 100.0000\""},
		{"USD", "EUR", "100.0000", "%19q", " \"USD/EUR 100.0000\""},
		{"USD", "EUR", "100.0000", "%019q", "\"USD/EUR 0100.0000\""},
		{"USD", "EUR", "100.0000", "%+19q", " \"USD/EUR 100.0000\""}, // '+' is ignored
		{"USD", "EUR", "100.0000", "%-19q", "\"USD/EUR 100.0000\" "},
		{"USD", "EUR", "100.0000", "%+-021q", "\"USD/EUR 100.0000\"   "}, // '+' and '0' are ignored
		// %s verb
		{"USD", "EUR", "100.0000", "%s", "USD/EUR 100.0000"},
		{"USD", "EUR", "100.0000", "%+s", "USD/EUR 100.0000"},  // '+' is ignored
		{"USD", "EUR", "100.0000", "% s", "USD/EUR 100.0000"},  // ' ' is ignored
		{"USD", "EUR", "100.0000", "%.6s", "USD/EUR 100.0000"}, // precision is ignored
		{"USD", "EUR", "100.0000", "%16s", "USD/EUR 100.0000"},
		{"USD", "EUR", "100.0000", "%17s", " USD/EUR 100.0000"},
		{"USD", "EUR", "100.0000", "%18s", "  USD/EUR 100.0000"},
		{"USD", "EUR", "100.0000", "%19s", "   USD/EUR 100.0000"},
		{"USD", "EUR", "100.0000", "%019s", "USD/EUR 000100.0000"},
		{"USD", "EUR", "100.0000", "%+19s", "   USD/EUR 100.0000"}, // '+' is ignored
		{"USD", "EUR", "100.0000", "%-19s", "USD/EUR 100.0000   "},
		{"USD", "EUR", "100.0000", "%+-021s", "USD/EUR 100.0000     "}, // '+' and '0' are ignored
		// %v verb
		{"USD", "EUR", "100.0000", "%v", "USD/EUR 100.0000"},
		{"USD", "EUR", "100.0000", "%+v", "USD/EUR 100.0000"},  // '+' is ignored
		{"USD", "EUR", "100.0000", "% v", "USD/EUR 100.0000"},  // ' ' is ignored
		{"USD", "EUR", "100.0000", "%.6v", "USD/EUR 100.0000"}, // precision is ignored
		{"USD", "EUR", "100.0000", "%16v", "USD/EUR 100.0000"},
		{"USD", "EUR", "100.0000", "%17v", " USD/EUR 100.0000"},
		{"USD", "EUR", "100.0000", "%18v", "  USD/EUR 100.0000"},
		{"USD", "EUR", "100.0000", "%19v", "   USD/EUR 100.0000"},
		{"USD", "EUR", "100.0000", "%019v", "USD/EUR 000100.0000"},
		{"USD", "EUR", "100.0000", "%+19v", "   USD/EUR 100.0000"}, // '+' is ignored
		{"USD", "EUR", "100.0000", "%-19v", "USD/EUR 100.0000   "},
		{"USD", "EUR", "100.0000", "%+-021v", "USD/EUR 100.0000     "}, // '+' and '0' are ignored
		// %f verb
		{"JPY", "EUR", "0.01", "%f", "0.01"},
		{"JPY", "EUR", "100.00", "%f", "100.00"},
		{"OMR", "EUR", "0.01", "%f", "0.01"},
		{"OMR", "EUR", "100.00", "%f", "100.00"},
		{"USD", "EUR", "0.01", "%f", "0.01"},
		{"USD", "EUR", "100.00", "%f", "100.00"},
		{"USD", "EUR", "9.996208266660", "%.1f", "10.00"},
		{"USD", "EUR", "0.9996208266660", "%.1f", "1.00"},
		{"USD", "EUR", "0.09996208266660", "%.1f", "0.10"},
		{"USD", "EUR", "0.009996208266660", "%.1f", "0.01"},
		{"USD", "EUR", "0.0009996208266660", "%.1f", "0.00"},
		{"USD", "EUR", "9.996208266660", "%.4f", "9.9962"},
		{"USD", "EUR", "0.9996208266660", "%.4f", "0.9996"},
		{"USD", "EUR", "0.09996208266660", "%.4f", "0.1000"},
		{"USD", "EUR", "0.009996208266660", "%.4f", "0.0100"},
		{"USD", "EUR", "100.0000", "%+f", "100.0000"}, // '+' is ignored
		{"USD", "EUR", "100.0000", "% f", "100.0000"}, // ' ' is ignored
		{"USD", "EUR", "100.0000", "%.3f", "100.000"},
		{"USD", "EUR", "100.0000", "%.4f", "100.0000"},
		{"USD", "EUR", "100.0000", "%.5f", "100.00000"},
		{"USD", "EUR", "100.0000", "%.6f", "100.000000"},
		{"USD", "EUR", "100.0000", "%.7f", "100.0000000"},
		{"USD", "EUR", "100.0000", "%.8f", "100.00000000"},
		{"USD", "EUR", "100.0000", "%9f", " 100.0000"},
		{"USD", "EUR", "100.0000", "%10f", "  100.0000"},
		{"USD", "EUR", "100.0000", "%11f", "   100.0000"},
		{"USD", "EUR", "100.0000", "%12f", "    100.0000"},
		{"USD", "EUR", "100.0000", "%013f", "00000100.0000"},
		{"USD", "EUR", "100.0000", "%+13f", "     100.0000"}, // '+' is ignored
		{"USD", "EUR", "100.0000", "%-13f", "100.0000     "},
		// %b verb
		{"USD", "EUR", "100.00", "%b", "USD"},
		{"USD", "EUR", "100.00", "%+b", "USD"}, // '+' is ignored
		{"USD", "EUR", "100.00", "% b", "USD"}, // ' ' is ignored
		{"USD", "EUR", "100.00", "%#b", "USD"}, // '#' is ignored
		{"USD", "EUR", "100.00", "%9b", "      USD"},
		{"USD", "EUR", "100.00", "%09b", "      USD"}, // '0' is ignored
		{"USD", "EUR", "100.00", "%#9b", "      USD"}, // '#' is ignored
		{"USD", "EUR", "100.00", "%-9b", "USD      "},
		{"USD", "EUR", "100.00", "%-#9b", "USD      "}, // '#' is ignored
		// %c verb
		{"USD", "EUR", "100.00", "%c", "EUR"},
		{"USD", "EUR", "100.00", "%+c", "EUR"}, // '+' is ignored
		{"USD", "EUR", "100.00", "% c", "EUR"}, // ' ' is ignored
		{"USD", "EUR", "100.00", "%#c", "EUR"}, // '#' is ignored
		{"USD", "EUR", "100.00", "%9c", "      EUR"},
		{"USD", "EUR", "100.00", "%09c", "      EUR"}, // '0' is ignored
		{"USD", "EUR", "100.00", "%#9c", "      EUR"}, // '#' is ignored
		{"USD", "EUR", "100.00", "%-9c", "EUR      "},
		{"USD", "EUR", "100.00", "%-#9c", "EUR      "}, // '#' is ignored
		// wrong verbs
		{"USD", "EUR", "12.3400", "%d", "%!d(money.ExchangeRate=USD/EUR 12.3400)"},
		{"USD", "EUR", "12.3400", "%e", "%!e(money.ExchangeRate=USD/EUR 12.3400)"},
		{"USD", "EUR", "12.3400", "%E", "%!E(money.ExchangeRate=USD/EUR 12.3400)"},
		{"USD", "EUR", "12.3400", "%g", "%!g(money.ExchangeRate=USD/EUR 12.3400)"},
		{"USD", "EUR", "12.3400", "%G", "%!G(money.ExchangeRate=USD/EUR 12.3400)"},
		{"USD", "EUR", "12.3400", "%x", "%!x(money.ExchangeRate=USD/EUR 12.3400)"},
		{"USD", "EUR", "12.3400", "%X", "%!X(money.ExchangeRate=USD/EUR 12.3400)"},
	}
	for _, tt := range tests {
		r := MustParseExchRate(tt.m, tt.n, tt.d)
		got := fmt.Sprintf(tt.format, r)
		if got != tt.want {
			t.Errorf("fmt.Sprintf(%q, %q) = %q, want %q", tt.format, r, got, tt.want)
		}
	}
}

func TestExchangeRate_Ceil(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			m, n, d string
			scale   int
			want    string
		}{
			{"USD", "EUR", "0.8000", 2, "0.80"},
		}
		for _, tt := range tests {
			r := MustParseExchRate(tt.m, tt.n, tt.d)
			got, err := r.Ceil(tt.scale)
			if err != nil {
				t.Errorf("%q.Ceil(%v) failed: %v", r, tt.scale, err)
				continue
			}
			want := MustParseExchRate(tt.m, tt.n, tt.want)
			if got != want {
				t.Errorf("%q.Ceil(%v) = %q, want %q", r, tt.scale, got, want)
			}
		}
	})
}

func TestExchangeRate_Floor(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			m, n, d string
			scale   int
			want    string
		}{
			{"USD", "EUR", "0.8000", 1, "0.80"},
			{"USD", "EUR", "0.0800", 2, "0.08"},
		}
		for _, tt := range tests {
			r := MustParseExchRate(tt.m, tt.n, tt.d)
			got, err := r.Floor(tt.scale)
			if err != nil {
				t.Errorf("%q.Floor(%v) failed: %v", r, tt.scale, err)
				continue
			}
			want := MustParseExchRate(tt.m, tt.n, tt.want)
			if got != want {
				t.Errorf("%q.Floor(%v) = %q, want %q", r, tt.scale, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string]struct {
			m, n, d string
			scale   int
		}{
			"zero rate 0": {"USD", "EUR", "0.8000", 0},
			"zero rate 1": {"USD", "EUR", "0.0800", 1},
			"zero rate 2": {"USD", "EUR", "0.0080", 2},
			"zero rate 3": {"USD", "EUR", "0.0008", 3},
		}
		for _, tt := range tests {
			r := MustParseExchRate(tt.m, tt.n, tt.d)
			_, err := r.Floor(tt.scale)
			if err == nil {
				t.Errorf("%q.Floor(%v) did not fail", r, tt.scale)
			}
		}
	})
}

func TestExchangeRate_Trunc(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			m, n, d string
			scale   int
			want    string
		}{
			{"USD", "EUR", "0.8000", 2, "0.80"},
			{"USD", "EUR", "0.0800", 2, "0.08"},
		}
		for _, tt := range tests {
			r := MustParseExchRate(tt.m, tt.n, tt.d)
			got, err := r.Trunc(tt.scale)
			if err != nil {
				t.Errorf("%q.Trunc(%v) failed: %v", r, tt.scale, err)
				continue
			}
			want := MustParseExchRate(tt.m, tt.n, tt.want)
			if got != want {
				t.Errorf("%q.Trunc(%v) = %q, want %q", r, tt.scale, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string]struct {
			m, n, d string
			scale   int
		}{
			"zero rate 1": {"USD", "EUR", "0.0080", 2},
			"zero rate 2": {"USD", "EUR", "0.0008", 3},
		}
		for _, tt := range tests {
			r := MustParseExchRate(tt.m, tt.n, tt.d)
			_, err := r.Trunc(tt.scale)
			if err == nil {
				t.Errorf("%q.Trunc(%v) did not fail", r, tt.scale)
			}
		}
	})
}

func TestExchangeRate_Round(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			m, n, d string
			scale   int
			want    string
		}{
			{"USD", "EUR", "0.8000", 2, "0.80"},
			{"USD", "EUR", "0.0800", 2, "0.08"},
			{"USD", "EUR", "0.0080", 2, "0.01"},
		}
		for _, tt := range tests {
			r := MustParseExchRate(tt.m, tt.n, tt.d)
			got, err := r.Round(tt.scale)
			if err != nil {
				t.Errorf("%q.Round(%v) failed: %v", r, tt.scale, err)
				continue
			}
			want := MustParseExchRate(tt.m, tt.n, tt.want)
			if got != want {
				t.Errorf("%q.Round(%v) = %q, want %q", r, tt.scale, got, want)
			}
		}
	})
	t.Run("error", func(t *testing.T) {
		tests := map[string]struct {
			m, n, d string
			scale   int
		}{
			"zero rate 1": {"USD", "EUR", "0.0050", 2},
			"zero rate 2": {"USD", "EUR", "0.0005", 3},
		}
		for _, tt := range tests {
			r := MustParseExchRate(tt.m, tt.n, tt.d)
			_, err := r.Round(tt.scale)
			if err == nil {
				t.Errorf("%q.Round(%v) did not fail", r, tt.scale)
			}
		}
	})
}

func TestExchangeRate_Rescale(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			m, n, d string
			scale   int
			want    string
		}{
			// Padding
			{"EUR", "JPY", "1", 0, "1"},
			{"EUR", "JPY", "1", 1, "1.0"},
			{"EUR", "JPY", "1", 2, "1.00"},
			{"EUR", "JPY", "1", 3, "1.000"},
			{"EUR", "USD", "1", 0, "1.00"},
			{"EUR", "USD", "1", 1, "1.00"},
			{"EUR", "USD", "1", 2, "1.00"},
			{"EUR", "USD", "1", 3, "1.000"},
			{"EUR", "OMR", "1", 0, "1.000"},
			{"EUR", "OMR", "1", 1, "1.000"},
			{"EUR", "OMR", "1", 2, "1.000"},
			{"EUR", "OMR", "1", 3, "1.000"},
			{"EUR", "USD", "1", 17, "1.00000000000000000"},
			{"EUR", "USD", "1", 18, "1.000000000000000000"},
			{"EUR", "USD", "1", 17, "1.00000000000000000"},
			{"EUR", "USD", "1", 18, "1.000000000000000000"},

			// Half-to-even rounding
			{"EUR", "USD", "0.0051", 2, "0.01"},
			{"EUR", "USD", "0.0149", 2, "0.01"},
			{"EUR", "USD", "0.0151", 2, "0.02"},
			{"EUR", "USD", "0.0150", 2, "0.02"},
			{"EUR", "USD", "0.0250", 2, "0.02"},
			{"EUR", "USD", "0.0350", 2, "0.04"},
			{"EUR", "USD", "3.0448", 2, "3.04"},
			{"EUR", "USD", "3.0450", 2, "3.04"},
			{"EUR", "USD", "3.0452", 2, "3.05"},
			{"EUR", "USD", "3.0956", 2, "3.10"},
		}
		for _, tt := range tests {
			r := MustParseExchRate(tt.m, tt.n, tt.d)
			got, err := r.Rescale(tt.scale)
			if err != nil {
				t.Errorf("%q.Rescale(%v) failed: %v", r, tt.scale, err)
				continue
			}
			want := MustParseExchRate(tt.m, tt.n, tt.want)
			if got != want {
				t.Errorf("%q.Rescale(%v) = %q, want %q", r, tt.scale, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string]struct {
			m, n, d string
			scale   int
		}{
			"zero rate 1": {"USD", "EUR", "0.0050", 2},
			"zero rate 2": {"USD", "EUR", "0.0005", 3},
		}
		for _, tt := range tests {
			r := MustParseExchRate(tt.m, tt.n, tt.d)
			_, err := r.Rescale(tt.scale)
			if err == nil {
				t.Errorf("%q.Rescale(%v) did not fail", r, tt.scale)
			}
		}
	})
}

func TestExchangeRate_Quantize(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			m, n, d, e string
			want       string
		}{
			{"EUR", "JPY", "1", "0.01", "1.00"},
			{"EUR", "JPY", "1", "0.001", "1.000"},
			{"EUR", "JPY", "1", "0.0001", "1.0000"},
			{"EUR", "JPY", "1", "0.00001", "1.00000"},
			{"EUR", "JPY", "1", "0.000001", "1.000000"},
		}
		for _, tt := range tests {
			r := MustParseExchRate(tt.m, tt.n, tt.d)
			q := MustParseExchRate(tt.m, tt.n, tt.e)
			got, err := r.Quantize(q)
			if err != nil {
				t.Errorf("%q.Quantize(%q) failed: %v", r, q, err)
				continue
			}
			want := MustParseExchRate(tt.m, tt.n, tt.want)
			if got != want {
				t.Errorf("%q.Quantize(%q) = %q, want %q", r, q, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string]struct {
			m, n, d, e string
		}{
			"zero rate 1": {"USD", "EUR", "0.0050", "0.01"},
			"zero rate 2": {"USD", "EUR", "0.0005", "0.001"},
		}
		for _, tt := range tests {
			r := MustParseExchRate(tt.m, tt.n, tt.d)
			q := MustParseExchRate(tt.m, tt.n, tt.e)
			_, err := r.Quantize(q)
			if err == nil {
				t.Errorf("%q.Quantize(%q) did not fail", r, q)
			}
		}
	})
}
