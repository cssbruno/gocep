// Package cep provides high-level CEP search primitives.
//
// The package queries multiple providers concurrently, returns the first
// successful normalized address, and supports user-provided cache backends.
//
// For isolated configuration, prefer NewClient and SearchContext.
package cep
