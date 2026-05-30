# Changelog

All notable changes to this project are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.0.0] - 2026-05-30

### Added
- `cep.Client` API with isolated per-client configuration.
- `cep.SearchContext(ctx, cep)` with context-based cancellation/deadline support.
- Typed lookup errors: `cep.ErrInvalidCEP`, `cep.ErrNotFound`, `cep.ErrTimeout`.
- Provider policy controls: ordered fallback, preferred sources, disabled sources, per-source timeouts.
- Observability hooks for cache/provider events.
- CI/release quality gates: `staticcheck`, `golangci-lint`, `govulncheck`.
- CI example build validation via `go test ./examples/...`.

### Changed
- `cep.Search` now returns typed errors for invalid CEP, timeout, and not-found outcomes.
- Provider request errors are now surfaced in provider hook events.
- Normalized address payload now includes `cep` (`cep`, `cidade`, `uf`, `logradouro`, `bairro`).
- `models.Endpoints` is no longer exported; use `models.GetEndpoints()` and `models.SetEndpoints(...)`.
- `cep.NewClient()` no longer inherits the package global cache provider unless `cep.WithCacheProvider(...)` is supplied.
- Module imports now use the Go v2 module path: `github.com/cssbruno/gocep/v2`.

### Fixed
- `cep.SearchContext` no longer shares cancelable caller contexts through singleflight.
- Parallel provider lookup no longer races a successful result against cancellation.
- Republica Virtual responses now combine `tipo_logradouro` and `logradouro`.
- Provider redirects that downgrade from HTTPS are rejected.

## [1.0.1] - 2026-03-02

### Added
- Professionalized package/API documentation comments.
- Package-level docs for `models`, `pkg/cep`, `pkg/util`, and `service/gocache`.

### Changed
- README improved with clearer API semantics and usage examples.
