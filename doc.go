/*
Package money implements immutable monetary values in various currencies.
It leverages the capabilities of [decimal.Decimal] for handling decimal floating-point
numbers and combines them with a [Currency] to represent different currencies.

# Features

  - Immutable monetary values, ensuring safe usage across multiple goroutines
  - Support for various currencies and their corresponding scales
  - Arithmetic and comparison operations between monetary values
  - Conversion of monetary values using exchange rates

# Supported Ranges

The range of monetary values supported refer to the Supported Ranges section
of the [decimal] package description.

# Operations

The package provides various arithmetic and comparison operations for
monetary values, such as Add, Sub, Mul, FMA, Quo, and Rat.
It also supports splitting an amount into equal parts and converting between
different currencies using exchange rates.

# Rounding

The package allows for rounding monetary values using several methods,
including Ceil, Floor, Trunc, and Round.
*/
package money
