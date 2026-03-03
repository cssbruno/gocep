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
- Normalized CEP address output (`cidade`, `uf`, `logradouro`, `bairro`)
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
	"fmt"

	"github.com/cssbruno/gocep/pkg/cep"
)

func main() {
	resultJSON, normalized, err := cep.Search("01001-000")
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

## API Reference
- `cep.Search(cep string) (string, models.CEPAddress, error)`:
  runs all configured providers in parallel and returns the first complete address.
- `cep.ValidCEP(models.CEPAddress) bool`:
  validates normalized address completeness (`cidade`, `uf`, `logradouro`, `bairro`).
- `cep.GetOptions()` / `cep.SetOptions(...)`:
  read/update runtime search behavior.
- `cep.SetHTTPClient(client *http.Client)`:
  override outbound HTTP behavior (timeout, proxy, transport).
- `cep.SetCacheProvider(provider)`:
  register a user cache backend.
- `models.GetEndpoints()` / `models.SetEndpoints(...)`:
  read/update provider list safely.
- `util.CheckCEP`, `util.NormalizeCEP`, `util.FormatCEP`:
  input validation and format helpers.

Behavior notes:
- Search accepts `00000000` and `00000-000`.
- Only `https` provider URLs are accepted.
- Cache is used only when `Options.CacheEnabled` is true and a provider is set with `cep.SetCacheProvider(...)`.
- If no provider returns a complete address, `Search` returns `Options.DefaultJSON`.
- `Search` currently returns `nil` error and uses the JSON/address return values for lookup outcome.

Global configuration notes:
- `cep.SetOptions`, `cep.SetHTTPClient`, `cep.SetCacheProvider`, and `models.SetEndpoints` update process-wide state.
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
2. Create an annotated tag:
   ```bash
   git tag -a v1.4.0 -m "release v1.4.0"
   ```
3. Push the tag:
   ```bash
   git push origin v1.4.0
   ```
4. Workflow [`release.yml`](.github/workflows/release.yml) runs tests/race/vet and publishes a GitHub Release with generated notes.

## Credits
Original project and base implementation by **Jeffotoni**:
- GitHub: https://github.com/jeffotoni
- Repository: https://github.com/jeffotoni/gocep
