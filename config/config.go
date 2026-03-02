package config

import (
	"runtime"
	"strings"
	"time"

	"github.com/cssbruno/gocep/pkg/env"
)

var (
	Port        = env.GetString("PORT", "0.0.0.0:8080")
	JsonDefault = `{"cidade":"","uf":"","logradouro":"","bairro":""}`

	NumCounters      = env.GetInt64("NUM_COUNTERS", 1e7) // Number of keys to track frequency.
	MaxCost          = env.GetInt64("MAX_COST", 1<<30)   // Maximum cost of cache (2GB).
	BufferItems      = env.GetInt64("BUFFER_ITEMS", 64)  // Number of keys per Get buffer.
	NumCPU           = env.GetInt("NUM_CPU", runtime.NumCPU())
	TimeoutSearchCEP = env.GetInt("TIMEOUT_SEARCH_CEP", 15) // seconds
	TTLCache         = env.GetInt("TTL_CACHE", 172800)      // seconds (2 days)
	MaxProviderBody  = env.GetInt64("MAX_PROVIDER_BODY", 1<<20)

	// httpClient is used to make HTTP requests with TLS configuration
	InsecureSkipVerify  = env.GetBool("INSECURE_SKIP_VERIFY", false)        // Skip TLS verification for testing purposes
	MaxIdleConns        = env.GetInt("HTTP_CLIENT_MAXIDLECONNS", 100)       // Maximum number of idle connections
	MaxIdleConnsPerHost = env.GetInt("HTTP_CLIENT_MAXIDLECONNSPERHOST", 10) // Maximum number of idle connections per host
	IdleConnTimeout     = time.Duration(env.GetInt("IDLE_CONN_TIMEOUT", 90)) * time.Second
	Timeout             = time.Duration(env.GetInt("TIMEOUT", 30)) * time.Second
	ServerReadHeaderTO  = time.Duration(env.GetInt("SERVER_READ_HEADER_TIMEOUT", 5)) * time.Second
	ServerReadTO        = time.Duration(env.GetInt("SERVER_READ_TIMEOUT", 10)) * time.Second
	ServerWriteTO       = time.Duration(env.GetInt("SERVER_WRITE_TIMEOUT", 15)) * time.Second
	ServerIdleTO        = time.Duration(env.GetInt("SERVER_IDLE_TIMEOUT", 120)) * time.Second
	ServerMaxHeaderB    = env.GetInt("SERVER_MAX_HEADER_BYTES", 1<<20)

	// Search cancellation timeout.
	CancelCTXSearch = time.Duration(env.GetInt("CANCEL_CTX_SEARCH", 30)) * time.Second

	CacheEnabled = env.GetBool("CACHE_ENABLE", true)
	CacheBackend = strings.ToLower(env.GetString("CACHE_BACKEND", "memory"))
	RedisAddr    = env.GetString("REDIS_ADDR", "127.0.0.1:6379")
	RedisUser    = env.GetString("REDIS_USER", "")
	RedisPass    = env.GetString("REDIS_PASSWORD", "")
	RedisDB      = env.GetInt("REDIS_DB", 0)
	RedisPrefix  = env.GetString("REDIS_PREFIX", "gocep:")
)
