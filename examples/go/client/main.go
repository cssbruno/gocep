package main

import (
	"fmt"
	"time"

	"github.com/cssbruno/gocep/pkg/cep"
)

type inMemoryCache struct {
	items map[string]any
}

func (m *inMemoryCache) SetAnyTTL(key string, value any, _ time.Duration) bool {
	if key == "" || value == nil {
		return false
	}
	if m.items == nil {
		m.items = make(map[string]any)
	}
	m.items[key] = value
	return true
}

func (m *inMemoryCache) GetAny(key string) (any, bool) {
	if m.items == nil {
		return nil, false
	}
	v, ok := m.items[key]
	return v, ok
}

func main() {
	opts := cep.GetOptions()
	opts.CacheEnabled = true
	opts.CacheTTL = 10 * time.Minute
	opts.SearchTimeout = 8 * time.Second
	cep.SetOptions(opts)
	cep.SetCacheProvider(&inMemoryCache{})

	cepCode := "01001-000"
	result, address, err := cep.Search(cepCode)
	if err != nil {
		fmt.Println("search error:", err)
		return
	}

	fmt.Println("cep:", cepCode)
	fmt.Println("json:", result)
	fmt.Println("address:", address)
}
