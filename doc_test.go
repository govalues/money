package money_test

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"

	"github.com/govalues/decimal"
	"github.com/govalues/money"
)

func TaxAmount(price money.Amount, taxRate decimal.Decimal) (money.Amount, money.Amount, error) {
	// Subtotal
	one := taxRate.One()
	taxRate, err := taxRate.Add(one)
	if err != nil {
		return money.Amount{}, money.Amount{}, err
	}
	subtotal, err := price.Quo(taxRate)
	if err != nil {
		return money.Amount{}, money.Amount{}, err
	}

	// Function depends on the locax tax laws
	subtotal = subtotal.TruncToCurr()

	// Tax Amount
	tax, err := price.Sub(subtotal)
	if err != nil {
		return money.Amount{}, money.Amount{}, err
	}

	return subtotal, tax, nil
}

// In this example, the sales tax amount is calculated for a product with
// a given price after tax, using a specified tax rate.
func Example_taxCalculation() {
	price := money.MustParseAmount("USD", "9.99")
	taxRate := decimal.MustParse("0.0725")

	subtotal, tax, err := TaxAmount(price, taxRate)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Subtotal         = %v\n", subtotal)
	fmt.Printf("Sales tax %-6k = %v\n", taxRate, tax)
	fmt.Printf("Total price      = %v\n", price)
	// Output:
	// Subtotal         = USD 9.31
	// Sales tax 7.25%  = USD 0.68
	// Total price      = USD 9.99
}

type StatementLine struct {
	Month    int
	Days     int
	Interest money.Amount
	Balance  money.Amount
}

type Statement []StatementLine

func (s Statement) Append(month, days int, interest, balance money.Amount) Statement {
	line := StatementLine{
		Month:    month,
		Days:     days,
		Interest: interest,
		Balance:  balance,
	}
	return append(s, line)
}

func (s Statement) IncomingBalance() (money.Amount, error) {
	if len(s) == 0 {
		return money.Amount{}, fmt.Errorf("empty statement")
	}
	a, err := s[0].Balance.Sub(s[0].Interest)
	if err != nil {
		return money.Amount{}, err
	}
	return a, nil
}

func (s Statement) OutgoingBalance() (money.Amount, error) {
	if len(s) == 0 {
		return money.Amount{}, fmt.Errorf("empty statement")
	}
	return s[len(s)-1].Balance, nil
}

// PercChange computes (OutgoingBalance - IncomingBalance) / IncomingBalance.
func (s Statement) PercChange() (decimal.Decimal, error) {
	inc, err := s.IncomingBalance()
	if err != nil {
		return decimal.Decimal{}, err
	}
	out, err := s.OutgoingBalance()
	if err != nil {
		return decimal.Decimal{}, err
	}
	diff, err := out.Sub(inc)
	if err != nil {
		return decimal.Decimal{}, err
	}
	rat, err := diff.Rat(inc)
	if err != nil {
		return decimal.Decimal{}, err
	}
	return rat, nil
}

func (s Statement) TotalInterest() (money.Amount, error) {
	if len(s) == 0 {
		return money.Amount{}, fmt.Errorf("empty statement")
	}
	var err error
	total := s[0].Interest.Zero()
	for _, line := range s {
		total, err = total.Add(line.Interest)
		if err != nil {
			return money.Amount{}, err
		}
	}
	return total, nil
}

// DailyRate computes YearlyRate / 365.
func DailyRate(yearlyRate decimal.Decimal) (decimal.Decimal, error) {
	daysInYear, err := decimal.New(365, 0)
	if err != nil {
		return decimal.Decimal{}, err
	}
	dailyRate, err := yearlyRate.Quo(daysInYear)
	if err != nil {
		return decimal.Decimal{}, err
	}
	return dailyRate, nil
}

// MonthlyInterest computes Balance * DailyRate * DaysInMonth.
func MonthlyInterest(balance money.Amount, dailyRate decimal.Decimal, daysInMonth int) (money.Amount, error) {
	var err error
	interest := balance.Zero()
	for range daysInMonth {
		interest, err = interest.AddMul(balance, dailyRate)
		if err != nil {
			return money.Amount{}, err
		}
	}
	interest = interest.RoundToCurr()
	return interest, nil
}

func SimulateStatement(balance money.Amount, yearlyRate decimal.Decimal) (Statement, error) {
	daysInMonths := [...]int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	dailyRate, err := DailyRate(yearlyRate)
	if err != nil {
		return nil, err
	}
	statement := Statement{}
	for m := 0; m < len(daysInMonths); m++ {
		// Compute the interest
		interest, err := MonthlyInterest(balance, dailyRate, daysInMonths[m])
		if err != nil {
			return nil, err
		}
		// Compound the balance
		balance, err = balance.Add(interest)
		if err != nil {
			return nil, err
		}
		// Append month
		statement = statement.Append(m+1, daysInMonths[m], interest, balance)
	}
	return statement, nil
}

// This example calculates the effective interest rate for a 10% nominal
// interest rate compounded monthly on a USD 10,000 balance.
func Example_effectiveRate() {
	// Set up initial balance and nominal interest rate
	initialBalance := money.MustParseAmount("USD", "10000")
	nominalRate := decimal.MustParse("0.10")

	// Display initial balance and nominal interest rate
	fmt.Printf("Initial Balance = %v\n", initialBalance)
	fmt.Printf("Nominal Rate    = %.2k\n\n", nominalRate)

	// Generate the simulated statement for a year
	statement, err := SimulateStatement(initialBalance, nominalRate)
	if err != nil {
		panic(err)
	}

	// Display monthly balances, including the interest accrued each month
	fmt.Printf("%-5s %-5s %-12s %s\n", "Month", "Days", "Interest", "Balance")
	for _, line := range statement {
		fmt.Printf("%5v %5v %+11f %11f\n", line.Month, line.Days, line.Interest, line.Balance)
	}

	// Calculate total interest accrued over the year and effective interest rate
	totalInterest, err := statement.TotalInterest()
	if err != nil {
		panic(err)
	}
	effRate, err := statement.PercChange()
	if err != nil {
		panic(err)
	}

	// Display the total interest accrued and the effective interest rate
	fmt.Printf("      Total %+11f\n\n", totalInterest)
	fmt.Printf("Effective Rate = %.4k\n", effRate)

	// Output:
	// Initial Balance = USD 10000.00
	// Nominal Rate    = 10.00%
	//
	// Month Days  Interest     Balance
	//     1    31      +84.93    10084.93
	//     2    28      +77.36    10162.29
	//     3    31      +86.31    10248.60
	//     4    30      +84.24    10332.84
	//     5    31      +87.76    10420.60
	//     6    30      +85.65    10506.25
	//     7    31      +89.23    10595.48
	//     8    31      +89.99    10685.47
	//     9    30      +87.83    10773.30
	//    10    31      +91.50    10864.80
	//    11    30      +89.30    10954.10
	//    12    31      +93.03    11047.13
	//       Total    +1047.13
	//
	// Effective Rate = 10.4713%
}

type ScheduleLine struct {
	Period    int
	Repayment money.Amount
	Principal money.Amount
	Interest  money.Amount
	Balance   money.Amount
}

type AmortizationSchedule []ScheduleLine

func (p AmortizationSchedule) Append(period int, repayment, principal, interest, balance money.Amount) AmortizationSchedule {
	newLine := ScheduleLine{
		Period:    period,
		Repayment: repayment,
		Principal: principal,
		Interest:  interest,
		Balance:   balance,
	}
	return append(p, newLine)
}

func (p AmortizationSchedule) TotalRepayment() (money.Amount, error) {
	if len(p) == 0 {
		return money.Amount{}, fmt.Errorf("empty schedule")
	}
	var err error
	total := p[0].Repayment.Zero()
	for _, line := range p {
		total, err = total.Add(line.Repayment)
		if err != nil {
			return money.Amount{}, err
		}
	}
	return total, nil
}

func (p AmortizationSchedule) TotalPrincipal() (money.Amount, error) {
	if len(p) == 0 {
		return money.Amount{}, fmt.Errorf("empty schedule")
	}
	var err error
	total := p[0].Principal.Zero()
	for _, line := range p {
		total, err = total.Add(line.Principal)
		if err != nil {
			return money.Amount{}, err
		}
	}
	return total, nil
}

func (p AmortizationSchedule) TotalInterest() (money.Amount, error) {
	if len(p) == 0 {
		return money.Amount{}, fmt.Errorf("empty schedule")
	}
	var err error
	total := p[0].Interest.Zero()
	for _, line := range p {
		total, err = total.Add(line.Interest)
		if err != nil {
			return money.Amount{}, err
		}
	}
	return total, nil
}

// MonthlyRate computes YearlyRate / 12.
func MonthlyRate(yearlyRate decimal.Decimal) (decimal.Decimal, error) {
	monthsInYear := decimal.MustNew(12, 0)
	return yearlyRate.Quo(monthsInYear)
}

// AnnuityPayment computes Amount * Rate / (1 - (1 + Rate)^(-Periods)).
func AnnuityPayment(amount money.Amount, rate decimal.Decimal, periods int) (money.Amount, error) {
	// Denominator
	one := rate.One()
	den, err := rate.Add(one)
	if err != nil {
		return money.Amount{}, err
	}
	den, err = den.PowInt(-periods)
	if err != nil {
		return money.Amount{}, err
	}
	den, err = one.Sub(den)
	if err != nil {
		return money.Amount{}, err
	}
	// Numerator
	num, err := amount.Mul(rate)
	if err != nil {
		return money.Amount{}, err
	}
	// Payment
	res, err := num.Quo(den)
	if err != nil {
		return money.Amount{}, err
	}
	return res.RoundToCurr(), nil
}

