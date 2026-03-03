package cep

import (
	"net/http"
	"time"
)

const (
	defaultHTTPMaxIdleConns        = 100
	defaultHTTPMaxIdleConnsPerHost = 10
	defaultHTTPIdleConnTimeout     = 90 * time.Second
	defaultHTTPTimeout             = 30 * time.Second
)

// SetHTTPClient overrides the HTTP client used for provider requests.
// Passing nil restores the package default client.
// The setting is global for the current process.
func SetHTTPClient(client *http.Client) {
	defaultClient.SetHTTPClient(client)
}

func getHTTPClient() *http.Client {
	return defaultClient.HTTPClient()
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
