package gocache

import (
	"reflect"
	"testing"
	"time"

	"github.com/cssbruno/gocep/config"
	gcache "github.com/patrickmn/go-cache"
)

// go test -run ^TestRun'$ -v
func TestRun(t *testing.T) {
	resetCacheForTests()

	tests := []struct {
		name    string
		want    *gcache.Cache
		wantErr bool
	}{
		{name: "test_run_", want: Run(), wantErr: false},
		{name: "test_run_", want: gcache.New(24*time.Hour, 24*time.Hour), wantErr: true},
		{name: "test_run_", want: new(gcache.Cache), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Run(); !reflect.DeepEqual(got, tt.want) && !tt.wantErr {
				t.Errorf("Run() = %v, want %v", got, tt.want)
			}
		})
	}
}

// go test -run ^TestSetTTL'$ -v
func TestSetTTL(t *testing.T) {
	resetCacheForTests()
	TestRun(t)
	type args struct {
		key   string
		value string
		ttl   time.Duration
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test_setTTL_",
			args: args{
				key:   `08226021`,
				value: `{"cidade":"São Paulo","uf":"SP","logradouro":"18 de Abril","bairro":"Cidade Antônio Estevão de Carvalho"}`,
				ttl:   time.Duration(5) * time.Second,
			},
			want: true,
		},
		{
			name: "test_setTTL_",
			args: args{
				key:   ``,
				value: `{"cidade":"São Paulo","uf":"SP","logradouro":"18 de Abril","bairro":"Cidade Antônio Estevão de Carvalho"}`,
				ttl:   time.Duration(5) * time.Second,
			},
			want: false,
		},
		{
			name: "test_setTTL_",
			args: args{
				key:   `08226021`,
				value: ``,
				ttl:   time.Duration(5) * time.Second,
			},
			want: false,
		},
		{
			name: "test_setTTL_",
			args: args{
				key:   ``,
				value: ``,
				ttl:   time.Duration(5) * time.Second,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SetTTL(tt.args.key, tt.args.value, tt.args.ttl); got != tt.want {
				t.Errorf("SetTTL() = %v, want %v", got, tt.want)
			}
		})
	}
}

// go test -run ^TestGet'$ -v
func TestGet(t *testing.T) {
	resetCacheForTests()
	TestRun(t)
	TestSetTTL(t)
	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test_get_",
			args: args{
				key: `08226021`,
			},
			want: `{"cidade":"São Paulo","uf":"SP","logradouro":"18 de Abril","bairro":"Cidade Antônio Estevão de Carvalho"}`,
		},
		{
			name: "test_get_",
			args: args{
				key: ``,
			},
			want: ``,
		},
		{
			name: "test_get_",
			args: args{
				key: `01001000`,
			},
			want: ``,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Get(tt.args.key); got != tt.want {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetAnyTTLAndGetAny(t *testing.T) {
	resetCacheForTests()
	TestRun(t)

	type cacheValue struct {
		Name string
	}

	value := cacheValue{Name: "gocep"}
	if ok := SetAnyTTL("typed", value, time.Second); !ok {
		t.Fatalf("SetAnyTTL() = false, want true")
	}

	got, found := GetAny("typed")
	if !found {
		t.Fatalf("GetAny() found = false, want true")
	}

	cast, ok := got.(cacheValue)
	if !ok {
		t.Fatalf("GetAny() type assertion failed: %T", got)
	}
	if cast != value {
		t.Fatalf("GetAny() = %+v, want %+v", cast, value)
	}
}

func TestSetAnyTTL_InvalidInput(t *testing.T) {
	resetCacheForTests()
	if ok := SetAnyTTL("", "value", time.Second); ok {
		t.Fatalf("SetAnyTTL() with empty key = true, want false")
	}
	if ok := SetAnyTTL("key", nil, time.Second); ok {
		t.Fatalf("SetAnyTTL() with nil value = true, want false")
	}
}

func TestGetAny_EmptyKey(t *testing.T) {
	resetCacheForTests()
	got, found := GetAny("")
	if found {
		t.Fatalf("GetAny(\"\") found = true, want false")
	}
	if got != nil {
		t.Fatalf("GetAny(\"\") value = %v, want nil", got)
	}
}

func TestGet_NonStringValue(t *testing.T) {
	resetCacheForTests()
	if ok := SetAnyTTL("typed-not-string", struct{ ID int }{ID: 1}, time.Second); !ok {
		t.Fatalf("SetAnyTTL() = false, want true")
	}
	if got := Get("typed-not-string"); got != "" {
		t.Fatalf("Get() = %q, want empty string", got)
	}
}

func TestSerializeForRedis_JSONField(t *testing.T) {
	type cachedLike struct {
		JSON string
	}

	got, ok := serializeForRedis(cachedLike{JSON: `{"cidade":"São Paulo"}`})
	if !ok {
		t.Fatalf("serializeForRedis() ok = false, want true")
	}
	if got != `{"cidade":"São Paulo"}` {
		t.Fatalf("serializeForRedis() = %q, want %q", got, `{"cidade":"São Paulo"}`)
	}
}

func TestSerializeForRedis_UnsupportedType(t *testing.T) {
	_, ok := serializeForRedis(struct{ Value int }{Value: 1})
	if ok {
		t.Fatalf("serializeForRedis() ok = true, want false for unsupported type")
	}
}

func TestRun_RedisFallbackStillProvidesMemoryCache(t *testing.T) {
	resetCacheForTests()

	oldBackend := config.CacheBackend
	oldAddr := config.RedisAddr
	oldPrefix := config.RedisPrefix
	config.CacheBackend = "redis"
	config.RedisAddr = "127.0.0.1:1"
	config.RedisPrefix = "gocep:"
	t.Cleanup(func() {
		config.CacheBackend = oldBackend
		config.RedisAddr = oldAddr
		config.RedisPrefix = oldPrefix
	})

	if got := Run(); got == nil {
		t.Fatalf("Run() = nil, want non-nil memory cache fallback")
	}
}
