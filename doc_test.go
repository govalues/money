package money_test

import (
	"fmt"
	"github.com/govalues/decimal"
	"github.com/govalues/money"
	"strconv"
)

func ExampleNewAmount() {
	d := decimal.New(100, 0)
	a, err := money.NewAmount(money.USD, d)
	if err != nil {
		panic(err)
	}
	fmt.Println(a)
	// Output: USD 100.00
}

// In this example, the sales tax amount is calculated for a product with
// a given price after tax, using a specified tax rate.
func Example_taxCalculation() {
	one := decimal.New(1, 0)
	priceAfterTax := money.MustParseAmount("USD", "10")
	taxRate := decimal.MustParse("0.065")
	priceBeforeTax := priceAfterTax.Quo(one.Add(taxRate)).RoundToCurr()
	salesTax := priceAfterTax.Sub(priceBeforeTax)
	fmt.Printf("Price (before tax) = %v\n", priceBeforeTax)
	fmt.Printf("Tax Rate           = %k\n", taxRate)
	fmt.Printf("Sales Tax          = %v\n", salesTax)
	fmt.Printf("Price (after tax)  = %v\n", priceAfterTax)
	// Output:
	// Price (before tax) = USD 9.39
	// Tax Rate           = 6.5%
	// Sales Tax          = USD 0.61
	// Price (after tax)  = USD 10.00
}

type StatementLine struct {
	Month    int
	Days     int
	Incoming money.Amount
	Interest money.Amount
	Outgoing money.Amount
}

type Statement []StatementLine

func NewStatement(month, days int, balance money.Amount) Statement {
	return Statement{{month, days, balance, balance.Zero(), balance}}
}

func (s Statement) IncomingBalance() money.Amount {
	return s[0].Incoming
}

func (s Statement) OutgoingBalance() money.Amount {
	return s[len(s)-1].Outgoing
}

func (s Statement) PercChange() decimal.Decimal {
	return s.OutgoingBalance().Sub(s.IncomingBalance()).Rat(s.IncomingBalance())
}

func (s Statement) Append(month, days int, interest money.Amount) Statement {
	line := StatementLine{
		Month:    month,
		Days:     days,
		Incoming: s.OutgoingBalance(),
		Interest: interest,
		Outgoing: s.OutgoingBalance().Add(interest),
	}
	if len(s) == 1 && s[0].Interest.IsZero() {
		return Statement{line}
	}
	return append(s, line)
}

func (s Statement) TotalInterest() money.Amount {
	if len(s) == 0 {
		return money.Amount{}
	}
	total := s[0].Interest
	for i := 1; i < len(s); i++ {
		total = total.Add(s[i].Interest)
	}
	return total
}

func DailyRate(yearlyRate decimal.Decimal) decimal.Decimal {
	daysInYear := decimal.New(365, 0)
	return yearlyRate.Quo(daysInYear)
}

func SimulateStatement(balance money.Amount, yearlyRate decimal.Decimal) Statement {
	daysInMonths := [...]int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	dailyRate := DailyRate(yearlyRate)
	statement := NewStatement(0, 0, balance)
	for m := 0; m < len(daysInMonths); m++ {
		interestAccumulated := balance.Zero()
		for i := 0; i < daysInMonths[m]; i++ {
			interestAccumulated = balance.FMA(dailyRate, interestAccumulated)
		}
		interestAccumulated = interestAccumulated.RoundToCurr()
		statement = statement.Append(m+1, daysInMonths[m], interestAccumulated)
		balance = balance.Add(interestAccumulated)
	}
	return statement
}

// This example calculates the effective interest rate for a 10% nominal
// interest rate compounded monthly on a USD 10,000 balance.
func Example_effectiveRate() {

	// Parameters
	balance := money.MustParseAmount("USD", "10000")
	nominalRate := decimal.MustParse("0.10")

	// Calculate statement
	statement := SimulateStatement(balance, nominalRate)

	// Print header
	fmt.Printf("Nominal Rate = %.2k, Effective Rate = %.4k\n\n", nominalRate, statement.PercChange())
	fmt.Printf("Month Days  Incoming    Interest     Outgoing\n")

	// Print statement lines
	for _, line := range statement {
		fmt.Printf(
			"%5v %5v %11f %+11f %11f\n",
			line.Month, line.Days,
			line.Incoming, line.Interest,
			line.Outgoing)
	}

	// Print total
	fmt.Printf("                  Total %+11f\n",
		statement.TotalInterest())

	// Output:
	// Nominal Rate = 10.00%, Effective Rate = 10.4713%
	//
	// Month Days  Incoming    Interest     Outgoing
	//     1    31    10000.00      +84.93    10084.93
	//     2    28    10084.93      +77.36    10162.29
	//     3    31    10162.29      +86.31    10248.60
	//     4    30    10248.60      +84.24    10332.84
	//     5    31    10332.84      +87.76    10420.60
	//     6    30    10420.60      +85.65    10506.25
	//     7    31    10506.25      +89.23    10595.48
	//     8    31    10595.48      +89.99    10685.47
	//     9    30    10685.47      +87.83    10773.30
	//    10    31    10773.30      +91.50    10864.80
	//    11    30    10864.80      +89.30    10954.10
	//    12    31    10954.10      +93.03    11047.13
	//                   Total    +1047.13
}

