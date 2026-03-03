package cep

import (
	"sync"
	"time"
)

type Options struct {
	DefaultJSON     string
	CacheEnabled    bool
	CacheTTL        time.Duration
	SearchTimeout   time.Duration
	MaxProviderBody int64
}

var (
	optionsMu sync.RWMutex
	options   = Options{
		DefaultJSON:     `{"cidade":"","uf":"","logradouro":"","bairro":""}`,
		CacheEnabled:    true,
		CacheTTL:        48 * time.Hour,
		SearchTimeout:   15 * time.Second,
		MaxProviderBody: 1 << 20,
	}
)

func GetOptions() Options {
	optionsMu.RLock()
	defer optionsMu.RUnlock()
	return options
}

func SetOptions(next Options) {
	optionsMu.Lock()
	defer optionsMu.Unlock()
	options = normalizeOptions(next)
}

func normalizeOptions(in Options) Options {
	out := in
	if out.DefaultJSON == "" {
		out.DefaultJSON = `{"cidade":"","uf":"","logradouro":"","bairro":""}`
	}
	if out.CacheTTL <= 0 {
		out.CacheTTL = 48 * time.Hour
	}
	if out.SearchTimeout < 0 {
		out.SearchTimeout = 15 * time.Second
	}
	if out.MaxProviderBody <= 0 {
		out.MaxProviderBody = 1 << 20
	}
	return out
}
