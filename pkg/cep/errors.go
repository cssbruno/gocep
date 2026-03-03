package cep

import "errors"

var (
	// ErrInvalidCEP indicates the input cannot be normalized to 8 digits.
	ErrInvalidCEP = errors.New("invalid cep")
	// ErrNotFound indicates no provider returned a complete CEP address.
	ErrNotFound = errors.New("cep not found")
	// ErrTimeout indicates the search timed out before any provider succeeded.
	ErrTimeout = errors.New("cep lookup timeout")
)
