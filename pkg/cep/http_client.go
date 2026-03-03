package cep

import (
	"net/http"
	"sync"
	"time"
)

const (
	defaultHTTPMaxIdleConns        = 100
	defaultHTTPMaxIdleConnsPerHost = 10
	defaultHTTPIdleConnTimeout     = 90 * time.Second
	defaultHTTPTimeout             = 30 * time.Second
)

var (
	httpClientMu sync.RWMutex
	httpClient   = newDefaultHTTPClient()
)

// SetHTTPClient overrides the HTTP client used for provider requests.
// Passing nil restores the package default client.
// The setting is global for the current process.
func SetHTTPClient(client *http.Client) {
	httpClientMu.Lock()
	defer httpClientMu.Unlock()

	if client == nil {
		httpClient = newDefaultHTTPClient()
		return
	}
	httpClient = client
}

func getHTTPClient() *http.Client {
	httpClientMu.RLock()
	defer httpClientMu.RUnlock()
	return httpClient
}

func newDefaultHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        defaultHTTPMaxIdleConns,
			MaxIdleConnsPerHost: defaultHTTPMaxIdleConnsPerHost,
			IdleConnTimeout:     defaultHTTPIdleConnTimeout,
		},
		Timeout: defaultHTTPTimeout,
	}
}