type ScheduleLine struct {
	Month       int
	Repayment   money.Amount
	Principal   money.Amount
	Interest    money.Amount
	Outstanding money.Amount
}

type AmortizationSchedule []ScheduleLine

func (p AmortizationSchedule) Append(month int, repayment, principal, interest, outstanding money.Amount) AmortizationSchedule {
	newLine := ScheduleLine{
		Month:       month,
		Repayment:   repayment,
		Principal:   principal,
		Interest:    interest,
		Outstanding: outstanding,
	}
	return append(p, newLine)
}

func (p AmortizationSchedule) TotalRepayment() money.Amount {
	if len(p) == 0 {
		return money.Amount{}
	}
	total := p[0].Repayment
	for i := 1; i < len(p); i++ {
		total = total.Add(p[i].Repayment)
	}
	return total
}

func (p AmortizationSchedule) TotalPrincipal() money.Amount {
	if len(p) == 0 {
		return money.Amount{}
	}
	total := p[0].Principal
	for i := 1; i < len(p); i++ {
		total = total.Add(p[i].Principal)
	}
	return total
}

func (p AmortizationSchedule) TotalInterest() money.Amount {
	if len(p) == 0 {
		return money.Amount{}
	}
	total := p[0].Interest
	for i := 1; i < len(p); i++ {
		total = total.Add(p[i].Interest)
	}
	return total
}

func MonthlyRate(yearlyRate decimal.Decimal) decimal.Decimal {
	monthsInYear := decimal.New(12, 0)
	return yearlyRate.Quo(monthsInYear).Round(9)
}

func AnnuityPayment(amount money.Amount, rate decimal.Decimal, periods int) money.Amount {
	one := rate.One()
	num := amount.Mul(rate)
	den := one.Sub(one.Add(rate).Pow(-periods))
	return num.Quo(den)
}

func SimulateSchedule(principal money.Amount, yearlyRate decimal.Decimal, years int) AmortizationSchedule {
	months := years * 12
	monthlyRate := MonthlyRate(yearlyRate)
	totalAmount := AnnuityPayment(principal, monthlyRate, months)
	outgoingBalance := principal
	schedule := AmortizationSchedule{}
	for i := 0; i < months; i++ {
		interestPart := outgoingBalance.Mul(monthlyRate).RoundToCurr()
		principalPart := totalAmount.Sub(interestPart)
		outgoingBalance = outgoingBalance.Sub(principalPart)
		schedule = schedule.Append(i+1, totalAmount, principalPart, interestPart, outgoingBalance)
	}
	return schedule
}

// In this example, a loan amortization table is generated for a loan with
// an initial amount of USD 12,000, an annual interest rate of 10%, and
// a repayment period of 1 year.
func Example_loanAmortization() {
	// Loan parameters
	initialAmount := money.MustParseAmount("USD", "12000")
	yearlyInterestRate := decimal.MustParse("0.1")
	years := 1

	// Calculate loan repayment plan
	schedule := SimulateSchedule(initialAmount, yearlyInterestRate, years)

	// Print header
	fmt.Printf("Initial Amount = %v, Interest Rate = %.2k\n\n", initialAmount, yearlyInterestRate)
	fmt.Println("Month  Repayment   Principal   Interest    Outstanding")

	// Print plan lines
	for _, line := range schedule {
		fmt.Printf("%5d %12f %11f %11f %11f\n",
			line.Month, line.Repayment, line.Principal,
			line.Interest, line.Outstanding)
	}

	// Print total
	fmt.Printf("Total %12f %11f %11f\n",
		schedule.TotalRepayment(),
		schedule.TotalPrincipal(),
		schedule.TotalInterest())

	// Output:
	// Initial Amount = USD 12000.00, Interest Rate = 10.00%
	//
	// Month  Repayment   Principal   Interest    Outstanding
	//     1      1054.99      954.99      100.00    11045.01
	//     2      1054.99      962.95       92.04    10082.06
	//     3      1054.99      970.97       84.02     9111.09
	//     4      1054.99      979.06       75.93     8132.03
	//     5      1054.99      987.22       67.77     7144.81
	//     6      1054.99      995.45       59.54     6149.36
	//     7      1054.99     1003.75       51.24     5145.61
	//     8      1054.99     1012.11       42.88     4133.49
	//     9      1054.99     1020.54       34.45     3112.95
	//    10      1054.99     1029.05       25.94     2083.90
	//    11      1054.99     1037.62       17.37     1046.28
	//    12      1054.99     1046.27        8.72        0.01
	// Total     12659.89    11999.99      659.90
}

