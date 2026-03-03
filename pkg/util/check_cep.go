package util

import (
	"errors"
	"strings"
)

// ErrInvalidCEP is returned when a CEP cannot be normalized to 8 digits.
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

// NormalizeCEP removes common separators and validates the final CEP.
// Accepted separators are spaces, hyphen, dot, and slash.
func NormalizeCEP(cep string) (string, error) {
	cep = strings.TrimSpace(cep)
	if cep == "" {
		return "", ErrInvalidCEP
	}

	// Fast path for already-normalized CEP.
	if err := CheckCEP(cep); err == nil {
		return cep, nil
	}

	var digits [8]byte
	count := 0
	for i := 0; i < len(cep); i++ {
		switch c := cep[i]; {
		case c >= '0' && c <= '9':
			if count < len(digits) {
				digits[count] = c
				count++
				continue
			}
			return "", ErrInvalidCEP
		case c == '-' || c == '.' || c == '/' || c == ' ' || c == '\t' || c == '\n' || c == '\r':
			continue
		default:
			return "", ErrInvalidCEP
		}
	}

	if count != len(digits) {
		return "", ErrInvalidCEP
	}

	return string(digits[:]), nil
}

// FormatCEP returns a CEP in the standard "00000-000" format.
func FormatCEP(cep string) (string, error) {
	normalized, err := NormalizeCEP(cep)
	if err != nil {
		return "", err
	}
	return normalized[:5] + "-" + normalized[5:], nil
}
