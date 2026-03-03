package cep

import (
	"context"
	"net/http"
	"strings"

	"github.com/cssbruno/gocep/models"
)

func executeRequest(req *http.Request) (*http.Response, bool) {
	if req == nil || req.URL == nil || !strings.EqualFold(req.URL.Scheme, "https") {
		return nil, false
	}

	response, err := getHTTPClient().Do(req)
	if err != nil {
		if response != nil && response.Body != nil {
			_ = response.Body.Close()
		}
		return nil, false
	}
	if response == nil {
		return nil, false
	}
	if response.StatusCode != http.StatusOK {
		_ = response.Body.Close()
		return nil, false
	}
	return response, true
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
	return address.City != "" &&
		address.StateCode != "" &&
		address.Street != "" &&
		address.Neighborhood != ""
}