func SimulateSchedule(balance money.Amount, yearlyRate decimal.Decimal, years int) (AmortizationSchedule, error) {
	months := years * 12
	monthlyRate, err := MonthlyRate(yearlyRate)
	if err != nil {
		return nil, err
	}
	repayment, err := AnnuityPayment(balance, monthlyRate, months)
	if err != nil {
		return nil, err
	}

	schedule := AmortizationSchedule{}

	// All periods except the last
	for i := range months - 1 {
		interest, err := balance.Mul(monthlyRate)
		if err != nil {
			return nil, err
		}
		interest = interest.RoundToCurr()
		principal, err := repayment.Sub(interest)
		if err != nil {
			return nil, err
		}
		balance, err = balance.Sub(principal)
		if err != nil {
			return nil, err
		}
		schedule = schedule.Append(i+1, repayment, principal, interest, balance)
	}

	// The last period
	interest, err := balance.Mul(monthlyRate)
	if err != nil {
		return nil, err
	}
	interest = interest.RoundToCurr()
	principal := balance
	balance = balance.Zero()
	repayment, err = principal.Add(interest)
	if err != nil {
		return nil, err
	}
	schedule = schedule.Append(months, repayment, principal, interest, balance)

	return schedule, nil
}

// In this example, a loan amortization table is generated for a loan with
// an initial amount of USD 12,000, an annual interest rate of 10%, and
// a repayment period of 1 year.
func Example_loanAmortization() {
	// Set up initial loan balance and interest rate
	initialBalance := money.MustParseAmount("USD", "12000")
	yearlyRate := decimal.MustParse("0.1")
	years := 1

	// Display the initial loan balance and interest rate
	fmt.Printf("Initial Balance = %v\n", initialBalance)
	fmt.Printf("Interest Rate   = %.2k\n\n", yearlyRate)

	// Generate the amortization schedule
	schedule, err := SimulateSchedule(initialBalance, yearlyRate, years)
	if err != nil {
		panic(err)
	}

	// Display the amortization schedule, showing the monthly
	// repayment, principal, interest and outstanding loan balance
	fmt.Println("Month  Repayment   Principal   Interest    Outstanding")
	for _, line := range schedule {
		fmt.Printf("%5d %12f %11f %11f %11f\n", line.Period, line.Repayment, line.Principal, line.Interest, line.Balance)
	}

	// Calculate and display the total amounts repaid, both principal and interest
	totalRepayment, err := schedule.TotalRepayment()
	if err != nil {
		panic(err)
	}
	totalPrincipal, err := schedule.TotalPrincipal()
	if err != nil {
		panic(err)
	}
	totalInterest, err := schedule.TotalInterest()
	if err != nil {
		panic(err)
	}

	// Display the total repayment, principal repayment and interest payment over the loan period
	fmt.Printf("Total %12f %11f %11f\n", totalRepayment, totalPrincipal, totalInterest)

	// Output:
	// Initial Balance = USD 12000.00
	// Interest Rate   = 10.00%
	//
	// Month  Repayment   Principal   Interest    Outstanding
	//     1      1054.99      954.99      100.00    11045.01
	//     2      1054.99      962.95       92.04    10082.06
	//     3      1054.99      970.97       84.02     9111.09
	//     4      1054.99      979.06       75.93     8132.03
	//     5      1054.99      987.22       67.77     7144.81
	//     6      1054.99      995.45       59.54     6149.36
	//     7      1054.99     1003.75       51.24     5145.61
	//     8      1054.99     1012.11       42.88     4133.50
	//     9      1054.99     1020.54       34.45     3112.96
	//    10      1054.99     1029.05       25.94     2083.91
	//    11      1054.99     1037.62       17.37     1046.29
	//    12      1055.01     1046.29        8.72        0.00
	// Total     12659.90    12000.00      659.90
}

func FromISO8583(s string) (money.Amount, error) {
	// Amount
	n, err := strconv.ParseInt(s[4:], 10, 64)
	if err != nil {
		return money.Amount{}, err
	}
	a, err := money.NewAmountFromMinorUnits(s[:3], n)
	if err != nil {
		return money.Amount{}, err
	}
	// Sign
	if s[3:4] == "D" {
		a = a.Neg()
	}
	return a, nil
}

// In this example, we parse the string "840D000000001234", which represents -12.34 USD,
// according to the specification for "DE54, Additional Amounts" in ISO 8583.
func Example_parsingISO8583() {
	a, _ := FromISO8583("840D000000001234")
	fmt.Println(a)
	// Output:
	// USD -12.34
}

func FromMoneyProto(curr string, units int64, nanos int32) (money.Amount, error) {
	return money.NewAmountFromInt64(curr, units, int64(nanos), 9)
}

func ToMoneyProto(a money.Amount) (curr string, units int64, nanos int32, ok bool) {
	curr = a.Curr().Code()
	whole, frac, ok := a.Int64(9)
	return curr, whole, int32(frac), ok //nolint:gosec
}

// This is an example of how to a parse a monetary amount formatted as [money.proto].
//
// [money.proto]: https://github.com/googleapis/googleapis/blob/master/google/type/money.proto
func Example_parsingProtobuf() {
	a, _ := FromMoneyProto("USD", 5, 670000000)
	fmt.Println(a)
	fmt.Println(ToMoneyProto(a))
	// Output:
	// USD 5.67
	// USD 5 670000000 true
}

func FromStripe(curr string, units int64) (money.Amount, error) {
	return money.NewAmountFromMinorUnits(curr, units)
}

func ToStripe(a money.Amount) (curr string, units int64, ok bool) {
	curr = strings.ToLower(a.Curr().Code())
	units, ok = a.MinorUnits()
	return curr, units, ok
}

// This is an example of how to a parse a monetary amount
// formatted according to [Stripe API] specification.
//
// [Stripe API]: https://stripe.com/docs/api/balance/balance_object
func Example_parsingStripe() {
	a, _ := FromStripe("usd", 567)
	fmt.Println(a)
	fmt.Println(ToStripe(a))
	// Output:
	// USD 5.67
	// usd 567 true
}

func ExampleParseCurr_currencies() {
	fmt.Println(money.ParseCurr("JPY"))
	fmt.Println(money.ParseCurr("USD"))
	fmt.Println(money.ParseCurr("OMR"))
	// Output:
	// JPY <nil>
	// USD <nil>
	// OMR <nil>
}

func ExampleParseCurr_codes() {
	fmt.Println(money.ParseCurr("usd"))
	fmt.Println(money.ParseCurr("USD"))
	fmt.Println(money.ParseCurr("840"))
	// Output:
	// USD <nil>
	// USD <nil>
	// USD <nil>
}

func ExampleMustParseCurr_currencies() {
	fmt.Println(money.MustParseCurr("JPY"))
	fmt.Println(money.MustParseCurr("USD"))
	fmt.Println(money.MustParseCurr("OMR"))
	// Output:
	// JPY
	// USD
	// OMR
}

func ExampleMustParseCurr_codes() {
	fmt.Println(money.MustParseCurr("usd"))
	fmt.Println(money.MustParseCurr("USD"))
	fmt.Println(money.MustParseCurr("840"))
	// Output:
	// USD
	// USD
	// USD
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

type Account struct {
	Balance decimal.Decimal `json:"bal_amt"`
	Curr    money.Currency  `json:"bal_curr"`
}

func ExampleCurrency_UnmarshalJSON_json() {
	var a Account
	err := json.Unmarshal([]byte(`{"bal_amt": "1.23", "bal_curr":"USD"}`), &a)
	fmt.Println(a, err)
	// Output:
	// {1.23 USD} <nil>
}

func ExampleCurrency_MarshalJSON_json() {
	a := money.MustParseAmount("USD", "5.67")
	v := Account{
		Balance: a.Decimal(),
		Curr:    a.Curr(),
	}
	b, err := json.Marshal(v)
	fmt.Println(string(b), err)
	// Output:
	// {"bal_amt":"5.67","bal_curr":"USD"} <nil>
}

type Transaction struct {
	XMLName xml.Name `xml:"Txn"`
	Amount  struct {
		Value decimal.Decimal `xml:",chardata"`
		Curr  money.Currency  `xml:"Ccy,attr"`
	} `xml:"Amt"`
}

func ExampleCurrency_UnmarshalText_xml() {
	var t Transaction
	err := xml.Unmarshal([]byte(`<Txn><Amt Ccy="USD">5.67</Amt></Txn>`), &t)
	fmt.Println(t, err)
	// Output:
	// {{ Txn} {5.67 USD}} <nil>
}

func ExampleCurrency_MarshalText_xml() {
	a := money.MustParseAmount("USD", "5.67")
	t := Transaction{}
	t.Amount.Value = a.Decimal()
	t.Amount.Curr = a.Curr()
	b, err := xml.Marshal(t)
	fmt.Println(string(b), err)
	// Output:
	// <Txn><Amt Ccy="USD">5.67</Amt></Txn> <nil>
}

func ExampleCurrency_Scan() {
	u := money.XXX
	_ = u.Scan("USD")
	fmt.Println(u)
	// Output: USD
}

func ExampleCurrency_Value() {
	u := money.USD
	fmt.Println(u.Value())
	// Output: USD <nil>
}

func ExampleCurrency_Format() {
	fmt.Printf("%c\n", money.USD)
	// Output:
	// USD
}

func ExampleCurrency_AppendBinary() {
	c := money.USD
	var data []byte
	data = append(data, 0x03)
	data, err := c.AppendBinary(data)
	data = append(data, 0x00)
	fmt.Printf("[% x] %v\n", data, err)
	// Output:
	// [03 55 53 44 00] <nil>
}

func ExampleCurrency_AppendText() {
	c := money.USD
	var text []byte
	text = append(text, "<Curr>"...)
	text, err := c.AppendText(text)
	text = append(text, "</Curr>"...)
	fmt.Printf("%s %v\n", text, err)
	// Output:
	// <Curr>USD</Curr> <nil>
}

func ExampleCurrency_MarshalBSONValue_bson() {
	c := money.USD
	typ, data, err := c.MarshalBSONValue()
	fmt.Printf("%v [% x] %v\n", typ, data, err)
	// Output:
	// 2 [04 00 00 00 55 53 44 00] <nil>
}

func ExampleCurrency_UnmarshalBSONValue_bson() {
	data := []byte{
		0x04, 0x00, 0x00, 0x00,
		0x55, 0x53, 0x44, 0x00,
	}

	var c money.Currency
	err := c.UnmarshalBSONValue(2, data)
	fmt.Println(c, err)
	// Output:
	// USD <nil>
}

func ExampleCurrency_MarshalBinary_gob() {
	data, err := marshalGOB(money.USD)
	fmt.Printf("[% x] %v\n", data, err)
	// Output:
	// [13 7f 06 01 01 08 43 75 72 72 65 6e 63 79 01 ff 80 00 00 00 07 ff 80 00 03 55 53 44] <nil>
}

func marshalGOB(c money.Currency) ([]byte, error) {
	var data bytes.Buffer
	enc := gob.NewEncoder(&data)
	err := enc.Encode(c)
	if err != nil {
		return nil, err
	}
	return data.Bytes(), nil
}

func ExampleCurrency_UnmarshalBinary_gob() {
	data := []byte{
		0x13, 0x7f, 0x06, 0x01,
		0x01, 0x08, 0x43, 0x75,
		0x72, 0x72, 0x65, 0x6e,
		0x63, 0x79, 0x01, 0xff,
		0x80, 0x00, 0x00, 0x00,
		0x07, 0xff, 0x80, 0x00,
		0x03, 0x55, 0x53, 0x44,
	}
	fmt.Println(unmarshalGOB(data))
	// Output:
	// USD <nil>
}

func unmarshalGOB(data []byte) (money.Currency, error) {
	var c money.Currency
	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&c)
	if err != nil {
		return money.XXX, err
	}
	return c, nil
}

