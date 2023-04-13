/*
Package money implements monetary values in various currencies.
It leverages the [decimal] package's capabilities for handling decimal floating-point
numbers and combines it with a [Currency] struct for representing different currencies.

# Features

  - Immutable monetary values, ensuring safe usage across multiple goroutines
  - Support for various currencies and their corresponding scales
  - Arithmetic and comparison operations between monetary values
  - Rounding methods in accordance with the currency's scale
  - Conversion of monetary values using exchange rates

# Representation

The Money package consists of two main structs: Amount and Currency.
An Amount represents a monetary value and consists of a Currency and
a decimal.Decimal value.
The Currency struct represents a currency and is implemented as
an integer index into an in-memory array containing information
such as code, symbol, and scale.

# Supported Ranges

The range of monetary values supported depends on the Currency's scale and
the Decimal's coefficient.
For more information, refer to the Supported Ranges section of the [decimal]
package description.

# Operations

The Money package provides various arithmetic and comparison operations for
monetary values, such as Add, Sub, Mul, FMA, and Rat.
It also supports splitting an Amount into equal parts and converting between
different currencies using exchange rates.

# Rounding

The Money package allows for rounding monetary values using several methods,
including Ceil, Floor, Trunc, and Round.
Rounding is performed in accordance with the scale of the Currency.

# Errors

Errors may occur during the parsing of Amount and Currency values, as well
as during arithmetic operations when certain conditions are not met
(e.g., division by zero, coefficient overflow).
The package returns errors or panics, depending on the situation.
*/
package money
