# Changelog

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