func ExampleNullCurrency_Scan() {
	var n, m money.NullCurrency
	_ = n.Scan("USD")
	_ = m.Scan(nil)
	fmt.Println(n)
	fmt.Println(m)
	// Output:
	// {USD true}
	// {XXX false}
}

func ExampleNullCurrency_Value() {
	n := money.NullCurrency{
		Currency: money.USD,
		Valid:    true,
	}
	m := money.NullCurrency{
		Currency: money.XXX,
		Valid:    false,
	}
	fmt.Println(n.Value())
	fmt.Println(m.Value())
	// Output:
	// USD <nil>
	// <nil> <nil>
}

func ExampleNullCurrency_UnmarshalJSON_json() {
	var n money.NullCurrency
	err := json.Unmarshal([]byte(`null`), &n)
	fmt.Println(n, err)

	var m money.NullCurrency
	err = json.Unmarshal([]byte(`"USD"`), &m)
	fmt.Println(m, err)
	// Output:
	// {XXX false} <nil>
	// {USD true} <nil>
}

func ExampleNullCurrency_MarshalJSON_json() {
	n := money.NullCurrency{
		Valid: false,
	}
	text, err := json.Marshal(n)
	fmt.Println(string(text), err)

	m := money.NullCurrency{
		Currency: money.USD,
		Valid:    true,
	}
	text, err = json.Marshal(m)
	fmt.Println(string(text), err)
	// Output:
	// null <nil>
	// "USD" <nil>
}

func ExampleNullCurrency_UnmarshalBSONValue_bson() {
	var n money.NullCurrency
	err := n.UnmarshalBSONValue(10, nil)
	fmt.Println(n, err)

	data := []byte{
		0x04, 0x00, 0x00, 0x00,
		0x55, 0x53, 0x44, 0x00,
	}
	var m money.NullCurrency
	err = m.UnmarshalBSONValue(2, data)
	fmt.Println(m, err)
	// Output:
	// {XXX false} <nil>
	// {USD true} <nil>
}

func ExampleNullCurrency_MarshalBSONValue_bson() {
	n := money.NullCurrency{
		Valid: false,
	}
	t, data, err := n.MarshalBSONValue()
	fmt.Printf("%v [% x] %v\n", t, data, err)

	m := money.NullCurrency{
		Currency: money.USD,
		Valid:    true,
	}
	t, data, err = m.MarshalBSONValue()
	fmt.Printf("%v [% x] %v\n", t, data, err)
	// Output:
	// 10 [] <nil>
	// 2 [04 00 00 00 55 53 44 00] <nil>
}

func ExampleNewAmount_scales() {
	fmt.Println(money.NewAmount("USD", 567, 0))
	fmt.Println(money.NewAmount("USD", 567, 1))
	fmt.Println(money.NewAmount("USD", 567, 2))
	fmt.Println(money.NewAmount("USD", 567, 3))
	fmt.Println(money.NewAmount("USD", 567, 4))
	// Output:
	// USD 567.00 <nil>
	// USD 56.70 <nil>
	// USD 5.67 <nil>
	// USD 0.567 <nil>
	// USD 0.0567 <nil>
}

func ExampleNewAmount_currencies() {
	fmt.Println(money.NewAmount("JPY", 567, 2))
	fmt.Println(money.NewAmount("USD", 567, 2))
	fmt.Println(money.NewAmount("OMR", 567, 2))
	// Output:
	// JPY 5.67 <nil>
	// USD 5.67 <nil>
	// OMR 5.670 <nil>
}

func ExampleMustNewAmount_scales() {
	fmt.Println(money.MustNewAmount("USD", 567, 0))
	fmt.Println(money.MustNewAmount("USD", 567, 1))
	fmt.Println(money.MustNewAmount("USD", 567, 2))
	fmt.Println(money.MustNewAmount("USD", 567, 3))
	fmt.Println(money.MustNewAmount("USD", 567, 4))
	// Output:
	// USD 567.00
	// USD 56.70
	// USD 5.67
	// USD 0.567
	// USD 0.0567
}

func ExampleMustNewAmount_currencies() {
	fmt.Println(money.MustNewAmount("JPY", 567, 2))
	fmt.Println(money.MustNewAmount("USD", 567, 2))
	fmt.Println(money.MustNewAmount("OMR", 567, 2))
	// Output:
	// JPY 5.67
	// USD 5.67
	// OMR 5.670
}

func ExampleNewAmountFromDecimal() {
	d := decimal.MustParse("5.67")
	fmt.Println(money.NewAmountFromDecimal(money.JPY, d))
	fmt.Println(money.NewAmountFromDecimal(money.USD, d))
	fmt.Println(money.NewAmountFromDecimal(money.OMR, d))
	// Output:
	// JPY 5.67 <nil>
	// USD 5.67 <nil>
	// OMR 5.670 <nil>
}

func ExampleNewAmountFromInt64_scales() {
	fmt.Println(money.NewAmountFromInt64("USD", 5, 67, 2))
	fmt.Println(money.NewAmountFromInt64("USD", 5, 67, 3))
	fmt.Println(money.NewAmountFromInt64("USD", 5, 67, 4))
	fmt.Println(money.NewAmountFromInt64("USD", 5, 67, 5))
	fmt.Println(money.NewAmountFromInt64("USD", 5, 67, 6))
	// Output:
	// USD 5.67 <nil>
	// USD 5.067 <nil>
	// USD 5.0067 <nil>
	// USD 5.00067 <nil>
	// USD 5.000067 <nil>
}

func ExampleNewAmountFromInt64_currencies() {
	fmt.Println(money.NewAmountFromInt64("JPY", 5, 67, 2))
	fmt.Println(money.NewAmountFromInt64("USD", 5, 67, 2))
	fmt.Println(money.NewAmountFromInt64("OMR", 5, 67, 2))
	// Output:
	// JPY 5.67 <nil>
	// USD 5.67 <nil>
	// OMR 5.670 <nil>
}

func ExampleNewAmountFromFloat64_currencies() {
	fmt.Println(money.NewAmountFromFloat64("JPY", 5.67e0))
	fmt.Println(money.NewAmountFromFloat64("USD", 5.67e0))
	fmt.Println(money.NewAmountFromFloat64("OMR", 5.67e0))
	// Output:
	// JPY 5.67 <nil>
	// USD 5.67 <nil>
	// OMR 5.670 <nil>
}

func ExampleNewAmountFromFloat64_scales() {
	fmt.Println(money.NewAmountFromFloat64("USD", 5.67e-2))
	fmt.Println(money.NewAmountFromFloat64("USD", 5.67e-1))
	fmt.Println(money.NewAmountFromFloat64("USD", 5.67e0))
	fmt.Println(money.NewAmountFromFloat64("USD", 5.67e1))
	fmt.Println(money.NewAmountFromFloat64("USD", 5.67e2))
	// Output:
	// USD 0.0567 <nil>
	// USD 0.567 <nil>
	// USD 5.67 <nil>
	// USD 56.70 <nil>
	// USD 567.00 <nil>
}

func ExampleNewAmountFromMinorUnits_currencies() {
	fmt.Println(money.NewAmountFromMinorUnits("JPY", 567))
	fmt.Println(money.NewAmountFromMinorUnits("USD", 567))
	fmt.Println(money.NewAmountFromMinorUnits("OMR", 567))
	// Output:
	// JPY 567 <nil>
	// USD 5.67 <nil>
	// OMR 0.567 <nil>
}

func ExampleNewAmountFromMinorUnits_scales() {
	fmt.Println(money.NewAmountFromMinorUnits("USD", 5))
	fmt.Println(money.NewAmountFromMinorUnits("USD", 56))
	fmt.Println(money.NewAmountFromMinorUnits("USD", 567))
	fmt.Println(money.NewAmountFromMinorUnits("USD", 5670))
	fmt.Println(money.NewAmountFromMinorUnits("USD", 56700))
	// Output:
	// USD 0.05 <nil>
	// USD 0.56 <nil>
	// USD 5.67 <nil>
	// USD 56.70 <nil>
	// USD 567.00 <nil>
}

