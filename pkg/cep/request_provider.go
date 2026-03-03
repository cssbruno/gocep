package cep

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/cssbruno/gocep/models"
)

var marshalAddressJSON = json.Marshal

// requestProvider performs a concurrent request to one CEP provider endpoint.
func requestProvider(ctx context.Context, cancel context.CancelFunc, cep, source, method,
	endpoint string, chResult chan<- Result) {
	if source == models.SourceCdnApiCep && len(cep) > 7 {
		cep = addHyphen(cep)
	}
	endpoint = strings.Replace(endpoint, "%s", cep, 1)
	req, err := http.NewRequestWithContext(ctx, method, endpoint, nil)
	if err != nil {
		return
	}

	response, ok := executeRequest(req)
	if !ok {
		return
	}
	defer response.Body.Close()

	maxBody := GetOptions().MaxProviderBody
	body, err := io.ReadAll(io.LimitReader(response.Body, maxBody+1))
	if err != nil {
		return
	}
	if int64(len(body)) > maxBody {
		return
	}

	if len(body) == 0 {
		return
	}

	address, err := ParseCEPAddress(source, body)
	if err != nil {
		return
	}

	sendAddressResult(ctx, cancel, chResult, address)
}

func addHyphen(s string) string {
	n := len(s)
	if n <= 5 {
		return s
	}
	return s[:5] + "-" + s[5:]
}
