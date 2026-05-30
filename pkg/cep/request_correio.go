package cep

import (
	"context"

	"github.com/cssbruno/gocep/v2/models"
)

// requestCorreio performs concurrent lookups against Correio SOAP API.
func requestCorreio(ctx context.Context, cancel context.CancelFunc, cep, method, endpoint, payload string, chResult chan<- Result) {
	result, err := queryCorreioEndpoint(ctx, getHTTPClient(), GetOptions().MaxProviderBody, cep, models.Endpoint{
		Method: method,
		Source: models.SourceCorreio,
		URL:    endpoint,
		Body:   payload,
	})
	if err != nil {
		return
	}

	sendResult(ctx, cancel, chResult, result)
}
