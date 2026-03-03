# Changelog

All notable changes to this project are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- `cep.Client` API with isolated per-client configuration.
- `cep.SearchContext(ctx, cep)` with context-based cancellation/deadline support.
- Typed lookup errors: `cep.ErrInvalidCEP`, `cep.ErrNotFound`, `cep.ErrTimeout`.
- Provider policy controls: ordered fallback, preferred sources, disabled sources, per-source timeouts.
- Observability hooks for cache/provider events.
- CI/release quality gates: `staticcheck`, `golangci-lint`, `govulncheck`.
- CI example validation via `go run ./examples/go/lib` and `go run ./examples/go/client`.

### Changed
- `cep.Search` now returns typed errors for invalid CEP, timeout, and not-found outcomes.
- Provider request errors are now surfaced in provider hook events.

## [1.0.1] - 2026-03-02

### Added
- Professionalized package/API documentation comments.
- Package-level docs for `models`, `pkg/cep`, `pkg/util`, and `service/gocache`.

### Changed
- README improved with clearer API semantics and usage examples.
