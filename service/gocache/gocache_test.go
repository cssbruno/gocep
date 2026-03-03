package gocache

import (
	"sync"
	"testing"
	"time"
)

func TestSetAnyTTL_WithoutProvider(t *testing.T) {
	resetCacheForTests()

	if ok := SetAnyTTL("key", "value", time.Second); ok {
		t.Fatalf("SetAnyTTL() = true, want false when no provider is configured")
	}
	if got, found := GetAny("key"); found || got != nil {
		t.Fatalf("GetAny() = (%v,%v), want (nil,false) when no provider is configured", got, found)
	}
}

func TestSetTTL(t *testing.T) {
	resetCacheForTests()
	SetProvider(&mockProvider{})

	if ok := SetTTL("08226021", `{"cidade":"São Paulo","uf":"SP"}`, 5*time.Second); !ok {
		t.Fatalf("SetTTL() = false, want true")
	}
	if ok := SetTTL("", "value", 5*time.Second); ok {
		t.Fatalf("SetTTL() with empty key = true, want false")
	}
	if ok := SetTTL("key", "", 5*time.Second); ok {
		t.Fatalf("SetTTL() with empty value = true, want false")
	}
}

func TestGet(t *testing.T) {
	resetCacheForTests()
	SetProvider(&mockProvider{})

	want := `{"cidade":"São Paulo","uf":"SP"}`
	if ok := SetTTL("08226021", want, 5*time.Second); !ok {
		t.Fatalf("SetTTL() = false, want true")
	}
	if got := Get("08226021"); got != want {
		t.Fatalf("Get() = %q, want %q", got, want)
	}
	if got := Get(""); got != "" {
		t.Fatalf("Get(\"\") = %q, want empty", got)
	}
	if got := Get("unknown"); got != "" {
		t.Fatalf("Get(\"unknown\") = %q, want empty", got)
	}
}

func TestSetAnyTTLAndGetAny(t *testing.T) {
	resetCacheForTests()
	SetProvider(&mockProvider{})

	type cacheValue struct {
		Name string
	}
	want := cacheValue{Name: "gocep"}

	if ok := SetAnyTTL("typed", want, time.Second); !ok {
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
	if cast != want {
		t.Fatalf("GetAny() = %+v, want %+v", cast, want)
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
	SetProvider(&mockProvider{})

	if ok := SetAnyTTL("typed-not-string", struct{ ID int }{ID: 1}, time.Second); !ok {
		t.Fatalf("SetAnyTTL() = false, want true")
	}
	if got := Get("typed-not-string"); got != "" {
		t.Fatalf("Get() = %q, want empty string", got)
	}
}

type mockProvider struct {
	mu    sync.RWMutex
	items map[string]any
}

func (m *mockProvider) SetAnyTTL(key string, value any, _ time.Duration) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if key == "" || value == nil {
		return false
	}
	if m.items == nil {
		m.items = map[string]any{}
	}
	m.items[key] = value
	return true
}

func (m *mockProvider) GetAny(key string) (any, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.items[key]
	return v, ok
}

func TestSetProvider_UsesCustomImplementation(t *testing.T) {
	resetCacheForTests()

	provider := &mockProvider{}
	SetProvider(provider)
	t.Cleanup(func() { SetProvider(nil) })

	if ok := SetAnyTTL("custom-key", "custom-value", time.Second); !ok {
		t.Fatalf("SetAnyTTL() = false, want true")
	}

	got, found := GetAny("custom-key")
	if !found {
		t.Fatalf("GetAny() found = false, want true")
	}
	if gotStr, ok := got.(string); !ok || gotStr != "custom-value" {
		t.Fatalf("GetAny() = %v (%T), want custom-value", got, got)
	}
}