func ExampleMustParseAmount_currencies() {
	fmt.Println(money.MustParseAmount("JPY", "5.67"))
	fmt.Println(money.MustParseAmount("USD", "5.67"))
	fmt.Println(money.MustParseAmount("OMR", "5.67"))
	// Output:
	// JPY 5.67
	// USD 5.67
	// OMR 5.670
}

func ExampleMustParseAmount_scales() {
	fmt.Println(money.MustParseAmount("USD", "0.0567"))
	fmt.Println(money.MustParseAmount("USD", "0.567"))
	fmt.Println(money.MustParseAmount("USD", "5.67"))
	fmt.Println(money.MustParseAmount("USD", "56.7"))
	fmt.Println(money.MustParseAmount("USD", "567"))
	// Output:
	// USD 0.0567
	// USD 0.567
	// USD 5.67
	// USD 56.70
	// USD 567.00
}

func ExampleParseAmount_currencies() {
	fmt.Println(money.ParseAmount("JPY", "5.67"))
	fmt.Println(money.ParseAmount("USD", "5.67"))
	fmt.Println(money.ParseAmount("OMR", "5.67"))
	// Output:
	// JPY 5.67 <nil>
	// USD 5.67 <nil>
	// OMR 5.670 <nil>
}

func ExampleParseAmount_scales() {
	fmt.Println(money.ParseAmount("USD", "0.0567"))
	fmt.Println(money.ParseAmount("USD", "0.567"))
	fmt.Println(money.ParseAmount("USD", "5.67"))
	fmt.Println(money.ParseAmount("USD", "56.7"))
	fmt.Println(money.ParseAmount("USD", "567"))
	// Output:
	// USD 0.0567 <nil>
	// USD 0.567 <nil>
	// USD 5.67 <nil>
	// USD 56.70 <nil>
	// USD 567.00 <nil>
}

func ExampleAmount_MinorUnits_currencies() {
	a := money.MustParseAmount("JPY", "5.678")
	b := money.MustParseAmount("USD", "5.678")
	c := money.MustParseAmount("OMR", "5.678")
	fmt.Println(a.MinorUnits())
	fmt.Println(b.MinorUnits())
	fmt.Println(c.MinorUnits())
	// Output:
	// 6 true
	// 568 true
	// 5678 true
}

func ExampleAmount_MinorUnits_scales() {
	a := money.MustParseAmount("USD", "0.0567")
	b := money.MustParseAmount("USD", "0.567")
	c := money.MustParseAmount("USD", "5.67")
	d := money.MustParseAmount("USD", "56.7")
	e := money.MustParseAmount("USD", "567")
	fmt.Println(a.MinorUnits())
	fmt.Println(b.MinorUnits())
	fmt.Println(c.MinorUnits())
	fmt.Println(d.MinorUnits())
	fmt.Println(e.MinorUnits())
	// Output:
	// 6 true
	// 57 true
	// 567 true
	// 5670 true
	// 56700 true
}

func ExampleAmount_Float64() {
	a := money.MustParseAmount("USD", "0.10")
	b := money.MustParseAmount("USD", "123.456")
	c := money.MustParseAmount("USD", "1234567890.123456789")
	fmt.Println(a.Float64())
	fmt.Println(b.Float64())
	fmt.Println(c.Float64())
	// Output:
	// 0.1 true
	// 123.456 true
	// 1.2345678901234567e+09 true
}

func ExampleAmount_Int64() {
	a := money.MustParseAmount("USD", "5.678")
	fmt.Println(a.Int64(0))
	fmt.Println(a.Int64(1))
	fmt.Println(a.Int64(2))
	fmt.Println(a.Int64(3))
	fmt.Println(a.Int64(4))
	// Output:
	// 6 0 true
	// 5 7 true
	// 5 68 true
	// 5 678 true
	// 5 6780 true
}

func ExampleAmount_Curr() {
	a := money.MustParseAmount("USD", "5.67")
	fmt.Println(a.Curr())
	// Output: USD
}

func ExampleAmount_Decimal() {
	a := money.MustParseAmount("USD", "5.67")
	fmt.Println(a.Decimal())
	// Output: 5.67
}

func ExampleAmount_Add() {
	a := money.MustParseAmount("USD", "5.67")
	b := money.MustParseAmount("USD", "23.00")
	fmt.Println(a.Add(b))
	// Output: USD 28.67 <nil>
}

func ExampleAmount_Sub() {
	a := money.MustParseAmount("USD", "5.67")
	b := money.MustParseAmount("USD", "23.00")
	fmt.Println(a.Sub(b))
	// Output: USD -17.33 <nil>
}

func ExampleAmount_SubAbs() {
	a := money.MustParseAmount("USD", "5.67")
	b := money.MustParseAmount("USD", "23.00")
	fmt.Println(a.SubAbs(b))
	// Output: USD 17.33 <nil>
}

func ExampleAmount_AddMul() {
	a := money.MustParseAmount("USD", "5.67")
	b := money.MustParseAmount("USD", "23.00")
	e := decimal.MustParse("2")
	fmt.Println(a.AddMul(b, e))
	// Output: USD 51.67 <nil>
}

func ExampleAmount_AddQuo() {
	a := money.MustParseAmount("USD", "5.67")
	b := money.MustParseAmount("USD", "23.00")
	e := decimal.MustParse("2")
	fmt.Println(a.AddQuo(b, e))
	// Output: USD 17.17 <nil>
}

func ExampleAmount_SubMul() {
	a := money.MustParseAmount("USD", "5.67")
	b := money.MustParseAmount("USD", "23.00")
	e := decimal.MustParse("2")
	fmt.Println(a.SubMul(b, e))
	// Output: USD -40.33 <nil>
}

func ExampleAmount_SubQuo() {
	a := money.MustParseAmount("USD", "5.67")
	b := money.MustParseAmount("USD", "23.00")
	e := decimal.MustParse("2")
	fmt.Println(a.SubQuo(b, e))
	// Output: USD -5.83 <nil>
}

func ExampleAmount_Mul() {
	a := money.MustParseAmount("USD", "5.67")
	e := decimal.MustParse("2")
	fmt.Println(a.Mul(e))
	// Output: USD 11.34 <nil>
}

func ExampleAmount_Quo() {
	a := money.MustParseAmount("USD", "5.67")
	e := decimal.MustParse("2")
	fmt.Println(a.Quo(e))
	// Output: USD 2.835 <nil>
}

func ExampleAmount_QuoRem() {
	a := money.MustParseAmount("JPY", "5.67")
	b := money.MustParseAmount("USD", "5.67")
	c := money.MustParseAmount("OMR", "5.67")
	e := decimal.MustParse("2")
	fmt.Println(a.QuoRem(e))
	fmt.Println(b.QuoRem(e))
	fmt.Println(c.QuoRem(e))
	// Output:
	// JPY 2 JPY 1.67 <nil>
	// USD 2.83 USD 0.01 <nil>
	// OMR 2.835 OMR 0.000 <nil>
}

func ExampleAmount_Split_scales() {
	a := money.MustParseAmount("USD", "0.0567")
	b := money.MustParseAmount("USD", "0.567")
	c := money.MustParseAmount("USD", "5.67")
	d := money.MustParseAmount("USD", "56.7")
	e := money.MustParseAmount("USD", "567")
	fmt.Println(a.Split(2))
	fmt.Println(b.Split(2))
	fmt.Println(c.Split(2))
	fmt.Println(d.Split(2))
	fmt.Println(e.Split(2))
	// Output:
	// [USD 0.0284 USD 0.0283] <nil>
	// [USD 0.284 USD 0.283] <nil>
	// [USD 2.84 USD 2.83] <nil>
	// [USD 28.35 USD 28.35] <nil>
	// [USD 283.50 USD 283.50] <nil>
}

func ExampleAmount_Split_parts() {
	a := money.MustParseAmount("USD", "5.67")
	fmt.Println(a.Split(1))
	fmt.Println(a.Split(2))
	fmt.Println(a.Split(3))
	fmt.Println(a.Split(4))
	fmt.Println(a.Split(5))
	// Output:
	// [USD 5.67] <nil>
	// [USD 2.84 USD 2.83] <nil>
	// [USD 1.89 USD 1.89 USD 1.89] <nil>
	// [USD 1.42 USD 1.42 USD 1.42 USD 1.41] <nil>
	// [USD 1.14 USD 1.14 USD 1.13 USD 1.13 USD 1.13] <nil>
}

func ExampleAmount_Rat() {
	a := money.MustParseAmount("EUR", "8")
	b := money.MustParseAmount("USD", "10")
	fmt.Println(a.Rat(b))
	// Output: 0.8 <nil>
}

func ExampleAmount_Rescale_currencies() {
	a := money.MustParseAmount("JPY", "5.678")
	b := money.MustParseAmount("USD", "5.678")
	c := money.MustParseAmount("OMR", "5.678")
	fmt.Println(a.Rescale(0))
	fmt.Println(b.Rescale(0))
	fmt.Println(c.Rescale(0))
	// Output:
	// JPY 6
	// USD 6.00
	// OMR 6.000
}

func ExampleAmount_Rescale_scales() {
	a := money.MustParseAmount("USD", "5.6789")
	fmt.Println(a.Rescale(0))
	fmt.Println(a.Rescale(1))
	fmt.Println(a.Rescale(2))
	fmt.Println(a.Rescale(3))
	fmt.Println(a.Rescale(4))
	// Output:
	// USD 6.00
	// USD 5.70
	// USD 5.68
	// USD 5.679
	// USD 5.6789
}

func ExampleAmount_Round_currencies() {
	a := money.MustParseAmount("JPY", "5.678")
	b := money.MustParseAmount("USD", "5.678")
	c := money.MustParseAmount("OMR", "5.678")
	fmt.Println(a.Round(0))
	fmt.Println(b.Round(0))
	fmt.Println(c.Round(0))
	// Output:
	// JPY 6
	// USD 6.00
	// OMR 6.000
}

