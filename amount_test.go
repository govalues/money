package money

import (
	"fmt"
	"reflect"
	"testing"
	"unsafe"

	"github.com/govalues/decimal"
)

func TestAmount_ZeroValue(t *testing.T) {
	got := Amount{}
	want := MustParseAmount("XXX", "0")
	if got != want {
		t.Errorf("Amount{} = %q, want %q", got, want)
	}
}

func TestAmount_Sizeof(t *testing.T) {
	a := Amount{}
	got := unsafe.Sizeof(a)
	want := uintptr(24)
	if got != want {
		t.Errorf("unsafe.Sizeof(%q) = %v, want %v", a, got, want)
	}
}

func TestNewAmount(t *testing.T) {
	tests := []struct {
		c      Currency
		a      string
		wantOk bool
	}{
		{JPY, "1", true},
		{USD, "1.00", true},
		{OMR, "1.000", true},
		{JPY, "1000000000000000000", true},
		{USD, "100000000000000000", false},
		{OMR, "10000000000000000", false},
	}
	for _, tt := range tests {
		amount := decimal.MustParse(tt.a)
		_, err := NewAmount(tt.c, amount)
		if !tt.wantOk && err == nil {
			t.Errorf("NewAmount(%v, %v) did not fail", tt.c, amount)
		}
		if tt.wantOk && err != nil {
			t.Errorf("NewAmount(%v, %v) failed: %v", tt.c, amount, err)
		}
	}
}

func TestParseAmount(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			c, a      string
			wantCurr  Currency
			wantCoef  int64
			wantScale int
		}{
			// There are few test cases here since decimal.Parse and money.ParseCurr
			// are already tested by the cases in their packages.
			{"USD", "1", USD, 100, 2},
			{"USD", "1.0", USD, 100, 2},
			{"USD", "1.00", USD, 100, 2},
			{"USD", "1.000", USD, 1000, 3},
			{"USD", "1.0000", USD, 10000, 4},
			{"USD", "1.00000", USD, 100000, 5},
		}
		for _, tt := range tests {
			got, err := ParseAmount(tt.c, tt.a)
			if err != nil {
				t.Errorf("ParseAmount(%q, %q) failed: %v", tt.c, tt.a, err)
				continue
			}
			wantAmount, err := decimal.New(tt.wantCoef, tt.wantScale)
			if err != nil {
				t.Errorf("decimal.New(%v, %v) failed: %v", tt.wantCoef, tt.wantScale, err)
				continue
			}
			want, err := NewAmount(tt.wantCurr, wantAmount)
			if err != nil {
				t.Errorf("NewAmount(%v, %v) failed: %v", tt.wantCurr, wantAmount, err)
				continue
			}
			if got != want {
				t.Errorf("ParseAmount(%q, %q) = %q, want %q", tt.c, tt.a, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string]struct {
			c, a string
		}{
			"currency 1": {"///", "1"},
			"currency 2": {"ZZZ", "1.0"},
			"overflow 1": {"USD", "100000000000000000"},
			"overflow 2": {"OMR", "10000000000000000"},
			"decimal 1":  {"USD", "abc"},
		}
		for name, tt := range tests {
			t.Run(name, func(t *testing.T) {
				_, err := ParseAmount(tt.c, tt.a)
				if err == nil {
					t.Errorf("ParseAmount(%q, %q) did not fail", tt.c, tt.a)
				}
			})
		}
	})
}

func TestAmount_MinorUnits(t *testing.T) {
	tests := []struct {
		c, a      string
		wantUnits int64
		wantOk    bool
	}{
		// Different signs
		{"USD", "-1", -100, true},
		{"USD", "0", 0, true},
		{"USD", "1", 100, true},

		// Different scales
		{"JPY", "1", 1, true},
		{"JPY", "1.0", 1, true},
		{"JPY", "1.00", 1, true},
		{"JPY", "1.000", 1, true},
		{"JPY", "1.0000", 1, true},
		{"USD", "1", 100, true},
		{"USD", "1.0", 100, true},
		{"USD", "1.00", 100, true},
		{"USD", "1.000", 100, true},
		{"USD", "1.0000", 100, true},
		{"OMR", "1", 1000, true},
		{"OMR", "1.0", 1000, true},
		{"OMR", "1.00", 1000, true},
		{"OMR", "1.000", 1000, true},
		{"OMR", "1.0000", 1000, true},

		// Rounding
		{"USD", "1", 100, true},
		{"USD", "1.5", 150, true},
		{"USD", "1.56", 156, true},
		{"USD", "1.567", 157, true},
		{"USD", "1.5678", 157, true},
		{"USD", "1.56789", 157, true},

		// Minimal value
		{"USD", "-92233720368547758.08", -9223372036854775808, true},
		{"USD", "-92233720368547758.09", 0, false},

		// Maximal value
		{"USD", "92233720368547758.07", 9223372036854775807, true},
		{"USD", "92233720368547758.08", 0, false},
	}
	for _, tt := range tests {
		a := MustParseAmount(tt.c, tt.a)
		gotUnits, gotOk := a.MinorUnits()
		if gotUnits != tt.wantUnits || gotOk != tt.wantOk {
			t.Errorf("%q.MinorUnits() = [%v %v], want [%v %v]", a, gotUnits, gotOk, tt.wantUnits, tt.wantOk)
		}
	}
}

