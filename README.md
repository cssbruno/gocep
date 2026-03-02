# gocep
[![GoDoc](https://godoc.org/gocep?status.svg)](https://godoc.org/gocep) ![Github Release](https://img.shields.io/github/v/release/cssbruno/gocep?include_prereleases)[![CircleCI](https://dl.circleci.com/status-badge/img/gh/cssbruno/gocep/tree/master.svg?style=svg)](https://dl.circleci.com/status-badge/redirect/gh/cssbruno/gocep/tree/master)[![Go Report](https://goreportcard.com/badge/gocep)](https://goreportcard.com/badge/gocep) [![License](https://img.shields.io/github/license/cssbruno/gocep)](https://img.shields.io/github/license/cssbruno/gocep) ![CircleCI](https://img.shields.io/circleci/build/github/cssbruno/github.com/cssbruno/gocep/master) ![Coveralls](https://img.shields.io/coverallsCoverage/github/cssbruno/gocep)

A fast CEP lookup service and library for Go.
It queries multiple public providers concurrently, returns the first successful response, and caches results in memory.

## Language Standardization
This codebase was standardized to English so contributors from any country can read and maintain it more easily.
English was chosen because it is the most common shared language across open-source tooling, docs, and teams.

Note: some JSON/XML field names from external providers (and response compatibility fields) remain in Portuguese because they are protocol/data-contract values, not internal naming.

## Credits
Original project and base implementation by **Jeffotoni**:
- GitHub: https://github.com/jeffotoni
- Repository: https://github.com/jeffotoni/gocep

## Features
- Concurrent CEP lookup across multiple providers
- In-memory cache to reduce repeated upstream calls
- REST endpoint: `GET /v1/cep/{cep}`
- Library usage via `pkg/cep`
- Deterministic JSON error format

## Current Providers
Configured in [`models/endpoints.go`](models/endpoints.go):
- CDN API CEP
- GitHub raw CEP base
- ViaCEP
- Postmon
- República Virtual
- Correio (SOAP)
- BrasilAPI

## Quick Start (Go)
```bash
git clone https://github.com/cssbruno/gocep.git
cd gocep
go build -o gocep main.go
./gocep
```

Server default address:
- `0.0.0.0:8080`

## Docker
```bash
docker run --name gocep --rm -p 8080:8080 cssbruno/gocep:latest
```

Or use:
```bash
make compose
```

## API Usage
### Request
```bash
curl -i -X GET http://localhost:8080/v1/cep/08226021
```

### Success Response (`200`)
```json
{
  "cidade": "São Paulo",
  "uf": "SP",
  "logradouro": "Rua Esperança",
  "bairro": "Cidade Antônio Estevão de Carvalho"
}
```

### No Content (`204`)
Returned when CEP format is valid but no provider returned a complete address.

### Error Response Pattern
All handler-level errors follow the same JSON shape:
```json
{
  "error": {
    "code": "invalid_cep",
    "message": "cep must contain exactly 8 digits"
  }
}
```

Examples of error codes:
- `method_not_allowed` (`405`)
- `invalid_endpoint` (`302`)
- `invalid_cep` (`400`)
- `search_error` (`400`)
- `not_found` (`404`)

## Library Usage
```go
package main

import (
	"fmt"

	"github.com/cssbruno/gocep/pkg/cep"
)

func main() {
	result, normalized, err := cep.Search("08226021")
	fmt.Println("error:", err)
	fmt.Println("json:", result)
	fmt.Println("struct:", normalized)
}
```

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
- `INSECURE_SKIP_VERIFY` (default: `false`)

## Examples
Check language examples under [`examples/`](examples/):
- `nodejs`
- `python`
- `php`
- `javascript`
- `go` (lib/client/server)
- `rust`
- `c`
- `c++`

## Development
```bash
go test ./...
go test -race ./...
go vet ./...
```