func ExampleAmount_Round_scales() {
	a := money.MustParseAmount("USD", "5.6789")
	fmt.Println(a.Round(0))
	fmt.Println(a.Round(1))
	fmt.Println(a.Round(2))
	fmt.Println(a.Round(3))
	fmt.Println(a.Round(4))
	// Output:
	// USD 6.00
	// USD 5.70
	// USD 5.68
	// USD 5.679
	// USD 5.6789
}

func ExampleAmount_RoundToCurr() {
	a := money.MustParseAmount("JPY", "5.678")
	b := money.MustParseAmount("USD", "5.678")
	c := money.MustParseAmount("OMR", "5.678")
	fmt.Println(a.RoundToCurr())
	fmt.Println(b.RoundToCurr())
	fmt.Println(c.RoundToCurr())
	// Output:
	// JPY 6
	// USD 5.68
	// OMR 5.678
}

func ExampleAmount_Quantize() {
	a := money.MustParseAmount("JPY", "5.678")
	x := money.MustParseAmount("JPY", "1")
	y := money.MustParseAmount("JPY", "0.1")
	z := money.MustParseAmount("JPY", "0.01")
	fmt.Println(a.Quantize(x))
	fmt.Println(a.Quantize(y))
	fmt.Println(a.Quantize(z))
	// Output:
	// JPY 6
	// JPY 5.7
	// JPY 5.68
}

func ExampleAmount_Ceil_currencies() {
	a := money.MustParseAmount("JPY", "5.678")
	b := money.MustParseAmount("USD", "5.678")
	c := money.MustParseAmount("OMR", "5.678")
	fmt.Println(a.Ceil(0))
	fmt.Println(b.Ceil(0))
	fmt.Println(c.Ceil(0))
	// Output:
	// JPY 6
	// USD 6.00
	// OMR 6.000
}

func ExampleAmount_Ceil_scales() {
	a := money.MustParseAmount("USD", "5.6789")
	fmt.Println(a.Ceil(0))
	fmt.Println(a.Ceil(1))
	fmt.Println(a.Ceil(2))
	fmt.Println(a.Ceil(3))
	fmt.Println(a.Ceil(4))
	// Output:
	// USD 6.00
	// USD 5.70
	// USD 5.68
	// USD 5.679
	// USD 5.6789
}

func ExampleAmount_CeilToCurr() {
	a := money.MustParseAmount("JPY", "5.678")
	b := money.MustParseAmount("USD", "5.678")
	c := money.MustParseAmount("OMR", "5.678")
	fmt.Println(a.CeilToCurr())
	fmt.Println(b.CeilToCurr())
	fmt.Println(c.CeilToCurr())
	// Output:
	// JPY 6
	// USD 5.68
	// OMR 5.678
}

func ExampleAmount_Floor_currencies() {
	a := money.MustParseAmount("JPY", "5.678")
	b := money.MustParseAmount("USD", "5.678")
	c := money.MustParseAmount("OMR", "5.678")
	fmt.Println(a.Floor(0))
	fmt.Println(b.Floor(0))
	fmt.Println(c.Floor(0))
	// Output:
	// JPY 5
	// USD 5.00
	// OMR 5.000
}

func ExampleAmount_Floor_scales() {
	a := money.MustParseAmount("USD", "5.6789")
	fmt.Println(a.Floor(0))
	fmt.Println(a.Floor(1))
	fmt.Println(a.Floor(2))
	fmt.Println(a.Floor(3))
	fmt.Println(a.Floor(4))
	// Output:
	// USD 5.00
	// USD 5.60
	// USD 5.67
	// USD 5.678
	// USD 5.6789
}

func ExampleAmount_FloorToCurr() {
	a := money.MustParseAmount("JPY", "5.678")
	b := money.MustParseAmount("USD", "5.678")
	c := money.MustParseAmount("OMR", "5.678")
	fmt.Println(a.FloorToCurr())
	fmt.Println(b.FloorToCurr())
	fmt.Println(c.FloorToCurr())
	// Output:
	// JPY 5
	// USD 5.67
	// OMR 5.678
}

func ExampleAmount_Trunc_currencies() {
	a := money.MustParseAmount("JPY", "5.678")
	b := money.MustParseAmount("USD", "5.678")
	c := money.MustParseAmount("OMR", "5.678")
	fmt.Println(a.Trunc(0))
	fmt.Println(b.Trunc(0))
	fmt.Println(c.Trunc(0))
	// Output:
	// JPY 5
	// USD 5.00
	// OMR 5.000
}

func ExampleAmount_Trunc_scales() {
	a := money.MustParseAmount("USD", "5.6789")
	fmt.Println(a.Trunc(0))
	fmt.Println(a.Trunc(1))
	fmt.Println(a.Trunc(2))
	fmt.Println(a.Trunc(3))
	fmt.Println(a.Trunc(4))
	// Output:
	// USD 5.00
	// USD 5.60
	// USD 5.67
	// USD 5.678
	// USD 5.6789
}

func ExampleAmount_TruncToCurr() {
	a := money.MustParseAmount("JPY", "5.678")
	b := money.MustParseAmount("USD", "5.678")
	c := money.MustParseAmount("OMR", "5.678")
	fmt.Println(a.TruncToCurr())
	fmt.Println(b.TruncToCurr())
	fmt.Println(c.TruncToCurr())
	// Output:
	// JPY 5
	// USD 5.67
	// OMR 5.678
}

func ExampleAmount_Trim_currencies() {
	a := money.MustParseAmount("JPY", "5.000")
	b := money.MustParseAmount("USD", "5.000")
	c := money.MustParseAmount("OMR", "5.000")
	fmt.Println(a.Trim(0))
	fmt.Println(b.Trim(0))
	fmt.Println(c.Trim(0))
	// Output:
	// JPY 5
	// USD 5.00
	// OMR 5.000
}

func ExampleAmount_Trim_scales() {
	a := money.MustParseAmount("USD", "5.0000")
	fmt.Println(a.Trim(0))
	fmt.Println(a.Trim(1))
	fmt.Println(a.Trim(2))
	fmt.Println(a.Trim(3))
	fmt.Println(a.Trim(4))
	// Output:
	// USD 5.00
	// USD 5.00
	// USD 5.00
	// USD 5.000
	// USD 5.0000
}

func ExampleAmount_TrimToCurr() {
	a := money.MustParseAmount("JPY", "5.000")
	b := money.MustParseAmount("USD", "5.000")
	c := money.MustParseAmount("OMR", "5.000")
	fmt.Println(a.TrimToCurr())
	fmt.Println(b.TrimToCurr())
	fmt.Println(c.TrimToCurr())
	// Output:
	// JPY 5
	// USD 5.00
	// OMR 5.000
}

func ExampleAmount_SameCurr() {
	a := money.MustParseAmount("JPY", "23")
	b := money.MustParseAmount("USD", "5.67")
	c := money.MustParseAmount("USD", "1.23")
	fmt.Println(a.SameCurr(b))
	fmt.Println(b.SameCurr(c))
	// Output:
	// false
	// true
}

func ExampleAmount_SameScale() {
	a := money.MustParseAmount("JPY", "23")
	b := money.MustParseAmount("USD", "5.67")
	c := money.MustParseAmount("USD", "1.23")
	fmt.Println(a.SameScale(b))
	fmt.Println(b.SameScale(c))
	// Output:
	// false
	// true
}

func ExampleAmount_SameScaleAsCurr() {
	a := money.MustParseAmount("USD", "5.67")
	b := money.MustParseAmount("USD", "5.678")
	c := money.MustParseAmount("USD", "5.6789")
	fmt.Println(a.SameScaleAsCurr())
	fmt.Println(b.SameScaleAsCurr())
	fmt.Println(c.SameScaleAsCurr())
	// Output:
	// true
	// false
	// false
}

func ExampleAmount_Scale() {
	a := money.MustParseAmount("USD", "23.0000")
	b := money.MustParseAmount("USD", "5.67")
	fmt.Println(a.Scale())
	fmt.Println(b.Scale())
	// Output:
	// 4
	// 2
}

func ExampleAmount_MinScale() {
	a := money.MustParseAmount("USD", "5.6000")
	b := money.MustParseAmount("USD", "5.6700")
	c := money.MustParseAmount("USD", "5.6780")
	fmt.Println(a.MinScale())
	fmt.Println(b.MinScale())
	fmt.Println(c.MinScale())
	// Output:
	// 1
	// 2
	// 3
}

func ExampleAmount_Format_verbs() {
	a := money.MustParseAmount("USD", "5.678")
	fmt.Printf("%v\n", a)
	fmt.Printf("%[1]f %[1]c\n", a)
	fmt.Printf("%f\n", a)
	fmt.Printf("%d\n", a)
	fmt.Printf("%c\n", a)
	// Output:
	// USD 5.678
	// 5.678 USD
	// 5.678
	// 568
	// USD
}

func ExampleAmount_Format_currencies() {
	a := money.MustParseAmount("JPY", "5")
	b := money.MustParseAmount("USD", "5")
	c := money.MustParseAmount("OMR", "5")
	fmt.Printf("| %s        | %s    | %s   | %s  |\n", "%v", "%f", "%d", "%c")
	fmt.Printf("| --------- | ----- | ---- | --- |\n")
	fmt.Printf("| %-9[1]v | %5[1]f | %4[1]d | %[1]c |\n", a)
	fmt.Printf("| %-9[1]v | %5[1]f | %4[1]d | %[1]c |\n", b)
	fmt.Printf("| %-9[1]v | %5[1]f | %4[1]d | %[1]c |\n", c)
	// Output:
	// | %v        | %f    | %d   | %c  |
	// | --------- | ----- | ---- | --- |
	// | JPY 5     |     5 |    5 | JPY |
	// | USD 5.00  |  5.00 |  500 | USD |
	// | OMR 5.000 | 5.000 | 5000 | OMR |
}