func TestAmount_SameScaleAsCurr(t *testing.T) {
	tests := []struct {
		c, a string
		want bool
	}{
		{"USD", "1.00", true},
		{"USD", "1.000", false},
		{"USD", "1.0000", false},
		{"USD", "1.00000", false},
	}
	for _, tt := range tests {
		a := MustParseAmount(tt.c, tt.a)
		got := a.SameScaleAsCurr()
		if got != tt.want {
			t.Errorf("%q.SameScaleAsCurr() = %v, want %v", a, got, tt.want)
		}
	}
}

func MustParseAmountSlice(curr string, amounts []string) []Amount {
	res := make([]Amount, len(amounts))
	for i := 0; i < len(amounts); i++ {
		res[i] = MustParseAmount(curr, amounts[i])
	}
	return res
}

func TestAmount_Add(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			c, a, b, want string
		}{
			// There are few test cases here since Decimal.Add
			// is already tested by the cases in the decimal package.
			{"USD", "1", "1", "2"},
			{"USD", "2", "3", "5"},
			{"USD", "5.75", "3.3", "9.05"},
			{"USD", "5", "-3", "2"},
			{"USD", "-5", "-3", "-8"},
			{"USD", "-7", "2.5", "-4.5"},
			{"USD", "0.7", "0.3", "1.0"},
			{"USD", "1.25", "1.25", "2.50"},
			{"USD", "1.1", "0.11", "1.21"},
			{"USD", "1.234567890", "1.000000000", "2.234567890"},
			{"USD", "1.234567890", "1.000000110", "2.234568000"},
			{"USD", "99999999999999999.99", "0.004", "99999999999999999.99"},
			{"USD", "-99999999999999999.99", "-0.004", "-99999999999999999.99"},
			{"USD", "0.01", "-99999999999999999.99", "-99999999999999999.98"},
			{"USD", "99999999999999999.99", "-0.01", "99999999999999999.98"},
		}
		for _, tt := range tests {
			a := MustParseAmount(tt.c, tt.a)
			b := MustParseAmount(tt.c, tt.b)
			got, err := a.Add(b)
			if err != nil {
				t.Errorf("%q.Add(%q) failed: %v", a, b, err)
				continue
			}
			want := MustParseAmount(tt.c, tt.want)
			if got != want {
				t.Errorf("%q.Add(%q) = %q, want %q", a, b, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string]struct {
			ca, a, cb, b string
		}{
			"currency 1": {"USD", "1", "JPY", "1"},
			"currency 2": {"USD", "1", "EUR", "1"},
			"overflow 1": {"USD", "99999999999999999.99", "USD", "0.01"},
			"overflow 2": {"USD", "99999999999999999.99", "USD", "0.006"},
			"overflow 3": {"USD", "-99999999999999999.99", "USD", "-0.01"},
			"overflow 4": {"USD", "-99999999999999999.99", "USD", "-0.006"},
		}
		for _, tt := range tests {
			a := MustParseAmount(tt.ca, tt.a)
			b := MustParseAmount(tt.cb, tt.b)
			_, err := a.Add(b)
			if err == nil {
				t.Errorf("%q.Add(%q) did not fail", a, b)
			}
		}
	})
}