func ExampleMustParseAmount() {
	a := money.MustParseAmount("USD", "-1.230")
	fmt.Println(a)
	// Output: USD -1.230
}

func ExampleAmount_Coef() {
	a := money.MustParseAmount("JPY", "-123")
	b := money.MustParseAmount("JPY", "5.7")
	c := money.MustParseAmount("JPY", "0.4")
	fmt.Println(a.Coef())
	fmt.Println(b.Coef())
	fmt.Println(c.Coef())
	// Output:
	// 123
	// 57
	// 4
}

func ExampleAmount_MinorUnits() {
	a := money.MustParseAmount("JPY", "-1.6789")
	b := money.MustParseAmount("USD", "-1.6789")
	c := money.MustParseAmount("OMR", "-1.6789")
	fmt.Println(a.MinorUnits())
	fmt.Println(b.MinorUnits())
	fmt.Println(c.MinorUnits())
	// Output:
	// -2 true
	// -168 true
	// -1679 true
}

func ExampleAmount_Prec() {
	a := money.MustParseAmount("JPY", "-123")
	b := money.MustParseAmount("JPY", "5.7")
	c := money.MustParseAmount("JPY", "0.4")
	fmt.Println(a.Prec())
	fmt.Println(b.Prec())
	fmt.Println(c.Prec())
	// Output:
	// 3
	// 2
	// 1
}

func ExampleAmount_Curr() {
	a := money.MustParseAmount("USD", "15.6")
	fmt.Println(a.Curr())
	// Output: USD
}

func ExampleAmount_Add() {
	a := money.MustParseAmount("USD", "15.6")
	b := money.MustParseAmount("USD", "8")
	fmt.Println(a.Add(b))
	// Output: USD 23.60
}

func ExampleAmount_Sub() {
	a := money.MustParseAmount("USD", "15.6")
	b := money.MustParseAmount("USD", "8")
	fmt.Println(a.Sub(b))
	// Output: USD 7.60
}

func ExampleAmount_FMA() {
	a := money.MustParseAmount("USD", "2")
	e := decimal.MustParse("3")
	b := money.MustParseAmount("USD", "4")
	fmt.Println(a.FMA(e, b))
	// Output: USD 10.00
}

func ExampleAmount_Mul() {
	a := money.MustParseAmount("USD", "5.7")
	e := decimal.MustParse("3")
	fmt.Println(a.Mul(e))
	// Output: USD 17.10
}

func ExampleAmount_Quo() {
	a := money.MustParseAmount("USD", "-15.67")
	e := decimal.MustParse("2")
	fmt.Println(a.Quo(e))
	// Output: USD -7.835
}

func ExampleAmount_Rat() {
	a := money.MustParseAmount("USD", "8")
	b := money.MustParseAmount("USD", "10")
	fmt.Println(a.Rat(b))
	// Output: 0.8
}

func ExampleAmount_Round() {
	a := money.MustParseAmount("USD", "15.679")
	fmt.Println(a.Round(6))
	fmt.Println(a.Round(5))
	fmt.Println(a.Round(4))
	fmt.Println(a.Round(3))
	fmt.Println(a.Round(2))
	fmt.Println(a.Round(1))
	fmt.Println(a.Round(0))
	// Output:
	// USD 15.679000
	// USD 15.67900
	// USD 15.6790
	// USD 15.679
	// USD 15.68
	// USD 15.68
	// USD 15.68
}

func ExampleAmount_RoundToCurr() {
	a := money.MustParseAmount("JPY", "1.5678")
	b := money.MustParseAmount("USD", "1.5678")
	c := money.MustParseAmount("OMR", "1.5678")
	fmt.Println(a.RoundToCurr())
	fmt.Println(b.RoundToCurr())
	fmt.Println(c.RoundToCurr())
	// Output:
	// JPY 2
	// USD 1.57
	// OMR 1.568
}