func ExampleAmount_String() {
	a := money.MustParseAmount("USD", "5.67")
	b := money.MustParseAmount("EUR", "-0.010000")
	fmt.Println(a.String())
	fmt.Println(b.String())
	// Output:
	// USD 5.67
	// EUR -0.010000
}

func ExampleAmount_Abs() {
	a := money.MustParseAmount("USD", "-5.67")
	fmt.Println(a.Abs())
	// Output: USD 5.67
}

func ExampleAmount_Neg() {
	a := money.MustParseAmount("USD", "5.67")
	fmt.Println(a.Neg())
	// Output: USD -5.67
}

func ExampleAmount_CopySign() {
	a := money.MustParseAmount("USD", "23.00")
	b := money.MustParseAmount("USD", "-5.67")
	fmt.Println(a.CopySign(b))
	fmt.Println(b.CopySign(a))
	// Output:
	// USD -23.00
	// USD 5.67
}

func ExampleAmount_Sign() {
	a := money.MustParseAmount("USD", "-5.67")
	b := money.MustParseAmount("USD", "23.00")
	c := money.MustParseAmount("USD", "0.00")
	fmt.Println(a.Sign())
	fmt.Println(b.Sign())
	fmt.Println(c.Sign())
	// Output:
	// -1
	// 1
	// 0
}

func ExampleAmount_IsNeg() {
	a := money.MustParseAmount("USD", "-5.67")
	b := money.MustParseAmount("USD", "23.00")
	c := money.MustParseAmount("USD", "0.00")
	fmt.Println(a.IsNeg())
	fmt.Println(b.IsNeg())
	fmt.Println(c.IsNeg())
	// Output:
	// true
	// false
	// false
}

func ExampleAmount_IsZero() {
	a := money.MustParseAmount("USD", "-5.67")
	b := money.MustParseAmount("USD", "23.00")
	c := money.MustParseAmount("USD", "0.00")
	fmt.Println(a.IsZero())
	fmt.Println(b.IsZero())
	fmt.Println(c.IsZero())
	// Output:
	// false
	// false
	// true
}

func ExampleAmount_IsOne() {
	a := money.MustParseAmount("USD", "1.00")
	b := money.MustParseAmount("USD", "2.00")
	fmt.Println(a.IsOne())
	fmt.Println(b.IsOne())
	// Output:
	// true
	// false
}

func ExampleAmount_WithinOne() {
	a := money.MustParseAmount("USD", "1.00")
	b := money.MustParseAmount("USD", "0.99")
	c := money.MustParseAmount("USD", "-1.00")
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
	a := money.MustParseAmount("USD", "-5.67")
	b := money.MustParseAmount("USD", "23.00")
	c := money.MustParseAmount("USD", "0.00")
	fmt.Println(a.IsPos())
	fmt.Println(b.IsPos())
	fmt.Println(c.IsPos())
	// Output:
	// false
	// true
	// false
}

func ExampleAmount_Zero() {
	a := money.MustParseAmount("JPY", "5")
	b := money.MustParseAmount("JPY", "5.6")
	c := money.MustParseAmount("JPY", "5.67")
	fmt.Println(a.Zero())
	fmt.Println(b.Zero())
	fmt.Println(c.Zero())
	// Output:
	// JPY 0
	// JPY 0.0
	// JPY 0.00
}

func ExampleAmount_One() {
	a := money.MustParseAmount("JPY", "5")
	b := money.MustParseAmount("JPY", "5.6")
	c := money.MustParseAmount("JPY", "5.67")
	fmt.Println(a.One())
	fmt.Println(b.One())
	fmt.Println(c.One())
	// Output:
	// JPY 1
	// JPY 1.0
	// JPY 1.00
}

func ExampleAmount_ULP() {
	a := money.MustParseAmount("JPY", "5")
	b := money.MustParseAmount("JPY", "5.6")
	c := money.MustParseAmount("JPY", "5.67")
	fmt.Println(a.ULP())
	fmt.Println(b.ULP())
	fmt.Println(c.ULP())
	// Output:
	// JPY 1
	// JPY 0.1
	// JPY 0.01
}

func ExampleAmount_Cmp() {
	a := money.MustParseAmount("USD", "-23.00")
	b := money.MustParseAmount("USD", "5.67")
	fmt.Println(a.Cmp(b))
	fmt.Println(a.Cmp(a))
	fmt.Println(b.Cmp(a))
	// Output:
	// -1 <nil>
	// 0 <nil>
	// 1 <nil>
}

func ExampleAmount_CmpAbs() {
	a := money.MustParseAmount("USD", "-23.00")
	b := money.MustParseAmount("USD", "5.67")
	fmt.Println(a.CmpAbs(b))
	fmt.Println(a.CmpAbs(a))
	fmt.Println(b.CmpAbs(a))
	// Output:
	// 1 <nil>
	// 0 <nil>
	// -1 <nil>
}

func ExampleAmount_CmpTotal() {
	a := money.MustParseAmount("USD", "2.00")
	b := money.MustParseAmount("USD", "2.000")
	fmt.Println(a.CmpTotal(b))
	fmt.Println(a.CmpTotal(a))
	fmt.Println(b.CmpTotal(a))
	// Output:
	// 1 <nil>
	// 0 <nil>
	// -1 <nil>
}

func ExampleAmount_Less() {
	a := money.MustParseAmount("USD", "-23")
	b := money.MustParseAmount("USD", "5.67")
	fmt.Println(a.Less(b))
	fmt.Println(b.Less(a))
	// Output:
	// true <nil>
	// false <nil>
}

func ExampleAmount_Equal() {
	a := money.MustParseAmount("USD", "-23")
	b := money.MustParseAmount("USD", "5.67")
	fmt.Println(a.Equal(b))
	fmt.Println(a.Equal(a))
	// Output:
	// false <nil>
	// true <nil>
}

func ExampleAmount_Max() {
	a := money.MustParseAmount("USD", "23.00")
	b := money.MustParseAmount("USD", "-5.67")
	fmt.Println(a.Max(b))
	// Output: USD 23.00 <nil>
}

func ExampleAmount_Min() {
	a := money.MustParseAmount("USD", "23.00")
	b := money.MustParseAmount("USD", "-5.67")
	fmt.Println(a.Min(b))
	// Output: USD -5.67 <nil>
}

//nolint:revive
func ExampleAmount_Clamp() {
	min := money.MustParseAmount("USD", "-20")
	max := money.MustParseAmount("USD", "20")
	a := money.MustParseAmount("USD", "-5.67")
	b := money.MustParseAmount("USD", "0")
	c := money.MustParseAmount("USD", "23")
	fmt.Println(a.Clamp(min, max))
	fmt.Println(b.Clamp(min, max))
	fmt.Println(c.Clamp(min, max))
	// Output:
	// USD -5.67 <nil>
	// USD 0.00 <nil>
	// USD 20.00 <nil>
}

func ExampleMustNewExchRate_scales() {
	fmt.Println(money.MustNewExchRate("EUR", "USD", 567, 0))
	fmt.Println(money.MustNewExchRate("EUR", "USD", 567, 1))
	fmt.Println(money.MustNewExchRate("EUR", "USD", 567, 2))
	fmt.Println(money.MustNewExchRate("EUR", "USD", 567, 3))
	fmt.Println(money.MustNewExchRate("EUR", "USD", 567, 4))
	// Output:
	// EUR/USD 567.00
	// EUR/USD 56.70
	// EUR/USD 5.67
	// EUR/USD 0.567
	// EUR/USD 0.0567
}

func ExampleMustNewExchRate_currencies() {
	fmt.Println(money.MustNewExchRate("EUR", "JPY", 567, 2))
	fmt.Println(money.MustNewExchRate("EUR", "USD", 567, 2))
	fmt.Println(money.MustNewExchRate("EUR", "OMR", 567, 2))
	// Output:
	// EUR/JPY 5.67
	// EUR/USD 5.67
	// EUR/OMR 5.670
}

func ExampleNewExchRate_scales() {
	fmt.Println(money.NewExchRate("EUR", "USD", 567, 0))
	fmt.Println(money.NewExchRate("EUR", "USD", 567, 1))
	fmt.Println(money.NewExchRate("EUR", "USD", 567, 2))
	fmt.Println(money.NewExchRate("EUR", "USD", 567, 3))
	fmt.Println(money.NewExchRate("EUR", "USD", 567, 4))
	// Output:
	// EUR/USD 567.00 <nil>
	// EUR/USD 56.70 <nil>
	// EUR/USD 5.67 <nil>
	// EUR/USD 0.567 <nil>
	// EUR/USD 0.0567 <nil>
}

func ExampleNewExchRate_currencies() {
	fmt.Println(money.NewExchRate("EUR", "JPY", 567, 2))
	fmt.Println(money.NewExchRate("EUR", "USD", 567, 2))
	fmt.Println(money.NewExchRate("EUR", "OMR", 567, 2))
	// Output:
	// EUR/JPY 5.67 <nil>
	// EUR/USD 5.67 <nil>
	// EUR/OMR 5.670 <nil>
}

func ExampleNewExchRateFromDecimal() {
	r := decimal.MustParse("5.67")
	fmt.Println(money.NewExchRateFromDecimal(money.EUR, money.JPY, r))
	fmt.Println(money.NewExchRateFromDecimal(money.EUR, money.USD, r))
	fmt.Println(money.NewExchRateFromDecimal(money.EUR, money.OMR, r))
	// Output:
	// EUR/JPY 5.67 <nil>
	// EUR/USD 5.67 <nil>
	// EUR/OMR 5.670 <nil>
}

func ExampleMustParseExchRate_currencies() {
	fmt.Println(money.MustParseExchRate("EUR", "JPY", "5.67"))
	fmt.Println(money.MustParseExchRate("EUR", "USD", "5.67"))
	fmt.Println(money.MustParseExchRate("EUR", "OMR", "5.67"))
	// Output:
	// EUR/JPY 5.67
	// EUR/USD 5.67
	// EUR/OMR 5.670
}