func TestAmount_FMA(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			c, a, e, b, want string
		}{
			// There are few test cases here since Decimal.FMA
			// is already tested by the cases in the decimal package.

			// Addition
			{"USD", "1", "1", "1", "2"},
			{"USD", "2", "1", "3", "5"},
			{"USD", "5.75", "1", "3.3", "9.05"},
			{"USD", "5", "1", "-3", "2"},
			{"USD", "-5", "1", "-3", "-8"},
			{"USD", "-7", "1", "2.5", "-4.5"},
			{"USD", "0.7", "1", "0.3", "1.0"},
			{"USD", "1.25", "1", "1.25", "2.50"},
			{"USD", "1.1", "1", "0.11", "1.21"},
			{"USD", "1.234567890", "1", "1.000000000", "2.234567890"},
			{"USD", "1.234567890", "1", "1.000000110", "2.234568000"},
			{"USD", "99999999999999999.99", "1", "0.004", "99999999999999999.99"},
			{"USD", "-99999999999999999.99", "1", "-0.004", "-99999999999999999.99"},
			{"USD", "0.01", "1", "-99999999999999999.99", "-99999999999999999.98"},
			{"USD", "99999999999999999.99", "1", "-0.01", "99999999999999999.98"},

			// Multiplication
			{"USD", "2", "2", "0", "4"},
			{"USD", "2", "3", "0", "6"},
			{"USD", "5", "1", "0", "5"},
			{"USD", "5", "2", "0", "10"},
			{"USD", "1.20", "2", "0", "2.40"},
			{"USD", "1.20", "0", "0", "0.00"},
			{"USD", "1.20", "-2", "0", "-2.40"},
			{"USD", "-1.20", "2", "0", "-2.40"},
			{"USD", "-1.20", "0", "0", "0.00"},
			{"USD", "-1.20", "-2", "0", "2.40"},
			{"USD", "5.09", "7.1", "0", "36.139"},
			{"USD", "2.5", "4", "0", "10.0"},
			{"USD", "2.50", "4", "0", "10.00"},
			{"USD", "0.70", "1.05", "0", "0.7350"},
		}
		for _, tt := range tests {
			a := MustParseAmount(tt.c, tt.a)
			e := decimal.MustParse(tt.e)
			b := MustParseAmount(tt.c, tt.b)
			got, err := a.FMA(e, b)
			if err != nil {
				t.Errorf("%q.FMA(%q, %q) failed: %v", a, e, b, err)
				continue
			}
			want := MustParseAmount(tt.c, tt.want)
			if got != want {
				t.Errorf("%q.FMA(%q, %q) = %q, want %q", a, e, b, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string]struct {
			ca, a, e, cb, b string
		}{
			"overflow 1": {"JPY", "1", "9999999999999999999", "JPY", "1"},
			"overflow 2": {"JPY", "1", "9999999999999999999", "JPY", "0.6"},
			"overflow 3": {"JPY", "1", "-9999999999999999999", "JPY", "-1"},
			"overflow 4": {"JPY", "1", "-9999999999999999999", "JPY", "-0.6"},
			"overflow 5": {"JPY", "10000000000", "1000000000", "JPY", "0"},
			"overflow 6": {"JPY", "1000000000000000000", "10", "JPY", "0"},
			"currency 1": {"JPY", "1", "1", "USD", "1"},
		}
		for _, tt := range tests {
			a := MustParseAmount(tt.ca, tt.a)
			e := decimal.MustParse(tt.e)
			b := MustParseAmount(tt.cb, tt.b)
			_, err := a.FMA(e, b)
			if err == nil {
				t.Errorf("%q.FMA(%q, %q) did not fail", a, e, b)
			}
		}
	})
}

func TestAmount_Rat(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			c, a, b, want string
		}{
			// There are few test cases here since Decimal.Quo
			// is already tested by the cases in the decimal package.
			{"USD", "1", "1", "1"},
			{"USD", "2", "1", "2"},
			{"USD", "1", "2", "0.5"},
			{"USD", "2", "2", "1"},
			{"USD", "0", "1", "0"},
			{"USD", "0", "2", "0"},
			{"USD", "1.5", "3", "0.5"},
			{"USD", "3", "3", "1"},
			{"USD", "2.4", "1", "2.4"},
			{"USD", "2.4", "-1", "-2.4"},
			{"USD", "-2.4", "1", "-2.4"},
			{"USD", "-2.4", "-1", "2.4"},
			{"USD", "2.40", "1", "2.4"},
			{"USD", "2.400", "1", "2.4"},
			{"USD", "2.4", "2", "1.2"},
			{"USD", "2.400", "2", "1.2"},
			{"USD", "99999999999999999.99", "1", "99999999999999999.99"},
			{"USD", "99999999999999999.99", "99999999999999999.99", "1"},
		}
		for _, tt := range tests {
			a := MustParseAmount(tt.c, tt.a)
			b := MustParseAmount(tt.c, tt.b)
			got, err := a.Rat(b)
			if err != nil {
				t.Errorf("%q.Rat(%q) failed: %v", a, b, err)
				continue
			}
			want := decimal.MustParse(tt.want)
			if got != want {
				t.Errorf("%q.Rat(%q) = %q, want %q", a, b, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string]struct {
			ca, a, cb, b string
		}{
			"overflow 1": {"USD", "10000000000000000", "USD", "0.001"},
			"zero 1":     {"USD", "1", "USD", "0"},
		}
		for _, tt := range tests {
			a := MustParseAmount(tt.ca, tt.a)
			b := MustParseAmount(tt.cb, tt.b)
			_, err := a.Rat(b)
			if err == nil {
				t.Errorf("%q.Rat(%q) did not fail", a, b)
			}
		}
	})
}

