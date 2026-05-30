package cep

import "time"

const (
	defaultJSON            = `{"cep":"","cidade":"","uf":"","logradouro":"","bairro":""}`
	defaultCacheTTL        = 48 * time.Hour
	defaultSearchTimeout   = 15 * time.Second
	defaultMaxProviderBody = 1 << 20
)

// Options controls runtime behavior for CEP searches.
type Options struct {
	// DefaultJSON is returned when no provider yields a complete address.
	DefaultJSON string
	// CacheEnabled toggles read/write operations to the configured cache provider.
	CacheEnabled bool
	// CacheTTL is used when storing search results in cache.
	CacheTTL time.Duration
	// SearchTimeout limits total search time across all providers.
	// A zero or negative value uses the package default.
	SearchTimeout time.Duration
	// MaxProviderBody is the maximum response body size accepted per provider.
	MaxProviderBody int64
}

func defaultOptions() Options {
	return Options{
		DefaultJSON:     defaultJSON,
		CacheEnabled:    true,
		CacheTTL:        defaultCacheTTL,
		SearchTimeout:   defaultSearchTimeout,
		MaxProviderBody: defaultMaxProviderBody,
	}
}

// GetOptions returns a copy of the current CEP options.
func GetOptions() Options {
	return defaultClient.Options()
}

// SetOptions replaces CEP package options for the current process.
// Empty or invalid fields are normalized to safe defaults.
// Applications should configure options during startup.
func SetOptions(next Options) {
	defaultClient.SetOptions(next)
}

func normalizeOptions(in Options) Options {
	out := in
	if out.DefaultJSON == "" {
		out.DefaultJSON = defaultJSON
	}
	if out.CacheTTL <= 0 {
		out.CacheTTL = defaultCacheTTL
	}
	if out.SearchTimeout <= 0 {
		out.SearchTimeout = defaultSearchTimeout
	}
	if out.MaxProviderBody <= 0 {
		out.MaxProviderBody = defaultMaxProviderBody
	}
	return out
}