func ExampleAmount_Quantize() {
	a := money.MustParseAmount("JPY", "15.679")
	x := money.MustParseAmount("JPY", "0.01")
	y := money.MustParseAmount("JPY", "0.1")
	z := money.MustParseAmount("JPY", "1")
	fmt.Println(a.Quantize(x))
	fmt.Println(a.Quantize(y))
	fmt.Println(a.Quantize(z))
	// Output:
	// JPY 15.68
	// JPY 15.7
	// JPY 16
}

func ExampleAmount_Ceil() {
	a := money.MustParseAmount("USD", "15.679")
	fmt.Println(a.Ceil(6))
	fmt.Println(a.Ceil(5))
	fmt.Println(a.Ceil(4))
	fmt.Println(a.Ceil(3))
	fmt.Println(a.Ceil(2))
	fmt.Println(a.Ceil(1))
	fmt.Println(a.Ceil(0))
	// Output:
	// USD 15.679000
	// USD 15.67900
	// USD 15.6790
	// USD 15.679
	// USD 15.68
	// USD 15.68
	// USD 15.68
}

func ExampleAmount_CeilToCurr() {
	a := money.MustParseAmount("JPY", "1.5678")
	b := money.MustParseAmount("USD", "1.5678")
	c := money.MustParseAmount("OMR", "1.5678")
	fmt.Println(a.CeilToCurr())
	fmt.Println(b.CeilToCurr())
	fmt.Println(c.CeilToCurr())
	// Output:
	// JPY 2
	// USD 1.57
	// OMR 1.568
}

func ExampleAmount_Floor() {
	a := money.MustParseAmount("USD", "15.679")
	fmt.Println(a.Floor(6))
	fmt.Println(a.Floor(5))
	fmt.Println(a.Floor(4))
	fmt.Println(a.Floor(3))
	fmt.Println(a.Floor(2))
	fmt.Println(a.Floor(1))
	fmt.Println(a.Floor(0))
	// Output:
	// USD 15.679000
	// USD 15.67900
	// USD 15.6790
	// USD 15.679
	// USD 15.67
	// USD 15.67
	// USD 15.67
}

func ExampleAmount_FloorToCurr() {
	a := money.MustParseAmount("JPY", "1.5678")
	b := money.MustParseAmount("USD", "1.5678")
	c := money.MustParseAmount("OMR", "1.5678")
	fmt.Println(a.FloorToCurr())
	fmt.Println(b.FloorToCurr())
	fmt.Println(c.FloorToCurr())
	// Output:
	// JPY 1
	// USD 1.56
	// OMR 1.567
}

func ExampleAmount_Trunc() {
	a := money.MustParseAmount("USD", "15.679")
	fmt.Println(a.Trunc(6))
	fmt.Println(a.Trunc(5))
	fmt.Println(a.Trunc(4))
	fmt.Println(a.Trunc(3))
	fmt.Println(a.Trunc(2))
	fmt.Println(a.Trunc(1))
	fmt.Println(a.Trunc(0))
	// Output:
	// USD 15.679000
	// USD 15.67900
	// USD 15.6790
	// USD 15.679
	// USD 15.67
	// USD 15.67
	// USD 15.67
}

func ExampleAmount_TruncToCurr() {
	a := money.MustParseAmount("JPY", "1.5678")
	b := money.MustParseAmount("USD", "1.5678")
	c := money.MustParseAmount("OMR", "1.5678")
	fmt.Println(a.TruncToCurr())
	fmt.Println(b.TruncToCurr())
	fmt.Println(c.TruncToCurr())
	// Output:
	// JPY 1
	// USD 1.56
	// OMR 1.567
}

func ExampleAmount_Reduce() {
	a := money.MustParseAmount("JPY", "10.0000")
	b := money.MustParseAmount("USD", "20.0000")
	c := money.MustParseAmount("OMR", "30.0000")
	fmt.Println(a.Reduce())
	fmt.Println(b.Reduce())
	fmt.Println(c.Reduce())
	// Output:
	// JPY 10
	// USD 20.00
	// OMR 30.000
}

func ExampleAmount_SameCurr() {
	a := money.MustParseAmount("JPY", "23.0000")
	b := money.MustParseAmount("USD", "-15.670")
	c := money.MustParseAmount("USD", "1.2340")
	fmt.Println(a.SameCurr(b))
	fmt.Println(b.SameCurr(c))
	// Output:
	// false
	// true
}

func ExampleAmount_SameScale() {
	a := money.MustParseAmount("USD", "23.0000")
	b := money.MustParseAmount("USD", "-15.670")
	c := money.MustParseAmount("USD", "1.2340")
	fmt.Println(a.SameScale(b))
	fmt.Println(a.SameScale(c))
	// Output:
	// false
	// true
}

