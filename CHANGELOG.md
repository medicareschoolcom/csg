# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.1] - 2026-01-13

### Added
- Go client library for CSG Actuarial Medicare Supplement API
- Type-safe client for Medicare Supplement quote retrieval
- Support for all standard Medicare Supplement plans (F, G, N)
- Automatic token management with caching and retry logic
- Companies listing endpoint (unauthenticated)
- Context support for all API methods
- Thread-safe token caching using double-checked locking
- Custom base URL configuration option
- Comprehensive README with usage examples
- Unit and integration test suite

### API Methods
- `ListQuotes` - Retrieve Medicare Supplement quotes
- `ListCompanies` - List available insurance companies

### Types
- `Client` - Main API client
- `Quote` - Quote response structure
- `QuoteRate` - Premium rate structure
- `Company` - Insurance company information
- `TokenProvider` - Interface for authentication

### Constants
- Gender constants: `Male`, `Female`
- Boolean constants: `True`, `False`
- Plan constants: `PlanF`, `PlanG`, `PlanN`

[Unreleased]: https://github.com/medicareschoolcom/csg/compare/v0.0.1...HEAD
[0.0.1]: https://github.com/medicareschoolcom/csg/releases/tag/v0.0.1