func TestAmount_Quo(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			c, a, e, want string
		}{
			// There are few test cases here since Decimal.Quo
			// is already tested by the cases in the decimal package.
			{"USD", "1", "1", "1"},
			{"USD", "2", "1", "2"},
			{"USD", "1", "2", "0.5"},
			{"USD", "2", "2", "1"},
			{"USD", "0", "1", "0"},
			{"USD", "0", "2", "0"},
			{"USD", "1.5", "3", "0.5"},
			{"USD", "3", "3", "1"},
			{"USD", "2.4", "1", "2.4"},
			{"USD", "2.4", "-1", "-2.4"},
			{"USD", "-2.4", "1", "-2.4"},
			{"USD", "-2.4", "-1", "2.4"},
			{"USD", "2.40", "1", "2.40"},
			{"USD", "2.400", "1", "2.400"},
			{"USD", "2.4", "2", "1.2"},
			{"USD", "2.400", "2", "1.200"},
			{"USD", "99999999999999999.99", "1", "99999999999999999.99"},
			{"USD", "99999999999999999.99", "99999999999999999.99", "1"},
		}
		for _, tt := range tests {
			a := MustParseAmount(tt.c, tt.a)
			e := decimal.MustParse(tt.e)
			got, err := a.Quo(e)
			if err != nil {
				t.Errorf("%q.Quo(%q) failed: %v", a, e, err)
				continue
			}
			want := MustParseAmount(tt.c, tt.want)
			if got != want {
				t.Errorf("%q.Quo(%q) = %q, want %q", a, e, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string]struct {
			c, a, e string
		}{
			"zero 1":     {"USD", "1", "0"},
			"overflow 1": {"USD", "99999999999999999", "0.1"},
		}
		for _, tt := range tests {
			a := MustParseAmount(tt.c, tt.a)
			e := decimal.MustParse(tt.e)
			_, err := a.Quo(e)
			if err == nil {
				t.Errorf("%q.Quo(%q) did not fail", a, e)
			}
		}
	})
}

func TestAmount_QuoRem(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			c, a, e, wantQuo, wantRem string
		}{
			{"USD", "1.00", "1", "1.00", "0.00"},
			{"USD", "2.00", "1", "2.00", "0.00"},
			{"USD", "1.00", "2", "0.50", "0.00"},
			{"USD", "2.00", "2", "1.00", "0.00"},
			{"USD", "0.00", "1", "0.00", "0.00"},
			{"USD", "1.510", "3", "0.50", "0.010"},
			{"USD", "3.333", "3", "1.11", "0.003"},
			{"USD", "2.401", "1", "2.40", "0.001"},
			{"USD", "2.401", "-1", "-2.40", "0.001"},
			{"USD", "-2.401", "1", "-2.40", "-0.001"},
			{"USD", "-2.401", "-1", "2.40", "-0.001"},
		}
		for _, tt := range tests {
			a := MustParseAmount(tt.c, tt.a)
			e := decimal.MustParse(tt.e)
			gotQuo, gotRem, err := a.QuoRem(e)
			if err != nil {
				t.Errorf("%q.QuoRem(%q) failed: %v", a, e, err)
				continue
			}
			wantQuo := MustParseAmount(tt.c, tt.wantQuo)
			wantRem := MustParseAmount(tt.c, tt.wantRem)
			if gotQuo != wantQuo || gotRem != wantRem {
				t.Errorf("%q.QuoRem(%q) = [%q %q], want [%q %q]", a, e, gotQuo, gotRem, wantQuo, wantRem)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string]struct {
			c, a, e string
		}{
			"zero 1":     {"USD", "1", "0"},
			"overflow 1": {"USD", "99999999999999999", "0.1"},
		}
		for _, tt := range tests {
			a := MustParseAmount(tt.c, tt.a)
			e := decimal.MustParse(tt.e)
			_, _, err := a.QuoRem(e)
			if err == nil {
				t.Errorf("%q.QuoRem(%q) did not fail", a, e)
			}
		}
	})
}

