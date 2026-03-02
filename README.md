# gocep
[![Go Reference](https://pkg.go.dev/badge/github.com/cssbruno/gocep.svg)](https://pkg.go.dev/github.com/cssbruno/gocep) ![GitHub Release](https://img.shields.io/github/v/release/cssbruno/gocep?include_prereleases) [![Go Report Card](https://goreportcard.com/badge/github.com/cssbruno/gocep)](https://goreportcard.com/report/github.com/cssbruno/gocep) [![License](https://img.shields.io/github/license/cssbruno/gocep)](https://github.com/cssbruno/gocep/blob/master/LICENSE)

Fast CEP lookup API and Go library.
It queries multiple public providers in parallel, returns the first successful address, and caches results in memory.

## What Is a CEP?
`CEP` means `Codigo de Enderecamento Postal` in Brazil.
It is the Brazilian postal code, similar to a ZIP code in the US.

For non-English readers:
- `CEP` = postal code in Brazil.
- Usually written as `00000-000` (for example `01001-000`) or as 8 digits (`01001000`). This API accepts both formats.

## Why This Project
- Concurrent lookup across multiple providers
- In-memory cache for repeated CEP queries
- REST endpoint: `GET /v1/cep/{cep}`
- Reusable Go package: [`pkg/cep`](pkg/cep)
- Deterministic JSON error contract

## Language Standardization
The codebase was standardized to English so contributors from different countries can read and maintain it more easily.

Some provider field names remain in Portuguese in JSON/XML mappings because they are external protocol contracts.

## Credits
Original project and base implementation by **Jeffotoni**:
- GitHub: https://github.com/jeffotoni
- Repository: https://github.com/jeffotoni/gocep

## Quick Start
```bash
git clone https://github.com/cssbruno/gocep.git
cd gocep
go run .
```

Default server address: `0.0.0.0:8080`

### Test the API
```bash
curl -i http://localhost:8080/v1/cep/01001000
```

## Docker
```bash
docker run --name gocep --rm -p 8080:8080 cssbruno/gocep:latest
```

Or:
```bash
make compose
```

## HTTP API
### Main route
- `GET /v1/cep/{cep}`

### Success response (`200`)
```json
{
  "cidade": "Sao Paulo",
  "uf": "SP",
  "logradouro": "Praca da Se",
  "bairro": "Se"
}
```

### No content (`204`)
Returned when the CEP format is valid, but no provider returned a complete address.

### Error response format
```json
{
  "error": {
    "code": "invalid_cep",
    "message": "cep must be in 00000000 or 00000-000 format"
  }
}
```

### Complete list of error codes
| HTTP status | Error code | When it happens |
| --- | --- | --- |
| `302` | `invalid_endpoint` | CEP path is malformed (example: extra slash `/v1/cep/01001000/`) |
| `400` | `invalid_cep` | CEP is not valid in `00000000` or `00000-000` format |
| `400` | `search_error` | Internal CEP search call returned an error |
| `404` | `not_found` | Route not mapped (for example `/any-other-path`) |
| `405` | `method_not_allowed` | Route exists but method is not `GET` |

Notes:
- `200` and `204` are successful outcomes and do not include an `error.code`.
- `302 invalid_endpoint` is kept for backward compatibility with current behavior.

## Go Library Usage
```go
package main

import (
	"fmt"

	"github.com/cssbruno/gocep/pkg/cep"
)

func main() {
	resultJSON, normalized, err := cep.Search("01001000")
	fmt.Println("error:", err)
	fmt.Println("json:", resultJSON)
	fmt.Println("address:", normalized)
}
```

## Current Providers
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

## Configuration
Main environment variables from [`config/config.go`](config/config.go):
- `PORT` (default: `0.0.0.0:8080`)
- `CACHE_ENABLE` (default: `true`)
- `TTL_CACHE` seconds (default: `172800`)
- `TIMEOUT_SEARCH_CEP` seconds (default: `15`)
- `HTTP_CLIENT_MAXIDLECONNS`
- `HTTP_CLIENT_MAXIDLECONNSPERHOST`
- `IDLE_CONN_TIMEOUT`
- `TIMEOUT`
- `MAX_PROVIDER_BODY` bytes (default: `1048576`)
- `SERVER_READ_HEADER_TIMEOUT` seconds (default: `5`)
- `SERVER_READ_TIMEOUT` seconds (default: `10`)
- `SERVER_WRITE_TIMEOUT` seconds (default: `15`)
- `SERVER_IDLE_TIMEOUT` seconds (default: `120`)
- `SERVER_MAX_HEADER_BYTES` bytes (default: `1048576`)
- `INSECURE_SKIP_VERIFY` (default: `false`)
- `CACHE_BACKEND` (`memory` or `redis`, default: `memory`)
- `REDIS_ADDR` (default: `127.0.0.1:6379`)
- `REDIS_USER`
- `REDIS_PASSWORD`
- `REDIS_DB` (default: `0`)
- `REDIS_PREFIX` (default: `gocep:`)

## Examples
Examples in [`examples/`](examples/):
- `go` (lib/client/server)
- `nodejs`
- `javascript`
- `python`
- `php`
- `rust`
- `c`
- `c++`

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
4. Workflow [`release.yml`](.github/workflows/release.yml) runs tests, race tests, vet, builds binaries, and publishes a GitHub Release with generated notes.
