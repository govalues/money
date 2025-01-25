# Changelog

## [0.2.4] - 2025-01-26

### Added

- Implemented `Amount.AddMul`, `Amount.AddQuo`, `Amount.SubMul`, `Amount.SubQuo`, `Amount.Equal`, `Amount.Less`.

- Implemented `Currency.AppendBinary`, `Currency.AppendText`, `Currency.UnmarshalJSON`, `Currency.MarshalJSON`, `Currency.UnmarshalBSONValue`, `Currency.MarshalBSONValue`.

### Changed

- `ExchangeRate.Conv` supports both direct and reverse conversions.
- Deprecated `Amount.FMA` and `ExchangeRate.Inv` methods.

## [0.2.3] - 2024-07-26

### Changed

- Allowed rounding to the scale smaller than the currency scale.

## [0.2.2] - 2023-12-18

### Changed

- Improved table formatting in documentation.

## [0.2.1] - 2023-12-18

### Changed

- Improved examples and documentation.
- Improved test coverage.

## [0.2.0] - 2023-12-12

### Added

- Implemented constructors:
  - `NewAmountFromInt64`,
  - `NewAmountFromFloat64`,
  - `NewAmountFromMinorUnits`,
  - `NewExchRareFromInt64`,
  - `NewExchRareFromFloat64`.
- Implemented methods:
  - `Amount.Decimal`,
  - `Amount.MinScale`,
  - `Amount.CmbAbs`,
  - `Amount.SubAbs`,
  - `Amount.Clamp`,
  - `ExchangeRate.Float64`,
  - `ExchangeRate.Int64`,
  - `ExchangeRate.Decimal`,
  - `ExchangeRate.Ceil`,
  - `ExchangeRate.Floor`,
  - `ExchangeRate.Trunc`,
  - `ExchangeRate.Trim`.
  - `ExchangeRate.IsPos`,
  - `ExchangeRate.Sign`,
  - `ExchangeRate.MinScale`,
  - `ExchangeRate.Quantize`,
- Implemented `NullCurrency` type.

### Changed

- Renamed `NewAmount` contructor to `NewAmountFromDecimal`.
- Renamed `NewExchRate` constructor to `NewExchRateFromDecimal`.
- Changed `ExchangeRate.Round`, now it returns an error.
- Changed `ExchangeRate.Format`, now `%c` returns quore currency, not a currency pair.

### Removed

- Removed methods:
  - `Amount.Prec`,
  - `Amount.Coef`,
  - `ExchangeRate.RoundToCurr`,
  - `ExchangeRate.Prec`,
  - `ExchangeRate.SameScaleAsCurr`.

## [0.1.3] - 2023-08-21

### Added

- Implemented `Currency.Scan` and `Currency.Value`.

### Changed

- `Amount.CopySign` treats 0 as a positive.
- Enabled `gocyclo`, `gosec`, `godot`, and `stylecheck` linters.

## [0.1.2] - 2023-08-15

### Added

- Implemented `Amount.QuoRem` method.

## [0.1.1] - 2023-08-04

### Changed

- Implemented `scale` argument for `Amount.Int64` method.

## [0.1.0] - 2023-06-04

### Changed

- All methods return error instead of panicing.
- Renamed `Amount.Round` to `Amount.Rescale`.
- Renamed `ExchangeRate.Round` to `ExchangeRate.Rescale`.

## [0.0.3] - 2023-04-22

### Changed

- Improved documentation.

## [0.0.2] - 2023-04-15

### Added

- Implemented `Amount.Int64` method.
- Implemented `Amount.Float64` method.

### Changed

- Reviewed and improved documentation.

## [0.0.1] - 2023-04-13

### Added

- Initial version
