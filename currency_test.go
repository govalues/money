package money

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"fmt"
	"testing"
)

func TestCurrency_Interfaces(t *testing.T) {
	var c any

	c = XXX
	_, ok := c.(fmt.Stringer)
	if !ok {
		t.Errorf("%T does not implement fmt.Stringer", c)
	}
	_, ok = c.(fmt.Formatter)
	if !ok {
		t.Errorf("%T does not implement fmt.Formatter", c)
	}
	_, ok = c.(encoding.TextMarshaler)
	if !ok {
		t.Errorf("%T does not implement encoding.TextMarshaler", c)
	}
	_, ok = c.(encoding.BinaryMarshaler)
	if !ok {
		t.Errorf("%T does not implement encoding.BinaryMarshaler", c)
	}
	// Uncomment when Go 1.24 is minimum supported version.
	// _, ok = d.(encoding.TextAppender)
	// if !ok {
	// 	t.Errorf("%T does not implement encoding.TextAppender", d)
	// }
	// _, ok = d.(encoding.BinaryAppender)
	// if !ok {
	// 	t.Errorf("%T does not implement encoding.BinaryAppender", d)
	// }
	_, ok = c.(driver.Valuer)
	if !ok {
		t.Errorf("%T does not implement driver.Valuer", c)
	}

	x := XXX
	c = &x
	_, ok = c.(encoding.TextUnmarshaler)
	if !ok {
		t.Errorf("%T does not implement encoding.TextUnmarshaler", c)
	}
	_, ok = c.(encoding.BinaryUnmarshaler)
	if !ok {
		t.Errorf("%T does not implement encoding.BinaryUnmarshaler", c)
	}
	_, ok = c.(sql.Scanner)
	if !ok {
		t.Errorf("%T does not implement sql.Scanner", c)
	}
}

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
		{XTS, 2},
		{AED, 2},
		{AFN, 2},
		{ALL, 2},
		{AMD, 2},
		{ANG, 2},
		{AOA, 2},
		{ARS, 2},
		{AUD, 2},
		{AWG, 2},
		{AZN, 2},
		{BAM, 2},
		{BBD, 2},
		{BDT, 2},
		{BGN, 2},
		{BHD, 3},
		{BIF, 0},
		{BMD, 2},
		{BND, 2},
		{BOB, 2},
		{BRL, 2},
		{BSD, 2},
		{BTN, 2},
		{BWP, 2},
		{BYN, 2},
		{BZD, 2},
		{CAD, 2},
		{CDF, 2},
		{CHF, 2},
		{CLP, 0},
		{CNY, 2},
		{COP, 2},
		{CRC, 2},
		{CUP, 2},
		{CVE, 2},
		{CZK, 2},
		{DJF, 0},
		{DKK, 2},
		{DOP, 2},
		{DZD, 2},
		{EGP, 2},
		{ERN, 2},
		{ETB, 2},
		{EUR, 2},
		{FJD, 2},
		{FKP, 2},
		{GBP, 2},
		{GEL, 2},
		{GHS, 2},
		{GIP, 2},
		{GMD, 2},
		{GNF, 0},
		{GTQ, 2},
		{GWP, 2},
		{GYD, 2},
		{HKD, 2},
		{HNL, 2},
		{HRK, 2},
		{HTG, 2},
		{HUF, 2},
		{IDR, 2},
		{ILS, 2},
		{INR, 2},
		{IQD, 3},
		{IRR, 2},
		{ISK, 2},
		{JMD, 2},
		{JOD, 3},
		{JPY, 0},
		{KES, 2},
		{KGS, 2},
		{KHR, 2},
		{KMF, 0},
		{KPW, 2},
		{KRW, 0},
		{KWD, 3},
		{KYD, 2},
		{KZT, 2},
		{LAK, 2},
		{LBP, 2},
		{LKR, 2},
		{LRD, 2},
		{LSL, 2},
		{LYD, 3},
		{MAD, 2},
		{MDL, 2},
		{MGA, 2},
		{MKD, 2},
		{MMK, 2},
		{MNT, 2},
		{MOP, 2},
		{MRU, 2},
		{MUR, 2},
		{MVR, 2},
		{MWK, 2},
		{MXN, 2},
		{MYR, 2},
		{MZN, 2},
		{NAD, 2},
		{NGN, 2},
		{NIO, 2},
		{NOK, 2},
		{NPR, 2},
		{NZD, 2},
		{OMR, 3},
		{PAB, 2},
		{PEN, 2},
		{PGK, 2},
		{PHP, 2},
		{PKR, 2},
		{PLN, 2},
		{PYG, 0},
		{QAR, 2},
		{RON, 2},
		{RSD, 2},
		{RUB, 2},
		{RWF, 0},
		{SAR, 2},
		{SBD, 2},
		{SCR, 2},
		{SDG, 2},
		{SEK, 2},
		{SGD, 2},
		{SHP, 2},
		{SLL, 2},
		{SOS, 2},
		{SRD, 2},
		{SSP, 2},
		{STN, 2},
		{SYP, 2},
		{SZL, 2},
		{THB, 2},
		{TJS, 2},
		{TMT, 2},
		{TND, 3},
		{TOP, 2},
		{TRY, 2},
		{TTD, 2},
		{TWD, 2},
		{TZS, 2},
		{UAH, 2},
		{UGX, 0},
		{USD, 2},
		{UYU, 2},
		{UZS, 2},
		{VES, 2},
		{VND, 0},
		{VUV, 0},
		{WST, 2},
		{XAF, 0},
		{XCD, 2},
		{XOF, 0},
		{XPF, 0},
		{YER, 2},
		{ZAR, 2},
		{ZMW, 2},
		{ZWL, 2},
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
		{XTS, "963"},
		{AED, "784"},
		{AFN, "971"},
		{ALL, "008"},
		{AMD, "051"},
		{ANG, "532"},
		{AOA, "973"},
		{ARS, "032"},
		{AUD, "036"},
		{AWG, "533"},
		{AZN, "944"},
		{BAM, "977"},
		{BBD, "052"},
		{BDT, "050"},
		{BGN, "975"},
		{BHD, "048"},
		{BIF, "108"},
		{BMD, "060"},
		{BND, "096"},
		{BOB, "068"},
		{BRL, "986"},
		{BSD, "044"},
		{BTN, "064"},
		{BWP, "072"},
		{BYN, "933"},
		{BZD, "084"},
		{CAD, "124"},
		{CDF, "976"},
		{CHF, "756"},
		{CLP, "152"},
		{CNY, "156"},
		{COP, "170"},
		{CRC, "188"},
		{CUP, "192"},
		{CVE, "132"},
		{CZK, "203"},
		{DJF, "262"},
		{DKK, "208"},
		{DOP, "214"},
		{DZD, "012"},
		{EGP, "818"},
		{ERN, "232"},
		{ETB, "230"},
		{EUR, "978"},
		{FJD, "242"},
		{FKP, "238"},
		{GBP, "826"},
		{GEL, "981"},
		{GHS, "936"},
		{GIP, "292"},
		{GMD, "270"},
		{GNF, "324"},
		{GTQ, "320"},
		{GWP, "624"},
		{GYD, "328"},
		{HKD, "344"},
		{HNL, "340"},
		{HRK, "191"},
		{HTG, "332"},
		{HUF, "348"},
		{IDR, "360"},
		{ILS, "376"},
		{INR, "356"},
		{IQD, "368"},
		{IRR, "364"},
		{ISK, "352"},
		{JMD, "388"},
		{JOD, "400"},
		{JPY, "392"},
		{KES, "404"},
		{KGS, "417"},
		{KHR, "116"},
		{KMF, "174"},
		{KPW, "408"},
		{KRW, "410"},
		{KWD, "414"},
		{KYD, "136"},
		{KZT, "398"},
		{LAK, "418"},
		{LBP, "422"},
		{LKR, "144"},
		{LRD, "430"},
		{LSL, "426"},
		{LYD, "434"},
		{MAD, "504"},
		{MDL, "498"},
		{MGA, "969"},
		{MKD, "807"},
		{MMK, "104"},
		{MNT, "496"},
		{MOP, "446"},
		{MRU, "929"},
		{MUR, "480"},
		{MVR, "462"},
		{MWK, "454"},
		{MXN, "484"},
		{MYR, "458"},
		{MZN, "943"},
		{NAD, "516"},
		{NGN, "566"},
		{NIO, "558"},
		{NOK, "578"},
		{NPR, "524"},
		{NZD, "554"},
		{OMR, "512"},
		{PAB, "590"},
		{PEN, "604"},
		{PGK, "598"},
		{PHP, "608"},
		{PKR, "586"},
		{PLN, "985"},
		{PYG, "600"},
		{QAR, "634"},
		{RON, "946"},
		{RSD, "941"},
		{RUB, "643"},
		{RWF, "646"},
		{SAR, "682"},
		{SBD, "090"},
		{SCR, "690"},
		{SDG, "938"},
		{SEK, "752"},
		{SGD, "702"},
		{SHP, "654"},
		{SLL, "694"},
		{SOS, "706"},
		{SRD, "968"},
		{SSP, "728"},
		{STN, "930"},
		{SYP, "760"},
		{SZL, "748"},
		{THB, "764"},
		{TJS, "972"},
		{TMT, "934"},
		{TND, "788"},
		{TOP, "776"},
		{TRY, "949"},
		{TTD, "780"},
		{TWD, "901"},
		{TZS, "834"},
		{UAH, "980"},
		{UGX, "800"},
		{USD, "840"},
		{UYU, "858"},
		{UZS, "860"},
		{VES, "928"},
		{VND, "704"},
		{VUV, "548"},
		{WST, "882"},
		{XAF, "950"},
		{XCD, "951"},
		{XOF, "952"},
		{XPF, "953"},
		{YER, "886"},
		{ZAR, "710"},
		{ZMW, "967"},
		{ZWL, "932"},
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
		{XTS, "XTS"},
		{AED, "AED"},
		{AFN, "AFN"},
		{ALL, "ALL"},
		{AMD, "AMD"},
		{ANG, "ANG"},
		{AOA, "AOA"},
		{ARS, "ARS"},
		{AUD, "AUD"},
		{AWG, "AWG"},
		{AZN, "AZN"},
		{BAM, "BAM"},
		{BBD, "BBD"},
		{BDT, "BDT"},
		{BGN, "BGN"},
		{BHD, "BHD"},
		{BIF, "BIF"},
		{BMD, "BMD"},
		{BND, "BND"},
		{BOB, "BOB"},
		{BRL, "BRL"},
		{BSD, "BSD"},
		{BTN, "BTN"},
		{BWP, "BWP"},
		{BYN, "BYN"},
		{BZD, "BZD"},
		{CAD, "CAD"},
		{CDF, "CDF"},
		{CHF, "CHF"},
		{CLP, "CLP"},
		{CNY, "CNY"},
		{COP, "COP"},
		{CRC, "CRC"},
		{CUP, "CUP"},
		{CVE, "CVE"},
		{CZK, "CZK"},
		{DJF, "DJF"},
		{DKK, "DKK"},
		{DOP, "DOP"},
		{DZD, "DZD"},
		{EGP, "EGP"},
		{ERN, "ERN"},
		{ETB, "ETB"},
		{EUR, "EUR"},
		{FJD, "FJD"},
		{FKP, "FKP"},
		{GBP, "GBP"},
		{GEL, "GEL"},
		{GHS, "GHS"},
		{GIP, "GIP"},
		{GMD, "GMD"},
		{GNF, "GNF"},
		{GTQ, "GTQ"},
		{GWP, "GWP"},
		{GYD, "GYD"},
		{HKD, "HKD"},
		{HNL, "HNL"},
		{HRK, "HRK"},
		{HTG, "HTG"},
		{HUF, "HUF"},
		{IDR, "IDR"},
		{ILS, "ILS"},
		{INR, "INR"},
		{IQD, "IQD"},
		{IRR, "IRR"},
		{ISK, "ISK"},
		{JMD, "JMD"},
		{JOD, "JOD"},
		{JPY, "JPY"},
		{KES, "KES"},
		{KGS, "KGS"},
		{KHR, "KHR"},
		{KMF, "KMF"},
		{KPW, "KPW"},
		{KRW, "KRW"},
		{KWD, "KWD"},
		{KYD, "KYD"},
		{KZT, "KZT"},
		{LAK, "LAK"},
		{LBP, "LBP"},
		{LKR, "LKR"},
		{LRD, "LRD"},
		{LSL, "LSL"},
		{LYD, "LYD"},
		{MAD, "MAD"},
		{MDL, "MDL"},
		{MGA, "MGA"},
		{MKD, "MKD"},
		{MMK, "MMK"},
		{MNT, "MNT"},
		{MOP, "MOP"},
		{MRU, "MRU"},
		{MUR, "MUR"},
		{MVR, "MVR"},
		{MWK, "MWK"},
		{MXN, "MXN"},
		{MYR, "MYR"},
		{MZN, "MZN"},
		{NAD, "NAD"},
		{NGN, "NGN"},
		{NIO, "NIO"},
		{NOK, "NOK"},
		{NPR, "NPR"},
		{NZD, "NZD"},
		{OMR, "OMR"},
		{PAB, "PAB"},
		{PEN, "PEN"},
		{PGK, "PGK"},
		{PHP, "PHP"},
		{PKR, "PKR"},
		{PLN, "PLN"},
		{PYG, "PYG"},
		{QAR, "QAR"},
		{RON, "RON"},
		{RSD, "RSD"},
		{RUB, "RUB"},
		{RWF, "RWF"},
		{SAR, "SAR"},
		{SBD, "SBD"},
		{SCR, "SCR"},
		{SDG, "SDG"},
		{SEK, "SEK"},
		{SGD, "SGD"},
		{SHP, "SHP"},
		{SLL, "SLL"},
		{SOS, "SOS"},
		{SRD, "SRD"},
		{SSP, "SSP"},
		{STN, "STN"},
		{SYP, "SYP"},
		{SZL, "SZL"},
		{THB, "THB"},
		{TJS, "TJS"},
		{TMT, "TMT"},
		{TND, "TND"},
		{TOP, "TOP"},
		{TRY, "TRY"},
		{TTD, "TTD"},
		{TWD, "TWD"},
		{TZS, "TZS"},
		{UAH, "UAH"},
		{UGX, "UGX"},
		{USD, "USD"},
		{UYU, "UYU"},
		{UZS, "UZS"},
		{VES, "VES"},
		{VND, "VND"},
		{VUV, "VUV"},
		{WST, "WST"},
		{XAF, "XAF"},
		{XCD, "XCD"},
		{XOF, "XOF"},
		{XPF, "XPF"},
		{YER, "YER"},
		{ZAR, "ZAR"},
		{ZMW, "ZMW"},
		{ZWL, "ZWL"},
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
		tests := []any{"UUU", 840, []byte{0x08, 0x40}, nil}
		for _, tt := range tests {
			var got Currency
			err := got.Scan(tt)
			if err == nil {
				t.Errorf("Scan(%q) did not fail", tt)
			}
		}
	})
}

func TestNullCurrency_Interfaces(t *testing.T) {
	var i any = NullCurrency{}
	_, ok := i.(driver.Valuer)
	if !ok {
		t.Errorf("%T does not implement driver.Valuer", i)
	}

	i = &NullCurrency{}
	_, ok = i.(sql.Scanner)
	if !ok {
		t.Errorf("%T does not implement sql.Scanner", i)
	}
}

func TestNullCurrency_Scan(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		tests := []any{"UUU", 840, []byte{0x08, 0x40}}
		for _, tt := range tests {
			var got NullCurrency
			err := got.Scan(tt)
			if err == nil {
				t.Errorf("Scan(%q) did not fail", tt)
			}
		}
	})
}