func ExampleMustParseExchRate_scales() {
	fmt.Println(money.MustParseExchRate("EUR", "USD", "0.0567"))
	fmt.Println(money.MustParseExchRate("EUR", "USD", "0.567"))
	fmt.Println(money.MustParseExchRate("EUR", "USD", "5.67"))
	fmt.Println(money.MustParseExchRate("EUR", "USD", "56.7"))
	fmt.Println(money.MustParseExchRate("EUR", "USD", "567"))
	// Output:
	// EUR/USD 0.0567
	// EUR/USD 0.567
	// EUR/USD 5.67
	// EUR/USD 56.70
	// EUR/USD 567.00
}

func ExampleParseExchRate_currencies() {
	fmt.Println(money.ParseExchRate("EUR", "JPY", "5.67"))
	fmt.Println(money.ParseExchRate("EUR", "USD", "5.67"))
	fmt.Println(money.ParseExchRate("EUR", "OMR", "5.67"))
	// Output:
	// EUR/JPY 5.67 <nil>
	// EUR/USD 5.67 <nil>
	// EUR/OMR 5.670 <nil>
}

func ExampleParseExchRate_scales() {
	fmt.Println(money.ParseExchRate("EUR", "USD", "0.0567"))
	fmt.Println(money.ParseExchRate("EUR", "USD", "0.567"))
	fmt.Println(money.ParseExchRate("EUR", "USD", "5.67"))
	fmt.Println(money.ParseExchRate("EUR", "USD", "56.7"))
	fmt.Println(money.ParseExchRate("EUR", "USD", "567"))
	// Output:
	// EUR/USD 0.0567 <nil>
	// EUR/USD 0.567 <nil>
	// EUR/USD 5.67 <nil>
	// EUR/USD 56.70 <nil>
	// EUR/USD 567.00 <nil>
}

func ExampleNewExchRateFromFloat64_currencies() {
	fmt.Println(money.NewExchRateFromFloat64("EUR", "JPY", 5.67e0))
	fmt.Println(money.NewExchRateFromFloat64("EUR", "USD", 5.67e0))
	fmt.Println(money.NewExchRateFromFloat64("EUR", "OMR", 5.67e0))
	// Output:
	// EUR/JPY 5.67 <nil>
	// EUR/USD 5.67 <nil>
	// EUR/OMR 5.670 <nil>
}

func ExampleNewExchRateFromFloat64_scales() {
	fmt.Println(money.NewExchRateFromFloat64("EUR", "USD", 5.67e-2))
	fmt.Println(money.NewExchRateFromFloat64("EUR", "USD", 5.67e-1))
	fmt.Println(money.NewExchRateFromFloat64("EUR", "USD", 5.67e0))
	fmt.Println(money.NewExchRateFromFloat64("EUR", "USD", 5.67e1))
	fmt.Println(money.NewExchRateFromFloat64("EUR", "USD", 5.67e2))
	// Output:
	// EUR/USD 0.0567 <nil>
	// EUR/USD 0.567 <nil>
	// EUR/USD 5.67 <nil>
	// EUR/USD 56.70 <nil>
	// EUR/USD 567.00 <nil>
}

func ExampleNewExchRateFromInt64_scales() {
	fmt.Println(money.NewExchRateFromInt64("EUR", "USD", 5, 67, 2))
	fmt.Println(money.NewExchRateFromInt64("EUR", "USD", 5, 67, 3))
	fmt.Println(money.NewExchRateFromInt64("EUR", "USD", 5, 67, 4))
	fmt.Println(money.NewExchRateFromInt64("EUR", "USD", 5, 67, 5))
	fmt.Println(money.NewExchRateFromInt64("EUR", "USD", 5, 67, 6))
	// Output:
	// EUR/USD 5.67 <nil>
	// EUR/USD 5.067 <nil>
	// EUR/USD 5.0067 <nil>
	// EUR/USD 5.00067 <nil>
	// EUR/USD 5.000067 <nil>
}

func ExampleNewExchRateFromInt64_currencies() {
	fmt.Println(money.NewExchRateFromInt64("EUR", "JPY", 5, 67, 2))
	fmt.Println(money.NewExchRateFromInt64("EUR", "USD", 5, 67, 2))
	fmt.Println(money.NewExchRateFromInt64("EUR", "OMR", 5, 67, 2))
	// Output:
	// EUR/JPY 5.67 <nil>
	// EUR/USD 5.67 <nil>
	// EUR/OMR 5.670 <nil>
}

func ExampleExchangeRate_Conv_currencies() {
	a := money.MustParseAmount("EUR", "100.00")
	r := money.MustParseExchRate("EUR", "JPY", "160.00")
	q := money.MustParseExchRate("EUR", "USD", "1.2500")
	p := money.MustParseExchRate("EUR", "OMR", "0.42000")
	fmt.Println(r.Conv(a))
	fmt.Println(q.Conv(a))
	fmt.Println(p.Conv(a))
	// Output:
	// JPY 16000.0000 <nil>
	// USD 125.000000 <nil>
	// OMR 42.0000000 <nil>
}

func ExampleExchangeRate_Conv_directions() {
	a := money.MustParseAmount("EUR", "100.00")
	b := money.MustParseAmount("JPY", "16000.00")
	r := money.MustParseExchRate("EUR", "JPY", "160.00")
	fmt.Println(r.Conv(a))
	fmt.Println(r.Conv(b))
	// Output:
	// JPY 16000.0000 <nil>
	// EUR 100.00 <nil>
}

func ExampleExchangeRate_Scale() {
	r := money.MustParseExchRate("USD", "EUR", "0.80")
	q := money.MustParseExchRate("OMR", "USD", "0.38000")
	fmt.Println(r.Scale())
	fmt.Println(q.Scale())
	// Output:
	// 2
	// 5
}

func ExampleExchangeRate_MinScale() {
	r := money.MustParseExchRate("EUR", "USD", "5.6000")
	q := money.MustParseExchRate("EUR", "USD", "5.6700")
	p := money.MustParseExchRate("EUR", "USD", "5.6780")
	fmt.Println(r.MinScale())
	fmt.Println(q.MinScale())
	fmt.Println(p.MinScale())
	// Output:
	// 1
	// 2
	// 3
}

func ExampleExchangeRate_Mul() {
	r := money.MustParseExchRate("EUR", "USD", "5.67")
	d := decimal.MustParse("0.9")
	e := decimal.MustParse("1.0")
	f := decimal.MustParse("1.1")
	fmt.Println(r.Mul(d))
	fmt.Println(r.Mul(e))
	fmt.Println(r.Mul(f))
	// Output:
	// EUR/USD 5.103 <nil>
	// EUR/USD 5.670 <nil>
	// EUR/USD 6.237 <nil>
}

func ExampleExchangeRate_Base() {
	r := money.MustParseExchRate("EUR", "USD", "1.2500")
	fmt.Println(r.Base())
	// Output: EUR
}

func ExampleExchangeRate_Quote() {
	r := money.MustParseExchRate("EUR", "USD", "1.2500")
	fmt.Println(r.Quote())
	// Output: USD
}

func ExampleExchangeRate_Decimal() {
	r := money.MustParseExchRate("EUR", "USD", "1.2500")
	fmt.Println(r.Decimal())
	// Output: 1.2500
}

func ExampleExchangeRate_Float64() {
	r := money.MustParseExchRate("EUR", "USD", "0.10")
	q := money.MustParseExchRate("EUR", "USD", "123.456")
	p := money.MustParseExchRate("EUR", "USD", "1234567890.123456789")
	fmt.Println(r.Float64())
	fmt.Println(q.Float64())
	fmt.Println(p.Float64())
	// Output:
	// 0.1 true
	// 123.456 true
	// 1.2345678901234567e+09 true
}

func ExampleExchangeRate_Int64() {
	r := money.MustParseExchRate("EUR", "USD", "5.678")
	fmt.Println(r.Int64(0))
	fmt.Println(r.Int64(1))
	fmt.Println(r.Int64(2))
	fmt.Println(r.Int64(3))
	fmt.Println(r.Int64(4))
	// Output:
	// 6 0 true
	// 5 7 true
	// 5 68 true
	// 5 678 true
	// 5 6780 true
}

func ExampleExchangeRate_SameCurr() {
	r := money.MustParseExchRate("EUR", "OMR", "0.42000")
	q := money.MustParseExchRate("EUR", "USD", "1.2500")
	p := money.MustParseExchRate("EUR", "USD", "5.6700")
	fmt.Println(r.SameCurr(q))
	fmt.Println(q.SameCurr(p))
	// Output:
	// false
	// true
}

func ExampleExchangeRate_SameScale() {
	r := money.MustParseExchRate("OMR", "EUR", "2.30000")
	q := money.MustParseExchRate("USD", "EUR", "0.9000")
	p := money.MustParseExchRate("SAR", "USD", "0.2700")
	fmt.Println(r.SameScale(q))
	fmt.Println(q.SameScale(p))
	// Output:
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
	// true
	// false
}

func ExampleExchangeRate_Format_currencies() {
	r := money.MustParseExchRate("EUR", "JPY", "5")
	q := money.MustParseExchRate("EUR", "USD", "5")
	p := money.MustParseExchRate("EUR", "OMR", "5")
	fmt.Println("| v             | f     | b   | c   |")
	fmt.Println("| ------------- | ----- | --- | --- |")
	fmt.Printf("| %-13[1]v | %5[1]f | %[1]b | %[1]c |\n", r)
	fmt.Printf("| %-13[1]v | %5[1]f | %[1]b | %[1]c |\n", q)
	fmt.Printf("| %-13[1]v | %5[1]f | %[1]b | %[1]c |\n", p)
	// Output:
	// | v             | f     | b   | c   |
	// | ------------- | ----- | --- | --- |
	// | EUR/JPY 5     |     5 | EUR | JPY |
	// | EUR/USD 5.00  |  5.00 | EUR | USD |
	// | EUR/OMR 5.000 | 5.000 | EUR | OMR |
}

