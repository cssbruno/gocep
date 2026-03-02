package util

import (
	"errors"
)

var ErrInvalidCEP = errors.New("invalid cep")

// CheckCEP validates whether the provided CEP has exactly 8 numeric digits.
func CheckCEP(cep string) error {
	if len(cep) != 8 {
		return ErrInvalidCEP
	}
	for i := 0; i < len(cep); i++ {
		if cep[i] < '0' || cep[i] > '9' {
			return ErrInvalidCEP
		}
	}
	return nil
}

// Deprecated: use CheckCEP.
func CheckCep(cep string) error {
	return CheckCEP(cep)
}