func ExampleAmount_SameScaleAsCurr() {
	a := money.MustParseAmount("USD", "23.00")
	b := money.MustParseAmount("OMR", "-15.670")
	c := money.MustParseAmount("USD", "1.2340")
	fmt.Println(a.SameScaleAsCurr())
	fmt.Println(b.SameScaleAsCurr())
	fmt.Println(c.SameScaleAsCurr())
	// Output:
	// true
	// true
	// false
}

func ExampleAmount_Scale() {
	a := money.MustParseAmount("USD", "23.0000")
	b := money.MustParseAmount("USD", "-15.670")
	fmt.Println(a.Scale())
	fmt.Println(b.Scale())
	// Output:
	// 4
	// 3
}

func ExampleAmount_Split() {
	a := money.MustParseAmount("USD", "1.01")
	fmt.Println(a.Split(6))
	fmt.Println(a.Split(5))
	fmt.Println(a.Split(4))
	fmt.Println(a.Split(3))
	fmt.Println(a.Split(2))
	fmt.Println(a.Split(1))
	// Output:
	// [USD 0.17 USD 0.17 USD 0.17 USD 0.17 USD 0.17 USD 0.16]
	// [USD 0.21 USD 0.20 USD 0.20 USD 0.20 USD 0.20]
	// [USD 0.26 USD 0.25 USD 0.25 USD 0.25]
	// [USD 0.34 USD 0.34 USD 0.33]
	// [USD 0.51 USD 0.50]
	// [USD 1.01]
}

func ExampleAmount_Format() {
	a := money.MustParseAmount("USD", "-123.456")
	fmt.Printf("%v\n", a)
	fmt.Printf("%f\n", a)
	fmt.Printf("%d\n", a)
	fmt.Printf("%c\n", a)
	// Output:
	// USD -123.456
	// -123.46
	// -12346
	// USD
}

func ExampleAmount_String() {
	a := money.MustParseAmount("USD", "-1234567890.123456789")
	fmt.Println(a.String())
	// Output: USD -1234567890.123456789
}

func ExampleAmount_Abs() {
	a := money.MustParseAmount("USD", "-15.67")
	fmt.Println(a.Abs())
	// Output: USD 15.67
}

func ExampleAmount_Neg() {
	a := money.MustParseAmount("USD", "15.67")
	fmt.Println(a.Neg())
	// Output: USD -15.67
}

func ExampleAmount_CopySign() {
	a := money.MustParseAmount("USD", "23.00")
	b := money.MustParseAmount("USD", "-15.67")
	fmt.Println(a.CopySign(b))
	fmt.Println(b.CopySign(a))
	// Output:
	// USD -23.00
	// USD 15.67
}

func ExampleAmount_Sign() {
	a := money.MustParseAmount("USD", "-15.67")
	b := money.MustParseAmount("USD", "23")
	c := money.MustParseAmount("USD", "0")
	fmt.Println(a.Sign())
	fmt.Println(b.Sign())
	fmt.Println(c.Sign())
	// Output:
	// -1
	// 1
	// 0
}

func ExampleAmount_IsNeg() {
	a := money.MustParseAmount("USD", "-15.67")
	b := money.MustParseAmount("USD", "23")
	c := money.MustParseAmount("USD", "0")
	fmt.Println(a.IsNeg())
	fmt.Println(b.IsNeg())
	fmt.Println(c.IsNeg())
	// Output:
	// true
	// false
	// false
}

func ExampleAmount_IsZero() {
	a := money.MustParseAmount("USD", "-15.67")
	b := money.MustParseAmount("USD", "23")
	c := money.MustParseAmount("USD", "0")
	fmt.Println(a.IsZero())
	fmt.Println(b.IsZero())
	fmt.Println(c.IsZero())
	// Output:
	// false
	// false
	// true
}

func ExampleAmount_IsOne() {
	a := money.MustParseAmount("USD", "1")
	b := money.MustParseAmount("USD", "2")
	fmt.Println(a.IsOne())
	fmt.Println(b.IsOne())
	// Output:
	// true
	// false
}

func ExampleAmount_WithinOne() {
	a := money.MustParseAmount("USD", "1")
	b := money.MustParseAmount("USD", "0.9")
	c := money.MustParseAmount("USD", "-1")
	fmt.Println(a.WithinOne())
	fmt.Println(b.WithinOne())
	fmt.Println(c.WithinOne())
	// Output:
	// false
	// true
	// false
}

func ExampleAmount_IsInt() {
	a := money.MustParseAmount("USD", "1.00")
	b := money.MustParseAmount("USD", "1.01")
	fmt.Println(a.IsInt())
	fmt.Println(b.IsInt())
	// Output:
	// true
	// false
}

