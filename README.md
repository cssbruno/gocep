# gocep
[![CI](https://github.com/cssbruno/gocep/actions/workflows/ci.yml/badge.svg)](https://github.com/cssbruno/gocep/actions/workflows/ci.yml) [![Go Reference](https://pkg.go.dev/badge/github.com/cssbruno/gocep.svg)](https://pkg.go.dev/github.com/cssbruno/gocep) ![GitHub Release](https://img.shields.io/github/v/release/cssbruno/gocep?include_prereleases) [![Go Report Card](https://goreportcard.com/badge/github.com/cssbruno/gocep)](https://goreportcard.com/report/github.com/cssbruno/gocep) [![License](https://img.shields.io/github/license/cssbruno/gocep)](https://github.com/cssbruno/gocep/blob/master/LICENSE)

Fast CEP lookup library for Go.
It queries multiple providers in parallel, returns the first successful address, and supports optional user-provided caching for repeated lookups.

## What Is a CEP?
`CEP` means `Código de Endereçamento Postal` in Brazil.
It is the Brazilian postal code, similar to a ZIP code in the US.
The library accepts both formats:
- `00000-000` (example: `01001-000`)
- `00000000` (example: `01001000`)

## Features
- Parallel provider lookup with first-success response
- Optional ordered fallback strategy with provider policy controls
- Normalized CEP address output (`cidade`, `uf`, `logradouro`, `bairro`)
- Isolated `cep.Client` API (per-client options, HTTP client, cache, endpoints, hooks)
- Pluggable cache provider (user implementation)
- CEP validation and normalization utilities
- Stable API for direct library integration

## Install
```bash
go get github.com/cssbruno/gocep@latest
```

## Basic Usage
```go
package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/cssbruno/gocep/pkg/cep"
)

func main() {
	resultJSON, normalized, err := cep.SearchContext(context.Background(), "01001-000")
	switch {
	case errors.Is(err, cep.ErrInvalidCEP):
		fmt.Println("invalid cep")
	case errors.Is(err, cep.ErrTimeout):
		fmt.Println("lookup timed out")
	case errors.Is(err, cep.ErrNotFound):
		fmt.Println("cep not found")
	}
	fmt.Println("error:", err)
	fmt.Println("json:", resultJSON)
	fmt.Println("address:", normalized)
}
```

## CEP Utility Helpers
```go
package main

import (
	"fmt"

	"github.com/cssbruno/gocep/pkg/util"
)

func main() {
	normalized, err := util.NormalizeCEP("01001-000")
	fmt.Println(normalized, err) // 01001000 <nil>

	formatted, err := util.FormatCEP("01001000")
	fmt.Println(formatted, err) // 01001-000 <nil>
}
```

## Providers
Configured in [`models/endpoints.go`](models/endpoints.go):
- CDN API CEP
- GitHub raw CEP base
- ViaCEP
- Postmon
- Republica Virtual
- Correio (SOAP)
- BrasilAPI
- OpenCEP
- AwesomeAPI CEP

To override providers at runtime, prefer `models.SetEndpoints(...)`.

## Configuration
The library is configured in code through [`pkg/cep/options.go`](pkg/cep/options.go).

Example:
```go
package main

import (
	"fmt"
	"time"

	"github.com/cssbruno/gocep/pkg/cep"
)

type myCacheProvider struct{}

func (myCacheProvider) SetAnyTTL(key string, value any, ttl time.Duration) bool { return true }
func (myCacheProvider) GetAny(key string) (any, bool) { return nil, false }

func main() {
	opts := cep.GetOptions()
	opts.CacheEnabled = true
	opts.SearchTimeout = 10 * time.Second
	opts.CacheTTL = 24 * time.Hour
	cep.SetOptions(opts)

	// Optional: provide your own cache implementation.
	cep.SetCacheProvider(myCacheProvider{})

	result, address, err := cep.Search("01001-000")
	fmt.Println(result, address, err)
}
```

Main options:
- `cep.Options.CacheEnabled`
- `cep.Options.CacheTTL`
- `cep.Options.SearchTimeout`
- `cep.Options.MaxProviderBody`

## Advanced Client API
Use `cep.NewClient(...)` when you need isolated configuration instead of global package state.

```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/cssbruno/gocep/pkg/cep"
)

type myCacheProvider struct{}

func (myCacheProvider) SetAnyTTL(key string, value any, ttl time.Duration) bool { return true }
func (myCacheProvider) GetAny(key string) (any, bool) { return nil, false }

func main() {
	opts := cep.Options{
		DefaultJSON:     `{"cidade":"","uf":"","logradouro":"","bairro":""}`,
		CacheEnabled:    true,
		CacheTTL:        24 * time.Hour,
		SearchTimeout:   5 * time.Second,
		MaxProviderBody: 1 << 20,
	}

	httpClient := &http.Client{Timeout: 6 * time.Second}

client := cep.NewClient(
	cep.WithOptions(opts),
	cep.WithHTTPClient(httpClient),
	cep.WithCacheProvider(myCacheProvider{}),
	cep.WithProviderPolicy(cep.ProviderPolicy{
		Strategy:         cep.SearchStrategyOrderedFallback,
		PreferredSources: []string{"brasilapi", "viacep"},
	}),
)

	result, address, err := client.SearchContext(context.Background(), "01001-000")
	fmt.Println(result, address, err)
}
```

## API Reference
- `cep.Search(cep string) (string, models.CEPAddress, error)`:
  runs lookup with the package default client.
- `cep.SearchContext(ctx, cep string) (string, models.CEPAddress, error)`:
  lookup with caller-controlled cancellation/deadline.
- `cep.NewClient(...)` and `client.SearchContext(...)`:
  isolated client configuration and lookups.
- `cep.ValidCEP(models.CEPAddress) bool`:
  validates normalized address completeness (`cidade`, `uf`, `logradouro`, `bairro`).
- `cep.GetOptions()` / `cep.SetOptions(...)`:
  read/update options for the package default client.
- `cep.SetHTTPClient(client *http.Client)`:
  override HTTP client for the package default client.
- `cep.SetCacheProvider(provider)`:
  register cache backend for the package default client.
- `cep.SetProviderPolicy(policy)`:
  configure ordered fallback, preferred sources, disabled sources, and per-source timeouts.
- `cep.SetHooks(hooks)`:
  attach cache/provider event callbacks.
- `models.GetEndpoints()` / `models.SetEndpoints(...)`:
  read/update global provider list safely.
- `util.CheckCEP`, `util.NormalizeCEP`, `util.FormatCEP`:
  input validation and format helpers.

Behavior notes:
- Search accepts `00000000` and `00000-000`.
- Only `https` provider URLs are accepted.
- Cache is used only when `Options.CacheEnabled` is true and a provider is set with `cep.SetCacheProvider(...)`.
- If no provider returns a complete address, lookup returns `Options.DefaultJSON` and `cep.ErrNotFound`.
- Invalid CEP returns `cep.ErrInvalidCEP`.
- Timeout returns `cep.ErrTimeout`.
- Provider-specific timeout failures also return `cep.ErrTimeout` when no provider succeeds.

Global configuration notes:
- `cep.SetOptions`, `cep.SetHTTPClient`, `cep.SetCacheProvider`, `cep.SetProviderPolicy`, `cep.SetHooks`, and `models.SetEndpoints` update process-wide state.
- Configure them during startup before handling concurrent calls to `cep.Search`.

## Examples
Examples in [`examples/`](examples/):
- `go/lib`: basic lookup
- `go/client`: advanced configuration (options + custom cache provider)

## Development
```bash
go test ./...
go test -race ./...
go vet ./...
make test
```

## Versioning And Releases
Releases are published from Git tags using GitHub Actions.

Tag format:
- `vX.Y.Z` (example: `v1.4.0`)

Release flow:
1. Commit and push your changes to `master`/`main`.
2. Update [`CHANGELOG.md`](CHANGELOG.md):
   - Move relevant items from `Unreleased` to a new version section.
   - Keep entries grouped under `Added`, `Changed`, `Fixed`, `Removed` when applicable.
3. Create an annotated tag:
   ```bash
   git tag -a v1.4.0 -m "release v1.4.0"
   ```
4. Push the tag:
   ```bash
   git push origin v1.4.0
   ```
5. Workflow [`release.yml`](.github/workflows/release.yml) runs tests/race/vet/staticcheck/golangci-lint/govulncheck, validates examples, and publishes a GitHub Release with generated notes.

## Credits
Original project and base implementation by **Jeffotoni**:
- GitHub: https://github.com/jeffotoni
- Repository: https://github.com/jeffotoni/gocep
