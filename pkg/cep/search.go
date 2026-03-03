package cep

import (
	"context"

	"github.com/cssbruno/gocep/models"
)

// Result represents one provider response payload.
type Result struct {
	Body    []byte
	Address models.CEPAddress
}

type cachedResult struct {
	JSON    string
	Address models.CEPAddress
}

// Search looks up a CEP using the package default client.
func Search(cep string) (jsonCep string, address models.CEPAddress, err error) {
	return defaultClient.Search(cep)
}

// SearchContext looks up a CEP using the package default client and context.
func SearchContext(ctx context.Context, cep string) (jsonCep string, address models.CEPAddress, err error) {
	return defaultClient.SearchContext(ctx, cep)
}

// ValidCEP reports whether an address has all required normalized fields.
func ValidCEP(address models.CEPAddress) bool {
	return isCompleteAddress(address)
}
