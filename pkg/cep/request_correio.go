package cep

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"net/http"
	"strings"

	"github.com/cssbruno/gocep/models"
)

// requestCorreio performs concurrent lookups against Correio SOAP API.
func requestCorreio(ctx context.Context, cancel context.CancelFunc, cep, method, endpoint, payload string, chResult chan<- Result) {
	payload = strings.Replace(payload, "%s", cep, 1)
	req, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader([]byte(payload)))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "text/xml; charset=utf-8")

	response, err := httpClient.Do(req)
	if err != nil {
		return
	}
	if response == nil {
		return
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return
	}

	correio := new(models.Correio)
	err = xml.NewDecoder(response.Body).Decode(correio)
	if err == nil {
		c := correio.Body.LookupCEPResponse.Return
		wecep := models.WeCep{
			City:         c.City,
			StateCode:    c.StateCode,
			Street:       c.Address,
			Neighborhood: c.Neighborhood,
		}
		b, err := json.Marshal(wecep)
		if err == nil {
			select {
			case chResult <- Result{Body: b, WeCep: wecep}:
				cancel()
			case <-ctx.Done():
			}
		}
	}
}

// Deprecated: use internal requestCorreio.
func NewRequestWithContextCorreio(ctx context.Context, cancel context.CancelFunc, cep, source, method, endpoint, payload string, chResult chan<- Result) {
	_ = source
	requestCorreio(ctx, cancel, cep, method, endpoint, payload, chResult)
}
