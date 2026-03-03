package util

import (
	"strings"

	"github.com/cssbruno/gocep/models"
)

// NormalizeAddress trims textual fields and uppercases state code.
func NormalizeAddress(address models.CEPAddress) models.CEPAddress {
	address.Street = strings.TrimSpace(address.Street)
	address.Neighborhood = strings.TrimSpace(address.Neighborhood)
	address.City = strings.TrimSpace(address.City)
	address.StateCode = strings.ToUpper(strings.TrimSpace(address.StateCode))
	return address
}