func TestAmount_Mul(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			c, a, e, want string
		}{
			// There are few test cases here since Decimal.Mul
			// is already tested by the cases in the decimal package.
			{"USD", "2", "2", "4"},
			{"USD", "2", "3", "6"},
			{"USD", "5", "1", "5"},
			{"USD", "5", "2", "10"},
			{"USD", "1.20", "2", "2.40"},
			{"USD", "1.20", "0", "0.00"},
			{"USD", "1.20", "-2", "-2.40"},
			{"USD", "-1.20", "2", "-2.40"},
			{"USD", "-1.20", "0", "0.00"},
			{"USD", "-1.20", "-2", "2.40"},
			{"USD", "5.09", "7.1", "36.139"},
			{"USD", "2.5", "4", "10.0"},
			{"USD", "2.50", "4", "10.00"},
			{"USD", "0.70", "1.05", "0.7350"},
		}
		for _, tt := range tests {
			a := MustParseAmount(tt.c, tt.a)
			e := decimal.MustParse(tt.e)
			got, err := a.Mul(e)
			if err != nil {
				t.Errorf("%q.Mul(%q) failed: %v", a, e, err)
				continue
			}
			want := MustParseAmount(tt.c, tt.want)
			if got != want {
				t.Errorf("%q.Mul(%q) = %q, want %q", a, e, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string]struct {
			c, a, e string
		}{
			"overflow 1": {"USD", "10000000000", "1000000000"},
			"overflow 2": {"USD", "10000000000000000", "1000"},
		}
		for _, tt := range tests {
			a := MustParseAmount(tt.c, tt.a)
			e := decimal.MustParse(tt.e)
			_, err := a.Mul(e)
			if err == nil {
				t.Errorf("%q.Mul(%q) did not fail", a, e)
			}
		}
	})
}

func TestAmount_Split(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			c, a  string
			parts int
			want  []string
		}{
			{"OMR", "0.0001", 3, []string{"0.0001", "0.0000", "0.0000"}},
			{"OMR", "-0.0001", 3, []string{"-0.0001", "0.0000", "0.0000"}},

			{"JPY", "0", 5, []string{"0", "0", "0", "0", "0"}},
			{"JPY", "0.0", 5, []string{"0.0", "0.0", "0.0", "0.0", "0.0"}},
			{"JPY", "0.00", 5, []string{"0.00", "0.00", "0.00", "0.00", "0.00"}},

			{"JPY", "0.001", 3, []string{"0.001", "0.000", "0.000"}},
			{"JPY", "0.01", 3, []string{"0.01", "0.00", "0.00"}},
			{"JPY", "0.1", 3, []string{"0.1", "0.0", "0.0"}},
			{"JPY", "1", 3, []string{"1", "0", "0"}},
			{"JPY", "1.0", 3, []string{"0.4", "0.3", "0.3"}},
			{"JPY", "1.00", 3, []string{"0.34", "0.33", "0.33"}},
			{"JPY", "1.000", 3, []string{"0.334", "0.333", "0.333"}},

			{"USD", "1.01", 1, []string{"1.01"}},
			{"USD", "1.01", 2, []string{"0.51", "0.50"}},
			{"USD", "1.01", 3, []string{"0.34", "0.34", "0.33"}},
			{"USD", "1.01", 4, []string{"0.26", "0.25", "0.25", "0.25"}},
			{"USD", "1.01", 5, []string{"0.21", "0.20", "0.20", "0.20", "0.20"}},
			{"USD", "1.01", 6, []string{"0.17", "0.17", "0.17", "0.17", "0.17", "0.16"}},

			{"USD", "-1.01", 1, []string{"-1.01"}},
			{"USD", "-1.01", 2, []string{"-0.51", "-0.50"}},
			{"USD", "-1.01", 3, []string{"-0.34", "-0.34", "-0.33"}},
			{"USD", "-1.01", 4, []string{"-0.26", "-0.25", "-0.25", "-0.25"}},
			{"USD", "-1.01", 5, []string{"-0.21", "-0.20", "-0.20", "-0.20", "-0.20"}},
			{"USD", "-1.01", 6, []string{"-0.17", "-0.17", "-0.17", "-0.17", "-0.17", "-0.16"}},
		}
		for _, tt := range tests {
			a := MustParseAmount(tt.c, tt.a)
			got, err := a.Split(tt.parts)
			if err != nil {
				t.Errorf("%q.Split(%v) failed: %v", a, tt.parts, err)
				continue
			}
			want := MustParseAmountSlice(tt.c, tt.want)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("%q.Split(%v) = %v, want %v", a, tt.parts, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		a := MustParseAmount("USD", "1")
		parts := -1
		_, err := a.Split(parts)
		if err == nil {
			t.Errorf("%q.Split(%v) did not fail", a, parts)
		}
	})
}