func ExampleExchangeRate_Format_verbs() {
	r := money.MustParseExchRate("USD", "EUR", "1.2500")
	fmt.Printf("%v\n", r)
	fmt.Printf("%[1]f %[1]b-%[1]c\n", r)
	fmt.Printf("%f\n", r)
	fmt.Printf("%b\n", r)
	fmt.Printf("%c\n", r)
	// Output:
	// USD/EUR 1.2500
	// 1.2500 USD-EUR
	// 1.2500
	// USD
	// EUR
}

func ExampleExchangeRate_IsZero() {
	r := money.ExchangeRate{}
	q := money.MustParseExchRate("USD", "EUR", "1.25")
	fmt.Println(r.IsZero())
	fmt.Println(q.IsZero())
	// Output:
	// true
	// false
}

func ExampleExchangeRate_IsOne() {
	r := money.MustParseExchRate("EUR", "USD", "1.00")
	q := money.MustParseExchRate("EUR", "USD", "1.25")
	fmt.Println(r.IsOne())
	fmt.Println(q.IsOne())
	// Output:
	// true
	// false
}

func ExampleExchangeRate_IsPos() {
	r := money.ExchangeRate{}
	q := money.MustParseExchRate("EUR", "USD", "1.25")
	fmt.Println(r.IsPos())
	fmt.Println(q.IsPos())
	// Output:
	// false
	// true
}

func ExampleExchangeRate_Sign() {
	r := money.ExchangeRate{}
	q := money.MustParseExchRate("EUR", "USD", "1.25")
	fmt.Println(r.Sign())
	fmt.Println(q.Sign())
	// Output:
	// 0
	// 1
}

func ExampleExchangeRate_WithinOne() {
	r := money.MustParseExchRate("EUR", "USD", "1.2500")
	q := money.MustParseExchRate("USD", "EUR", "0.8000")
	fmt.Println(r.WithinOne())
	fmt.Println(q.WithinOne())
	// Output:
	// false
	// true
}

func ExampleExchangeRate_String() {
	r := money.MustParseExchRate("EUR", "USD", "1.2500")
	q := money.MustParseExchRate("OMR", "USD", "0.0100")
	fmt.Println(r.String())
	fmt.Println(q.String())
	// Output:
	// EUR/USD 1.2500
	// OMR/USD 0.0100
}

func ExampleExchangeRate_Floor_currencies() {
	r := money.MustParseExchRate("EUR", "JPY", "5.678")
	q := money.MustParseExchRate("EUR", "USD", "5.678")
	p := money.MustParseExchRate("EUR", "OMR", "5.678")
	fmt.Println(r.Floor(0))
	fmt.Println(q.Floor(0))
	fmt.Println(p.Floor(0))
	// Output:
	// EUR/JPY 5 <nil>
	// EUR/USD 5.00 <nil>
	// EUR/OMR 5.000 <nil>
}

func ExampleExchangeRate_Floor_scales() {
	r := money.MustParseExchRate("EUR", "USD", "5.6789")
	fmt.Println(r.Floor(0))
	fmt.Println(r.Floor(1))
	fmt.Println(r.Floor(2))
	fmt.Println(r.Floor(3))
	fmt.Println(r.Floor(4))
	// Output:
	// EUR/USD 5.00 <nil>
	// EUR/USD 5.60 <nil>
	// EUR/USD 5.67 <nil>
	// EUR/USD 5.678 <nil>
	// EUR/USD 5.6789 <nil>
}

func ExampleExchangeRate_Rescale_currencies() {
	r := money.MustParseExchRate("EUR", "JPY", "5.678")
	q := money.MustParseExchRate("EUR", "USD", "5.678")
	p := money.MustParseExchRate("EUR", "OMR", "5.678")
	fmt.Println(r.Rescale(0))
	fmt.Println(q.Rescale(0))
	fmt.Println(p.Rescale(0))
	// Output:
	// EUR/JPY 6 <nil>
	// EUR/USD 6.00 <nil>
	// EUR/OMR 6.000 <nil>
}

func ExampleExchangeRate_Rescale_scales() {
	r := money.MustParseExchRate("EUR", "USD", "5.6789")
	fmt.Println(r.Rescale(0))
	fmt.Println(r.Rescale(1))
	fmt.Println(r.Rescale(2))
	fmt.Println(r.Rescale(3))
	fmt.Println(r.Rescale(4))
	// Output:
	// EUR/USD 6.00 <nil>
	// EUR/USD 5.70 <nil>
	// EUR/USD 5.68 <nil>
	// EUR/USD 5.679 <nil>
	// EUR/USD 5.6789 <nil>
}

func ExampleExchangeRate_Quantize() {
	r := money.MustParseExchRate("EUR", "JPY", "5.678")
	x := money.MustParseExchRate("EUR", "JPY", "1")
	y := money.MustParseExchRate("EUR", "JPY", "0.1")
	z := money.MustParseExchRate("EUR", "JPY", "0.01")
	fmt.Println(r.Quantize(x))
	fmt.Println(r.Quantize(y))
	fmt.Println(r.Quantize(z))
	// Output:
	// EUR/JPY 6 <nil>
	// EUR/JPY 5.7 <nil>
	// EUR/JPY 5.68 <nil>
}

func ExampleExchangeRate_Round_currencies() {
	r := money.MustParseExchRate("EUR", "JPY", "5.678")
	q := money.MustParseExchRate("EUR", "USD", "5.678")
	p := money.MustParseExchRate("EUR", "OMR", "5.678")
	fmt.Println(r.Round(0))
	fmt.Println(q.Round(0))
	fmt.Println(p.Round(0))
	// Output:
	// EUR/JPY 6 <nil>
	// EUR/USD 6.00 <nil>
	// EUR/OMR 6.000 <nil>
}

func ExampleExchangeRate_Round_scales() {
	r := money.MustParseExchRate("EUR", "USD", "5.6789")
	fmt.Println(r.Round(0))
	fmt.Println(r.Round(1))
	fmt.Println(r.Round(2))
	fmt.Println(r.Round(3))
	fmt.Println(r.Round(4))
	// Output:
	// EUR/USD 6.00 <nil>
	// EUR/USD 5.70 <nil>
	// EUR/USD 5.68 <nil>
	// EUR/USD 5.679 <nil>
	// EUR/USD 5.6789 <nil>
}

func ExampleExchangeRate_Trim_currencies() {
	r := money.MustParseExchRate("EUR", "JPY", "5.000")
	q := money.MustParseExchRate("EUR", "USD", "5.000")
	p := money.MustParseExchRate("EUR", "OMR", "5.000")
	fmt.Println(r.Trim(0))
	fmt.Println(q.Trim(0))
	fmt.Println(p.Trim(0))
	// Output:
	// EUR/JPY 5
	// EUR/USD 5.00
	// EUR/OMR 5.000
}

func ExampleExchangeRate_Trim_scales() {
	r := money.MustParseExchRate("EUR", "USD", "5.0000")
	fmt.Println(r.Trim(0))
	fmt.Println(r.Trim(1))
	fmt.Println(r.Trim(2))
	fmt.Println(r.Trim(3))
	fmt.Println(r.Trim(4))
	// Output:
	// EUR/USD 5.00
	// EUR/USD 5.00
	// EUR/USD 5.00
	// EUR/USD 5.000
	// EUR/USD 5.0000
}

func ExampleExchangeRate_Trunc_currencies() {
	r := money.MustParseExchRate("EUR", "JPY", "5.678")
	q := money.MustParseExchRate("EUR", "USD", "5.678")
	p := money.MustParseExchRate("EUR", "OMR", "5.678")
	fmt.Println(r.Trunc(0))
	fmt.Println(q.Trunc(0))
	fmt.Println(p.Trunc(0))
	// Output:
	// EUR/JPY 5 <nil>
	// EUR/USD 5.00 <nil>
	// EUR/OMR 5.000 <nil>
}

func ExampleExchangeRate_Trunc_scales() {
	r := money.MustParseExchRate("EUR", "USD", "5.6789")
	fmt.Println(r.Trunc(0))
	fmt.Println(r.Trunc(1))
	fmt.Println(r.Trunc(2))
	fmt.Println(r.Trunc(3))
	fmt.Println(r.Trunc(4))
	// Output:
	// EUR/USD 5.00 <nil>
	// EUR/USD 5.60 <nil>
	// EUR/USD 5.67 <nil>
	// EUR/USD 5.678 <nil>
	// EUR/USD 5.6789 <nil>
}

func ExampleExchangeRate_Ceil_currencies() {
	r := money.MustParseExchRate("EUR", "JPY", "5.678")
	q := money.MustParseExchRate("EUR", "USD", "5.678")
	p := money.MustParseExchRate("EUR", "OMR", "5.678")
	fmt.Println(r.Ceil(0))
	fmt.Println(q.Ceil(0))
	fmt.Println(p.Ceil(0))
	// Output:
	// EUR/JPY 6 <nil>
	// EUR/USD 6.00 <nil>
	// EUR/OMR 6.000 <nil>
}

func ExampleExchangeRate_Ceil_scales() {
	r := money.MustParseExchRate("EUR", "USD", "5.6789")
	fmt.Println(r.Ceil(0))
	fmt.Println(r.Ceil(1))
	fmt.Println(r.Ceil(2))
	fmt.Println(r.Ceil(3))
	fmt.Println(r.Ceil(4))
	// Output:
	// EUR/USD 6.00 <nil>
	// EUR/USD 5.70 <nil>
	// EUR/USD 5.68 <nil>
	// EUR/USD 5.679 <nil>
	// EUR/USD 5.6789 <nil>
}
