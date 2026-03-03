package gocache

import (
	"sync"
	"time"
)

// Provider allows applications to inject their own cache implementation.
type Provider interface {
	SetAnyTTL(key string, value any, ttl time.Duration) bool
	GetAny(key string) (any, bool)
}

var (
	providerMu sync.RWMutex
	provider   Provider
)

// SetProvider sets the active cache backend implementation.
func SetProvider(p Provider) {
	providerMu.Lock()
	provider = p
	providerMu.Unlock()
}

// SetTTL stores a string value in cache with a TTL.
func SetTTL(key, value string, ttl time.Duration) bool {
	if key == "" || value == "" {
		return false
	}
	return SetAnyTTL(key, value, ttl)
}

// SetAnyTTL stores any value in cache with a TTL.
func SetAnyTTL(key string, value any, ttl time.Duration) bool {
	if key == "" || value == nil {
		return false
	}

	providerMu.RLock()
	p := provider
	providerMu.RUnlock()
	if p == nil {
		return false
	}
	return p.SetAnyTTL(key, value, ttl)
}

// GetAny returns a cached value and whether it was found.
func GetAny(key string) (any, bool) {
	if key == "" {
		return nil, false
	}

	providerMu.RLock()
	p := provider
	providerMu.RUnlock()
	if p == nil {
		return nil, false
	}
	return p.GetAny(key)
}

// Get returns a cached value as string when possible.
func Get(key string) string {
	value, found := GetAny(key)
	if !found {
		return ""
	}
	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		return ""
	}
}

func resetCacheForTests() {
	SetProvider(nil)
}
