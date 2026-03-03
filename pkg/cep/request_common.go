package cep

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/cssbruno/gocep/models"
)

func executeRequest(req *http.Request) (*http.Response, bool) {
	response, err := executeRequestWithClient(getHTTPClient(), req)
	if err != nil {
		return nil, false
	}
	return response, true
}

func executeRequestWithClient(client *http.Client, req *http.Request) (*http.Response, error) {
	if req == nil || req.URL == nil || !strings.EqualFold(req.URL.Scheme, "https") {
		return nil, errors.New("non-tls provider URL")
	}
	if client == nil {
		client = newDefaultHTTPClient()
	}

	response, err := client.Do(req)
	if err != nil {
		if response != nil && response.Body != nil {
			_ = response.Body.Close()
		}
		return nil, err
	}
	if response == nil {
		return nil, errors.New("nil response")
	}
	if response.StatusCode != http.StatusOK {
		_ = response.Body.Close()
		return nil, fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}
	return response, nil
}

func sendAddressResult(ctx context.Context, cancel context.CancelFunc, chResult chan<- Result, address models.CEPAddress) {
	if !isCompleteAddress(address) {
		return
	}

	body, err := marshalAddressJSON(address)
	if err != nil {
		return
	}

	select {
	case chResult <- Result{Body: body, Address: address}:
		cancel()
	case <-ctx.Done():
	}
}

func isCompleteAddress(address models.CEPAddress) bool {
	return address.CEP != "" &&
		address.City != "" &&
		address.StateCode != "" &&
		address.Street != "" &&
		address.Neighborhood != ""
}
