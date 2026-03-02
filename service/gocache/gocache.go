package gocache

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/cssbruno/gocep/config"
	"github.com/dgraph-io/ristretto/v2"
	redis "github.com/redis/go-redis/v9"
)

var (
	cacheOnce sync.Once
	cacheInst *ristretto.Cache[string, any]
	redisInst *redis.Client
	backend   string
)

const (
	backendMemory = "memory"
	backendRedis  = "redis"
)

func Run() *ristretto.Cache[string, any] {
	ensureBackend()
	return cacheInst
}

func SetTTL(key, value string, ttl time.Duration) bool {
	if len(key) == 0 || len(value) == 0 {
		return false
	}
	ensureBackend()

	if backend == backendRedis {
		return redisSet(key, value, ttl)
	}

	cacheInst.SetWithTTL(key, value, 1, ttl)
	cacheInst.Wait()
	return true
}

func SetAnyTTL(key string, value any, ttl time.Duration) bool {
	if len(key) == 0 || value == nil {
		return false
	}
	ensureBackend()

	if backend == backendRedis {
		serialized, ok := serializeForRedis(value)
		if !ok {
			return false
		}
		return redisSet(key, serialized, ttl)
	}

	cacheInst.SetWithTTL(key, value, 1, ttl)
	cacheInst.Wait()
	return true
}

func GetAny(key string) (any, bool) {
	if len(key) == 0 {
		return nil, false
	}

	ensureBackend()

	if backend == backendRedis {
		value, found := redisGet(key)
		if !found {
			return nil, false
		}
		return value, true
	}

	value, found := cacheInst.Get(key)
	if !found {
		return nil, false
	}

	return value, true
}

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
	}
	return ""
}

func IsRedisBackend() bool {
	ensureBackend()
	return backend == backendRedis
}

func ensureBackend() {
	cacheOnce.Do(func() {
		ristrettoConfig := &ristretto.Config[string, any]{
			NumCounters: config.NumCounters,
			MaxCost:     config.MaxCost,
			BufferItems: config.BufferItems,
		}
		if ristrettoConfig.NumCounters <= 0 {
			ristrettoConfig.NumCounters = 1e7
		}
		if ristrettoConfig.MaxCost <= 0 {
			ristrettoConfig.MaxCost = 1 << 30
		}
		if ristrettoConfig.BufferItems <= 0 {
			ristrettoConfig.BufferItems = 64
		}

		cache, err := ristretto.NewCache(ristrettoConfig)
		if err != nil {
			// Fall back to known-safe defaults if env values are invalid.
			cache, _ = ristretto.NewCache(&ristretto.Config[string, any]{
				NumCounters: 1e7,
				MaxCost:     1 << 30,
				BufferItems: 64,
			})
		}
		cacheInst = cache

		configured := strings.ToLower(strings.TrimSpace(config.CacheBackend))
		if configured == backendRedis && initRedisBackend() {
			backend = backendRedis
			return
		}
		backend = backendMemory
	})
}

func initRedisBackend() bool {
	client := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Username: config.RedisUser,
		Password: config.RedisPass,
		DB:       config.RedisDB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return false
	}

	redisInst = client
	return true
}

func redisSet(key, value string, ttl time.Duration) bool {
	if redisInst == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	return redisInst.Set(ctx, redisKey(key), value, ttl).Err() == nil
}

func redisGet(key string) (string, bool) {
	if redisInst == nil {
		return "", false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	value, err := redisInst.Get(ctx, redisKey(key)).Result()
	switch {
	case err == nil:
		return value, true
	case errors.Is(err, redis.Nil):
		return "", false
	default:
		return "", false
	}
}

func redisKey(key string) string {
	if config.RedisPrefix == "" {
		return key
	}
	return config.RedisPrefix + key
}

func serializeForRedis(value any) (string, bool) {
	switch v := value.(type) {
	case string:
		return v, v != ""
	case []byte:
		if len(v) == 0 {
			return "", false
		}
		return string(v), true
	}

	rv := reflect.ValueOf(value)
	if rv.IsValid() {
		if rv.Kind() == reflect.Pointer && !rv.IsNil() {
			rv = rv.Elem()
		}
		if rv.Kind() == reflect.Struct {
			field := rv.FieldByName("JSON")
			if field.IsValid() && field.Kind() == reflect.String {
				cachedJSON := strings.TrimSpace(field.String())
				if cachedJSON != "" {
					return cachedJSON, true
				}
			}
		}
	}

	return "", false
}

func resetCacheForTests() {
	if cacheInst != nil {
		cacheInst.Close()
	}
	if redisInst != nil {
		_ = redisInst.Close()
	}
	cacheInst = nil
	redisInst = nil
	backend = ""
	cacheOnce = sync.Once{}
}
