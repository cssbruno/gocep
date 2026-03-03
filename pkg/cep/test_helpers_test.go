package cep

import (
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/cssbruno/gocep/models"
	"github.com/cssbruno/gocep/service/gocache"
)

func useServerHTTPClient(t *testing.T, server *httptest.Server) {
	t.Helper()

	oldClient := getHTTPClient()
	SetHTTPClient(server.Client())
	t.Cleanup(func() {
		SetHTTPClient(oldClient)
	})
}

func useTestOptions(t *testing.T, mutate func(*Options)) {
	t.Helper()

	old := GetOptions()
	next := old
	mutate(&next)
	SetOptions(next)
	t.Cleanup(func() {
		SetOptions(old)
	})
}

type testCacheProvider struct {
	mu    sync.RWMutex
	items map[string]any
}

func (m *testCacheProvider) SetAnyTTL(key string, value any, _ time.Duration) bool {
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

func (m *testCacheProvider) GetAny(key string) (any, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.items[key]
	return v, ok
}

func useTestCacheProvider(t *testing.T) *testCacheProvider {
	t.Helper()
	provider := &testCacheProvider{}
	gocache.SetProvider(provider)
	t.Cleanup(func() {
		gocache.SetProvider(nil)
	})
	return provider
}

func useTestEndpoints(t *testing.T, endpoints []models.Endpoint) {
	t.Helper()
	old := models.GetEndpoints()
	models.SetEndpoints(endpoints)
	t.Cleanup(func() {
		models.SetEndpoints(old)
	})
}