func ExampleAmount_IsPos() {
	a := money.MustParseAmount("USD", "-15.67")
	b := money.MustParseAmount("USD", "23")
	c := money.MustParseAmount("USD", "0")
	fmt.Println(a.IsPos())
	fmt.Println(b.IsPos())
	fmt.Println(c.IsPos())
	// Output:
	// false
	// true
	// false
}

func ExampleParseCurr() {
	c, err := money.ParseCurr("USD")
	if err != nil {
		panic(err)
	}
	fmt.Println(c)
	// Output: USD
}

func ExampleMustParseCurr() {
	c := money.MustParseCurr("USD")
	fmt.Println(c)
	// Output: USD
}

func ExampleCurrency_String() {
	c := money.USD
	fmt.Println(c.String())
	// Output: USD
}

func ExampleCurrency_Code() {
	j := money.JPY
	u := money.USD
	o := money.OMR
	fmt.Println(j.Code())
	fmt.Println(u.Code())
	fmt.Println(o.Code())
	// Output:
	// JPY
	// USD
	// OMR
}

func ExampleCurrency_Num() {
	j := money.JPY
	u := money.USD
	o := money.OMR
	fmt.Println(j.Num())
	fmt.Println(u.Num())
	fmt.Println(o.Num())
	// Output:
	// 392
	// 840
	// 512
}

func ExampleCurrency_Scale() {
	j := money.JPY
	u := money.USD
	o := money.OMR
	fmt.Println(j.Scale())
	fmt.Println(u.Scale())
	fmt.Println(o.Scale())
	// Output:
	// 0
	// 2
	// 3
}

func ExampleCurrency_MarshalText() {
	c := money.MustParseCurr("USD")
	b, err := c.MarshalText()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
	// Output: USD
}

func ExampleCurrency_UnmarshalText() {
	c := money.XXX
	b := []byte("USD")
	err := c.UnmarshalText(b)
	if err != nil {
		panic(err)
	}
	fmt.Println(c)
	// Output: USD
}

func ExampleCurrency_Format() {
	fmt.Printf("%c\n", money.USD)
	// Output:
	// USD
}

func ParseISO8583(s string) (money.Amount, error) {
	c, err := money.ParseCurr(s[:3])
	if err != nil {
		return money.Amount{}, err
	}
	i, err := strconv.ParseInt(s[4:], 10, 64)
	if err != nil {
		return money.Amount{}, err
	}
	if s[3:4] == "D" {
		i = -i
	}
	d := decimal.New(i, c.Scale())
	return money.NewAmount(c, d)
}

// In this example, we parse the string "840D000000001234", which represents -12.34 USD,
// according to the specification for "DE54, Additional Amounts" in ISO 8583.
func ExampleNewAmount_iso8583() {
	a, err := ParseISO8583("840D000000001234")
	if err != nil {
		panic(err)
	}
	fmt.Println(a)
	// Output: USD -12.34
}

func ParseMoneyProto(curr string, units int64, nanos int32) (money.Amount, error) {
	c, err := money.ParseCurr(curr)
	if err != nil {
		return money.Amount{}, err
	}
	i := decimal.New(units, 0)
	f := decimal.New(int64(nanos), 9).Reduce()
	if i.Prec() > decimal.MaxPrec-c.Scale() {
		return money.Amount{}, fmt.Errorf("with currency %v, the integer part of %T can have at most %v digit(s), but it has %v digit(s)", c, money.Amount{}, decimal.MaxPrec-c.Scale(), i.Prec())
	}
	d := i.AddExact(f, c.Scale())
	return money.NewAmount(c, d)
}

// This is an example of how to a parse a monetary amount formatted as [MoneyProto].
//
// [MoneyProto]: https://github.com/googleapis/googleapis/blob/master/google/type/money.proto
func ExampleNewAmount_protobuf() {
	a, err := ParseMoneyProto("840", -12, -340000000)
	if err != nil {
		panic(err)
	}
	fmt.Println(a)
	// Output: USD -12.34
}

func ParseStripe(currency string, amount int64) (money.Amount, error) {
	c, err := money.ParseCurr(currency)
	if err != nil {
		return money.Amount{}, err
	}
	d := decimal.New(amount, c.Scale())
	return money.NewAmount(c, d)
}

// This is an example of how to a parse a monetary amount
// formatted according to Stripe API specification.
func ExampleNewAmount_stripe() {
	a, err := ParseStripe("usd", -1234)
	if err != nil {
		panic(err)
	}
	fmt.Println(a)
	// Output: USD -12.34
}

func ExampleParseAmount() {
	a, err := money.ParseAmount("USD", "-12.34")
	if err != nil {
		panic(err)
	}
	fmt.Println(a)
	// Output: USD -12.34
}

