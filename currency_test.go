package money

import (
	"fmt"
	"testing"
)

func TestCurrency_Parse(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			code string
			want Currency
		}{
			{"999", XXX},
			{"xxx", XXX},
			{"XXX", XXX},
			{"392", JPY},
			{"jpy", JPY},
			{"JPY", JPY},
			{"840", USD},
			{"usd", USD},
			{"USD", USD},
			{"512", OMR},
			{"omr", OMR},
			{"OMR", OMR},
		}
		for _, tt := range tests {
			got, err := ParseCurr(tt.code)
			if err != nil {
				t.Errorf("ParseCurr(%q) failed: %v", tt.code, err)
				continue
			}
			if got != tt.want {
				t.Errorf("ParseCurr(%q) = %v, want %v", tt.code, got, tt.want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := []string{
			"", "000", "test", "xbt", "$", "AU$", "BTC",
		}
		for _, tt := range tests {
			_, err := ParseCurr(tt)
			if err == nil {
				t.Errorf("ParseCurr(%q) did not fail", tt)
			}
		}
	})
}

func TestMustParseCurr(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("MustParseCurr(\"UUU\") did not panic")
			}
		}()
		MustParseCurr("UUU")
	})
}

func TestCurrency_Scale(t *testing.T) {
	tests := []struct {
		curr Currency
		want int
	}{
		{XXX, 0},
		{JPY, 0},
		{AED, 2},
		{EUR, 2},
		{USD, 2},
		{OMR, 3},
		{IQD, 3},
	}
	for _, tt := range tests {
		got := tt.curr.Scale()
		if got != tt.want {
			t.Errorf("%v.Scale() = %v, want %v", tt.curr, got, tt.want)
		}
	}
}

func TestCurrency_Num(t *testing.T) {
	tests := []struct {
		curr Currency
		want string
	}{
		{XXX, "999"},
		{JPY, "392"},
		{USD, "840"},
		{OMR, "512"},
	}
	for _, tt := range tests {
		got := tt.curr.Num()
		if got != tt.want {
			t.Errorf("%v.Num() = %v, want %v", tt.curr, got, tt.want)
		}
	}
}

func TestCurrency_Code(t *testing.T) {
	tests := []struct {
		curr Currency
		want string
	}{
		{XXX, "XXX"},
		{JPY, "JPY"},
		{USD, "USD"},
		{OMR, "OMR"},
	}
	for _, tt := range tests {
		got := tt.curr.Code()
		if got != tt.want {
			t.Errorf("%v.Code() = %v, want %v", tt.curr, got, tt.want)
		}
	}
}

func TestCurrency_Format(t *testing.T) {
	tests := []struct {
		curr         Currency
		format, want string
	}{
		// %T verb
		{USD, "%T", "money.Currency"},
		// %q verb
		{USD, "%q", "\"USD\""},
		{USD, "%6q", " \"USD\""},
		{USD, "%7q", "  \"USD\""},
		{USD, "%07q", "  \"USD\""}, // '0' is ignored
		{USD, "%+7q", "  \"USD\""}, // '+' is ignored
		{USD, "%-7q", "\"USD\"  "},
		// %s verb
		{JPY, "%s", "JPY"},
		{JPY, "%4s", " JPY"},
		{JPY, "%5s", "  JPY"},
		{JPY, "%05s", "  JPY"}, // '0' is ignored
		{JPY, "%+5s", "  JPY"}, // '+' is ignored
		{JPY, "%-5s", "JPY  "},
		// %v verb
		{OMR, "%v", "OMR"},
		{OMR, "%4v", " OMR"},
		{OMR, "%5v", "  OMR"},
		{OMR, "%05v", "  OMR"}, // '0' is ignored
		{OMR, "%+5v", "  OMR"}, // '+' is ignored
		{OMR, "%-5v", "OMR  "},
		// %c verb
		{XXX, "%c", "XXX"},
		{JPY, "%c", "JPY"},
		{OMR, "%c", "OMR"},
		{USD, "%c", "USD"},
		{USD, "%+c", "USD"}, // '+' is ignored
		{USD, "% c", "USD"}, // ' ' is ignored
		{USD, "%#c", "USD"}, // '#' is ignored
		{USD, "%5c", "  USD"},
		{USD, "%05c", "  USD"}, // '0' is ignored
		{USD, "%#5c", "  USD"}, // '#' is ignored
		{USD, "%-5c", "USD  "},
		{USD, "%-#5c", "USD  "}, // '#' is ignored
		// wrong verbs
		{USD, "%b", "%!b(money.Currency=USD)"},
	}
	for _, tt := range tests {
		got := fmt.Sprintf(tt.format, tt.curr)
		if got != tt.want {
			t.Errorf("fmt.Sprintf(%q, %v) = %q, want %q", tt.format, tt.curr, got, tt.want)
		}
	}
}

func TestCurrency_Scan(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		c := XXX
		err := c.Scan([]byte("USD"))
		if err == nil {
			t.Errorf("c.Scan([]byte(\"USD\")) did not fail")
		}
	})
}

func TestNullCurrency_Scan(t *testing.T) {
	t.Run("[]byte", func(t *testing.T) {
		tests := []string{"UUU"}
		for _, tt := range tests {
			got := NullCurrency{}
			err := got.Scan([]byte(tt))
			if err == nil {
				t.Errorf("Scan(%q) did not fail", tt)
			}
		}
	})
}
