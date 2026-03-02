package cep

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/cssbruno/gocep/models"
)

// requestProvider performs a concurrent request to one CEP provider endpoint.
func requestProvider(ctx context.Context, cancel context.CancelFunc, cep, source, method,
	endpoint string, chResult chan<- Result) {
	if source == models.SourceCdnApiCep && len(cep) > 7 {
		cep = addHyphen(cep)
	}
	endpoint = strings.Replace(endpoint, "%s", cep, 1)
	req, err := http.NewRequestWithContext(ctx, method, endpoint, nil)
	if err != nil {
		log.Println("error creating provider request:", err)
		return
	}

	// fmt.Println(endpoint)

	response, err := httpClient.Do(req)
	if err != nil {
		// log.Println("Error httpClient:", err)
		return
	}
	if response == nil {
		return
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Println("Error io.ReadAll:", err)
		return
	}

	if len(body) == 0 {
		return
	}

	wecep, err := ParseWeCep(source, body)
	if err != nil {
		return
	}

	b, err := json.Marshal(wecep)
	if err != nil {
		return
	}

	select {
	case chResult <- Result{Body: b, WeCep: wecep}:
		cancel()
	case <-ctx.Done():
	}
}

// Deprecated: use internal requestProvider.
func NewRequestWithContext(ctx context.Context, cancel context.CancelFunc, cep, source, method,
	endpoint string, chResult chan<- Result) {
	requestProvider(ctx, cancel, cep, source, method, endpoint, chResult)
}

func addHyphen(s string) string {
	n := len(s)
	if n <= 5 {
		return s
	}
	return s[:5] + "-" + s[5:]
}