func ExampleAmount_Zero() {
	a := money.MustParseAmount("JPY", "23")
	b := money.MustParseAmount("JPY", "23.5")
	c := money.MustParseAmount("JPY", "23.56")
	fmt.Println(a.Zero())
	fmt.Println(b.Zero())
	fmt.Println(c.Zero())
	// Output:
	// JPY 0
	// JPY 0.0
	// JPY 0.00
}

func ExampleAmount_One() {
	a := money.MustParseAmount("JPY", "23")
	b := money.MustParseAmount("JPY", "23.5")
	c := money.MustParseAmount("JPY", "23.56")
	fmt.Println(a.One())
	fmt.Println(b.One())
	fmt.Println(c.One())
	// Output:
	// JPY 1
	// JPY 1.0
	// JPY 1.00
}

func ExampleAmount_ULP() {
	a := money.MustParseAmount("JPY", "23")
	b := money.MustParseAmount("JPY", "23.5")
	c := money.MustParseAmount("JPY", "23.56")
	fmt.Println(a.ULP())
	fmt.Println(b.ULP())
	fmt.Println(c.ULP())
	// Output:
	// JPY 1
	// JPY 0.1
	// JPY 0.01
}

func ExampleAmount_Cmp() {
	a := money.MustParseAmount("USD", "23")
	b := money.MustParseAmount("USD", "-15.67")
	fmt.Println(a.Cmp(b))
	fmt.Println(a.Cmp(a))
	fmt.Println(b.Cmp(a))
	// Output:
	// 1
	// 0
	// -1
}

func ExampleAmount_CmpTotal() {
	a := money.MustParseAmount("USD", "2.00")
	b := money.MustParseAmount("USD", "2.000")
	fmt.Println(a.CmpTotal(b))
	fmt.Println(a.CmpTotal(a))
	fmt.Println(b.CmpTotal(a))
	// Output:
	// 1
	// 0
	// -1
}

func ExampleAmount_Max() {
	a := money.MustParseAmount("USD", "23")
	b := money.MustParseAmount("USD", "-15.67")
	fmt.Println(a.Max(b))
	// Output: USD 23.00
}

func ExampleAmount_Min() {
	a := money.MustParseAmount("USD", "23")
	b := money.MustParseAmount("USD", "-15.67")
	fmt.Println(a.Min(b))
	// Output: USD -15.67
}

func ExampleNewExchRate() {
	d := decimal.New(12000, 4)
	r, err := money.NewExchRate(money.USD, money.EUR, d)
	if err != nil {
		panic(err)
	}
	fmt.Println(r)
	// Output: USD/EUR 1.2000
}

func ExampleParseExchRate() {
	r, err := money.ParseExchRate("USD", "EUR", "1.2000")
	if err != nil {
		panic(err)
	}
	fmt.Println(r)
	// Output: USD/EUR 1.2000
}

func ExampleMustParseExchRate() {
	r := money.MustParseExchRate("OMR", "USD", "0.38497")
	fmt.Println(r)
	// Output: OMR/USD 0.38497
}

func ExampleExchangeRate_Conv() {
	r := money.MustParseExchRate("USD", "JPY", "133.27")
	a := money.MustParseAmount("USD", "200.00")
	fmt.Println(r.Conv(a))
	// Output: JPY 26654.0000
}

func ExampleExchangeRate_Prec() {
	r := money.MustParseExchRate("USD", "EUR", "0.9097")
	q := money.MustParseExchRate("OMR", "USD", "0.38497")
	fmt.Println(r.Prec())
	fmt.Println(q.Prec())
	// Output:
	// 4
	// 5
}

func ExampleExchangeRate_Scale() {
	r := money.MustParseExchRate("USD", "EUR", "0.9097")
	q := money.MustParseExchRate("OMR", "USD", "0.38497")
	fmt.Println(r.Scale())
	fmt.Println(q.Scale())
	// Output:
	// 4
	// 5
}

func ExampleExchangeRate_Mul() {
	r := money.MustParseExchRate("USD", "EUR", "0.9000")
	d := decimal.MustParse("1.1")
	fmt.Println(r.Mul(d))
	// Output: USD/EUR 0.99000
}

func ExampleExchangeRate_Inv() {
	r := money.MustParseExchRate("EUR", "USD", "1.250")
	fmt.Println(r.Inv())
	// Output: USD/EUR 0.8000
}

func ExampleExchangeRate_Base() {
	r := money.MustParseExchRate("USD", "EUR", "0.9000")
	fmt.Println(r.Base())
	// Output: USD
}

