package cep

import (
	"context"
	"encoding/xml"
	"io"
	"net/http"
	"strings"

	"github.com/cssbruno/gocep/models"
)

// requestCorreio performs concurrent lookups against Correio SOAP API.
func requestCorreio(ctx context.Context, cancel context.CancelFunc, cep, method, endpoint, payload string, chResult chan<- Result) {
	payload = strings.Replace(payload, "%s", cep, 1)
	req, err := http.NewRequestWithContext(ctx, method, endpoint, strings.NewReader(payload))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "text/xml; charset=utf-8")

	response, ok := executeRequest(req)
	if !ok {
		return
	}
	defer response.Body.Close()

	maxBody := GetOptions().MaxProviderBody
	rawBody, err := io.ReadAll(io.LimitReader(response.Body, maxBody+1))
	if err != nil {
		return
	}
	if int64(len(rawBody)) > maxBody {
		return
	}

	correio := new(models.Correio)
	err = xml.Unmarshal(rawBody, correio)
	if err == nil {
		c := correio.Body.LookupCEPResponse.Return
		address := models.CEPAddress{
			City:         c.City,
			StateCode:    c.StateCode,
			Street:       c.Address,
			Neighborhood: c.Neighborhood,
		}
		sendAddressResult(ctx, cancel, chResult, address)
	}
}
