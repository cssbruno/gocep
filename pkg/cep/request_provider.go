package cep

import (
	"context"

	"github.com/cssbruno/gocep/v2/models"
)

// requestProvider performs a concurrent request to one CEP provider endpoint.
func requestProvider(ctx context.Context, cancel context.CancelFunc, cep, source, method,
	endpoint string, chResult chan<- Result) {
	result, err := queryJSONEndpoint(ctx, getHTTPClient(), GetOptions().MaxProviderBody, cep, models.Endpoint{
		Method: method,
		Source: source,
		URL:    endpoint,
	})
	if err != nil {
		return
	}

	sendResult(ctx, cancel, chResult, result)
}

func addHyphen(s string) string {
	n := len(s)
	if n <= 5 {
		return s
	}
	return s[:5] + "-" + s[5:]
}