func ExampleExchangeRate_Quote() {
	r := money.MustParseExchRate("USD", "EUR", "0.9000")
	fmt.Println(r.Quote())
	// Output: EUR
}

func ExampleExchangeRate_SameCurr() {
	r := money.MustParseExchRate("USD", "EUR", "0.9000")
	q := money.MustParseExchRate("USD", "EUR", "0.9500")
	p := money.MustParseExchRate("OMR", "EUR", "2.30000")
	fmt.Println(r.SameCurr(q))
	fmt.Println(r.SameCurr(p))
	// Output:
	// true
	// false
}

func ExampleExchangeRate_SameScale() {
	r := money.MustParseExchRate("USD", "EUR", "0.9000")
	q := money.MustParseExchRate("SAR", "USD", "0.2700")
	p := money.MustParseExchRate("OMR", "EUR", "2.30000")
	fmt.Println(r.SameScale(q))
	fmt.Println(r.SameScale(p))
	// Output:
	// true
	// false
}

func ExampleExchangeRate_SameScaleAsCurr() {
	r := money.MustParseExchRate("USD", "EUR", "0.9000")
	q := money.MustParseExchRate("SAR", "USD", "0.27000")
	p := money.MustParseExchRate("OMR", "EUR", "2.30000")
	fmt.Println(r.SameScaleAsCurr())
	fmt.Println(q.SameScaleAsCurr())
	fmt.Println(p.SameScaleAsCurr())
	// Output:
	// true
	// false
	// true
}

func ExampleExchangeRate_CanConv() {
	r := money.MustParseExchRate("USD", "EUR", "0.9000")
	a := money.MustParseAmount("USD", "123.00")
	b := money.MustParseAmount("EUR", "456.00")
	c := money.MustParseAmount("JPY", "789")
	fmt.Println(r.CanConv(a))
	fmt.Println(r.CanConv(b))
	fmt.Println(r.CanConv(c))
	// Output:
	// true
	// false
	// false
}

func ExampleExchangeRate_Format() {
	r := money.MustParseExchRate("USD", "EUR", "1.23456")
	fmt.Printf("%v\n", r)
	fmt.Printf("%f\n", r)
	fmt.Printf("%c\n", r)
	// Output:
	// USD/EUR 1.23456
	// 1.2346
	// USD/EUR
}

func ExampleExchangeRate_IsZero() {
	r := money.ExchangeRate{}
	q := money.MustParseExchRate("USD", "EUR", "1.2")
	fmt.Println(r.IsZero())
	fmt.Println(q.IsZero())
	// Output:
	// true
	// false
}

func ExampleExchangeRate_IsOne() {
	r := money.MustParseExchRate("USD", "EUR", "1")
	q := money.MustParseExchRate("USD", "EUR", "1.2")
	fmt.Println(r.IsOne())
	fmt.Println(q.IsOne())
	// Output:
	// true
	// false
}

func ExampleExchangeRate_WithinOne() {
	r := money.MustParseExchRate("EUR", "USD", "1")
	q := money.MustParseExchRate("EUR", "USD", "0.8")
	fmt.Println(r.WithinOne())
	fmt.Println(q.WithinOne())
	// Output:
	// false
	// true
}

func ExampleExchangeRate_String() {
	r := money.MustParseExchRate("USD", "EUR", "1.2345")
	fmt.Println(r.String())
	// Output: USD/EUR 1.2345
}

func ExampleExchangeRate_Round() {
	r := money.MustParseExchRate("EUR", "USD", "1.234567")
	fmt.Println(r.Round(8))
	fmt.Println(r.Round(7))
	fmt.Println(r.Round(6))
	fmt.Println(r.Round(5))
	fmt.Println(r.Round(4))
	fmt.Println(r.Round(3))
	fmt.Println(r.Round(2))
	fmt.Println(r.Round(1))
	fmt.Println(r.Round(0))
	// Output:
	// EUR/USD 1.23456700
	// EUR/USD 1.2345670
	// EUR/USD 1.234567
	// EUR/USD 1.23457
	// EUR/USD 1.2346
	// EUR/USD 1.2346
	// EUR/USD 1.2346
	// EUR/USD 1.2346
	// EUR/USD 1.2346
}

func ExampleExchangeRate_RoundToCurr() {
	r := money.MustParseExchRate("USD", "JPY", "133.859")
	q := money.MustParseExchRate("USD", "EUR", "0.915458")
	p := money.MustParseExchRate("USD", "OMR", "0.385013")
	fmt.Println(r.RoundToCurr())
	fmt.Println(q.RoundToCurr())
	fmt.Println(p.RoundToCurr())
	// Output:
	// USD/JPY 133.86
	// USD/EUR 0.9155
	// USD/OMR 0.38501
}
