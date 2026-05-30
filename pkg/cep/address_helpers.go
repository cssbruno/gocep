package cep

import (
	"github.com/cssbruno/gocep/v2/models"
	"github.com/cssbruno/gocep/v2/pkg/util"
)

func formattedCEPOrRaw(cep string) string {
	formatted, err := util.FormatCEP(cep)
	if err == nil {
		return formatted
	}
	return cep
}

func withCEP(address models.CEPAddress, cep string) models.CEPAddress {
	if address.CEP != "" {
		return util.NormalizeAddress(address)
	}
	address.CEP = formattedCEPOrRaw(cep)
	return util.NormalizeAddress(address)
}
