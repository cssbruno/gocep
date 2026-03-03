package cep

import (
	"time"

	"github.com/cssbruno/gocep/service/gocache"
)

// CacheProvider is an application-provided cache backend implementation.
type CacheProvider interface {
	SetAnyTTL(key string, value any, ttl time.Duration) bool
	GetAny(key string) (any, bool)
}

// SetCacheProvider registers a cache backend for Search cache reads/writes.
// Passing nil disables cache usage at the provider layer.
// The setting is global for the current process.
func SetCacheProvider(provider CacheProvider) {
	gocache.SetProvider(provider)
}
