package gocache

import (
	"sync"
	"time"

	gcache "github.com/patrickmn/go-cache"
)

var (
	cacheOnce sync.Once
	cacheInst *gcache.Cache
)

func Run() *gcache.Cache {
	cacheOnce.Do(func() {
		cacheInst = gcache.New(24*time.Hour, 24*time.Hour)
	})
	return cacheInst
}

func SetTTL(key, value string, ttl time.Duration) bool {
	if len(key) == 0 || len(value) == 0 {
		return false
	}
	g := Run()
	g.Set(key, value, ttl)
	return true
}

func SetAnyTTL(key string, value any, ttl time.Duration) bool {
	if len(key) == 0 || value == nil {
		return false
	}
	g := Run()
	g.Set(key, value, ttl)
	return true
}

func GetAny(key string) (any, bool) {
	if len(key) == 0 {
		return nil, false
	}
	g := Run()
	return g.Get(key)
}

func Get(key string) string {
	value, found := GetAny(key)
	if !found {
		return ""
	}
	strValue, ok := value.(string)
	if !ok {
		return ""
	}
	return strValue
}