func TestAmount_String(t *testing.T) {
	tests := []struct {
		c, a, want string
	}{
		// Zeroes
		{"JPY", "0", "JPY 0"},
		{"JPY", "0.0", "JPY 0.0"},
		{"USD", "0", "USD 0.00"},
		{"USD", "0.0", "USD 0.00"},
		{"USD", "0.00", "USD 0.00"},
		{"USD", "0.000", "USD 0.000"},
		{"OMR", "0", "OMR 0.000"},
		{"OMR", "0.0", "OMR 0.000"},
		{"OMR", "0.00", "OMR 0.000"},
		{"OMR", "0.000", "OMR 0.000"},
		{"OMR", "0.0000", "OMR 0.0000"},

		// Negative
		{"USD", "-1", "USD -1.00"},
		{"USD", "1", "USD 1.00"},
	}
	for _, tt := range tests {
		a := MustParseAmount(tt.c, tt.a)
		got := a.String()
		if got != tt.want {
			t.Errorf("%q.String() = %q, want %q", a, got, tt.want)
		}
	}
}

func TestAmount_Format(t *testing.T) {
	tests := []struct {
		c, a, format, want string
	}{
		// %T verb
		{"USD", "100.00", "%T", "money.Amount"},
		// %q verb
		{"USD", "100.00", "%q", "\"USD 100.00\""},
		{"USD", "100.00", "%+q", "\"USD +100.00\""},
		{"USD", "100.00", "% q", "\"USD  100.00\""},
		{"USD", "100.00", "%.6q", "\"USD 100.00\""}, // precision is ignored
		{"USD", "100.00", "%10q", "\"USD 100.00\""},
		{"USD", "100.00", "%11q", "\"USD 100.00\""},
		{"USD", "100.00", "%12q", "\"USD 100.00\""},
		{"USD", "100.00", "%13q", " \"USD 100.00\""},
		{"USD", "100.00", "%013q", "\"USD 0100.00\""},
		{"USD", "100.00", "%+13q", "\"USD +100.00\""},
		{"USD", "100.00", "%-13q", "\"USD 100.00\" "},
		{"USD", "100.00", "%+-015q", "\"USD +100.00\"  "}, // '0' is ignored
		// %s verb
		{"USD", "100.00", "%s", "USD 100.00"},
		{"USD", "100.00", "%+s", "USD +100.00"},
		{"USD", "100.00", "% s", "USD  100.00"},
		{"USD", "100.00", "%.6s", "USD 100.00"}, // precision is ignored
		{"USD", "100.00", "%10s", "USD 100.00"},
		{"USD", "100.00", "%11s", " USD 100.00"},
		{"USD", "100.00", "%12s", "  USD 100.00"},
		{"USD", "100.00", "%13s", "   USD 100.00"},
		{"USD", "100.00", "%013s", "USD 000100.00"},
		{"USD", "100.00", "%+13s", "  USD +100.00"},
		{"USD", "100.00", "%-13s", "USD 100.00   "},
		{"USD", "100.00", "%+-015s", "USD +100.00    "}, // '0' is ignored
		// %v verb
		{"USD", "100.00", "%v", "USD 100.00"},
		{"USD", "100.00", "%+v", "USD +100.00"},
		{"USD", "100.00", "% v", "USD  100.00"},
		{"USD", "100.00", "%.6v", "USD 100.00"}, // precision is ignored
		{"USD", "100.00", "%10v", "USD 100.00"},
		{"USD", "100.00", "%11v", " USD 100.00"},
		{"USD", "100.00", "%12v", "  USD 100.00"},
		{"USD", "100.00", "%13v", "   USD 100.00"},
		{"USD", "100.00", "%013v", "USD 000100.00"},
		{"USD", "100.00", "%+13v", "  USD +100.00"},
		{"USD", "100.00", "%-13v", "USD 100.00   "},
		{"USD", "100.00", "%+-015v", "USD +100.00    "}, // '0' is ignored
		// %f verb
		{"JPY", "0.00", "%f", "0"}, // default precision is 0 (JPY scale)
		{"JPY", "0.01", "%f", "0"},
		{"JPY", "100.00", "%f", "100"},
		{"OMR", "0.00", "%f", "0.000"}, // default precision is 3 (OMR scale)
		{"OMR", "0.01", "%f", "0.010"},
		{"OMR", "100.00", "%f", "100.000"},
		{"USD", "0.00", "%f", "0.00"}, // default precision is 2 (USD scale)
		{"USD", "0.01", "%f", "0.01"},
		{"USD", "100.00", "%f", "100.00"},
		{"USD", "9.996208266660", "%f", "10.00"},
		{"USD", "0.9996208266660", "%f", "1.00"},
		{"USD", "0.09996208266660", "%f", "0.10"},
		{"USD", "0.009996208266660", "%f", "0.01"},
		{"USD", "100.00", "%+f", "+100.00"},
		{"USD", "100.00", "% f", " 100.00"},
		{"USD", "100.00", "%.1f", "100.00"}, // precision cannot be smaller than curr scale
		{"USD", "100.00", "%.2f", "100.00"},
		{"USD", "100.00", "%.3f", "100.000"},
		{"USD", "100.00", "%.4f", "100.0000"},
		{"USD", "100.00", "%.5f", "100.00000"},
		{"USD", "100.00", "%.6f", "100.000000"},
		{"USD", "100.00", "%7f", " 100.00"},
		{"USD", "100.00", "%8f", "  100.00"},
		{"USD", "100.00", "%9f", "   100.00"},
		{"USD", "100.00", "%10f", "    100.00"},
		{"USD", "100.00", "%010f", "0000100.00"},
		{"USD", "100.00", "%+10f", "   +100.00"},
		{"USD", "100.00", "%-10f", "100.00    "},
		// %d verb
		{"JPY", "0.01", "%d", "0"}, // rounded to 0 digits (JPY scale)
		{"JPY", "100.00", "%d", "100"},
		{"OMR", "0.01", "%d", "10"}, // rounded to 3 digits (OMR scale)
		{"OMR", "100.00", "%d", "100000"},
		{"USD", "0.01", "%d", "1"}, // rounded to 2 digits (USD scale)
		{"USD", "100.00", "%d", "10000"},
		{"USD", "9.996208266660", "%d", "1000"},
		{"USD", "0.9996208266660", "%d", "100"},
		{"USD", "0.09996208266660", "%d", "10"},
		{"USD", "0.009996208266660", "%d", "1"},
		{"USD", "100.00", "%+d", "+10000"},
		{"USD", "100.00", "% d", " 10000"},
		{"USD", "100.00", "%.6d", "10000"}, // precision is ignored
		{"USD", "100.00", "%7d", "  10000"},
		{"USD", "100.00", "%8d", "   10000"},
		{"USD", "100.00", "%9d", "    10000"},
		{"USD", "100.00", "%10d", "     10000"},
		{"USD", "100.00", "%010d", "0000010000"},
		{"USD", "100.00", "%+10d", "    +10000"},
		{"USD", "100.00", "%-10d", "10000     "},
		// %c verb
		{"USD", "100.00", "%c", "USD"},
		{"USD", "100.00", "%+c", "USD"}, // '+' is ignored
		{"USD", "100.00", "% c", "USD"}, // ' ' is ignored
		{"USD", "100.00", "%#c", "USD"}, // '#' is ignored
		{"USD", "100.00", "%5c", "  USD"},
		{"USD", "100.00", "%05c", "  USD"}, // '0' is ignored
		{"USD", "100.00", "%#5c", "  USD"}, // '#' is ignored
		{"USD", "100.00", "%-5c", "USD  "},
		{"USD", "100.00", "%-#5c", "USD  "}, // '#' is ignored
		// wrong verbs
		{"USD", "12.34", "%b", "%!b(money.Amount=USD 12.34)"},
		{"USD", "12.34", "%e", "%!e(money.Amount=USD 12.34)"},
		{"USD", "12.34", "%E", "%!E(money.Amount=USD 12.34)"},
		{"USD", "12.34", "%g", "%!g(money.Amount=USD 12.34)"},
		{"USD", "12.34", "%G", "%!G(money.Amount=USD 12.34)"},
		{"USD", "12.34", "%x", "%!x(money.Amount=USD 12.34)"},
		{"USD", "12.34", "%X", "%!X(money.Amount=USD 12.34)"},
	}
	for _, tt := range tests {
		a := MustParseAmount(tt.c, tt.a)
		got := fmt.Sprintf(tt.format, a)
		if got != tt.want {
			t.Errorf("fmt.Sprintf(%q, %q) = %q, want %q", tt.format, a, got, tt.want)
		}
	}
}

