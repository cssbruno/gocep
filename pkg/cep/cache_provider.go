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

func SetCacheProvider(provider CacheProvider) {
	gocache.SetProvider(provider)
}
