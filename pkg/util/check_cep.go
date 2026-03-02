package util

import (
	"errors"
	"strings"
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

// NormalizeCEP removes common separators and validates the final CEP.
// Accepted separators are spaces, hyphen, dot, and slash.
func NormalizeCEP(cep string) (string, error) {
	cep = strings.TrimSpace(cep)
	if cep == "" {
		return "", ErrInvalidCEP
	}

	digits := make([]byte, 0, 8)
	for i := 0; i < len(cep); i++ {
		switch c := cep[i]; {
		case c >= '0' && c <= '9':
			digits = append(digits, c)
		case c == '-' || c == '.' || c == '/' || c == ' ' || c == '\t' || c == '\n' || c == '\r':
			continue
		default:
			return "", ErrInvalidCEP
		}
	}

	normalized := string(digits)
	if err := CheckCEP(normalized); err != nil {
		return "", err
	}
	return normalized, nil
}

// FormatCEP returns a CEP in the standard "00000-000" format.
func FormatCEP(cep string) (string, error) {
	normalized, err := NormalizeCEP(cep)
	if err != nil {
		return "", err
	}
	return normalized[:5] + "-" + normalized[5:], nil
}

// Deprecated: use CheckCEP.
func CheckCep(cep string) error {
	return CheckCEP(cep)
}