func TestAmount_Cmp(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			c, a, b string
			want    int
		}{
			{"USD", "-2", "-2", 0},
			{"USD", "-2", "-1", -1},
			{"USD", "-2", "0", -1},
			{"USD", "-2", "1", -1},
			{"USD", "-2", "2", -1},
			{"USD", "-1", "-2", 1},
			{"USD", "-1", "-1", 0},
			{"USD", "-1", "0", -1},
			{"USD", "-1", "1", -1},
			{"USD", "-1", "2", -1},
			{"USD", "0", "-2", 1},
			{"USD", "0", "-1", 1},
			{"USD", "0", "0", 0},
			{"USD", "0", "1", -1},
			{"USD", "0", "2", -1},
			{"USD", "1", "-2", 1},
			{"USD", "1", "-1", 1},
			{"USD", "1", "0", 1},
			{"USD", "1", "1", 0},
			{"USD", "1", "2", -1},
			{"USD", "2", "-2", 1},
			{"USD", "2", "-1", 1},
			{"USD", "2", "0", 1},
			{"USD", "2", "1", 1},
			{"JPY", "2", "2", 0},
			{"JPY", "2", "2.0", 0},
			{"JPY", "2", "2.00", 0},
			{"JPY", "2", "2.000", 0},
			{"JPY", "2", "2.0000", 0},
			{"JPY", "2", "2.00000", 0},
			{"JPY", "2", "2.000000", 0},
			{"JPY", "2", "2.0000000", 0},
			{"JPY", "2", "2.00000000", 0},
			{"USD", "99999999999999999.99", "0.9999999999999999999", 1},
			{"USD", "0.9999999999999999999", "99999999999999999.99", -1},
		}
		for _, tt := range tests {
			d := MustParseAmount(tt.c, tt.a)
			e := MustParseAmount(tt.c, tt.b)
			got, err := d.Cmp(e)
			if err != nil {
				t.Errorf("%q.Cmp(%q) failed: %v", d, e, err)
				continue
			}
			if got != tt.want {
				t.Errorf("%q.Cmp(%q) = %v, want %v", d, e, got, tt.want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		a := MustParseAmount("USD", "1")
		b := MustParseAmount("JPY", "1")
		_, err := a.Cmp(b)
		if err == nil {
			t.Errorf("%q.Cmp(%q) did not fail", a, b)
		}
	})
}

func TestAmount_CmpTotal(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			c, a, b string
			want    int
		}{
			{"USD", "-2", "-2", 0},
			{"USD", "-2", "-1", -1},
			{"USD", "-2", "0", -1},
			{"USD", "-2", "1", -1},
			{"USD", "-2", "2", -1},
			{"USD", "-1", "-2", 1},
			{"USD", "-1", "-1", 0},
			{"USD", "-1", "0", -1},
			{"USD", "-1", "1", -1},
			{"USD", "-1", "2", -1},
			{"USD", "0", "-2", 1},
			{"USD", "0", "-1", 1},
			{"USD", "0", "0", 0},
			{"USD", "0", "1", -1},
			{"USD", "0", "2", -1},
			{"USD", "1", "-2", 1},
			{"USD", "1", "-1", 1},
			{"USD", "1", "0", 1},
			{"USD", "1", "1", 0},
			{"USD", "1", "2", -1},
			{"USD", "2", "-2", 1},
			{"USD", "2", "-1", 1},
			{"USD", "2", "0", 1},
			{"USD", "2", "1", 1},
			{"JPY", "2", "2", 0},
			{"JPY", "2", "2.0", 1},
			{"JPY", "2", "2.00", 1},
			{"JPY", "2", "2.000", 1},
			{"JPY", "2", "2.0000", 1},
			{"JPY", "2", "2.00000", 1},
			{"JPY", "2", "2.000000", 1},
			{"JPY", "2", "2.0000000", 1},
			{"JPY", "2", "2.00000000", 1},
			{"USD", "99999999999999999.99", "0.9999999999999999999", 1},
			{"USD", "0.9999999999999999999", "99999999999999999.99", -1},
		}
		for _, tt := range tests {
			d := MustParseAmount(tt.c, tt.a)
			e := MustParseAmount(tt.c, tt.b)
			got, err := d.CmpTotal(e)
			if err != nil {
				t.Errorf("%q.CmpTotal(%q) failed: %v", d, e, err)
				continue
			}
			if got != tt.want {
				t.Errorf("%q.CmpTotal(%q) = %v, want %v", d, e, got, tt.want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		a := MustParseAmount("USD", "1")
		b := MustParseAmount("JPY", "1")
		_, err := a.CmpTotal(b)
		if err == nil {
			t.Errorf("%q.CmpTotal(%q) did not fail", a, b)
		}
	})
}
