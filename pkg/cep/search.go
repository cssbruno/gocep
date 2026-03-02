package cep

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/cssbruno/gocep/config"
	"github.com/cssbruno/gocep/models"
	"github.com/cssbruno/gocep/pkg/util"
	"github.com/cssbruno/gocep/service/gocache"
)

// Result represents one provider response payload.
type Result struct {
	Body    []byte
	Address models.CEPAddress
}

type cachedResult struct {
	JSON    string
	Address models.CEPAddress
}

// Search concurrently looks up a CEP using providers declared in [models/endpoints.go].
// It returns the first JSON response and the normalized CEP address payload.
func Search(cep string) (jsonCep string, address models.CEPAddress, err error) {
	if len(cep) == 0 {
		jsonCep = config.JsonDefault
		return
	}

	normalizedCEP, normalizeErr := util.NormalizeCEP(cep)
	if normalizeErr != nil {
		jsonCep = config.JsonDefault
		return
	}
	cep = normalizedCEP

	if config.CacheEnabled {
		if value, found := gocache.GetAny(cep); found {
			switch v := value.(type) {
			case cachedResult:
				return v.JSON, v.Address, nil
			case string:
				jsonCep = v
				if parseErr := json.Unmarshal([]byte(jsonCep), &address); parseErr == nil {
					_ = gocache.SetAnyTTL(cep, cachedResult{
						JSON:    jsonCep,
						Address: address,
					}, time.Duration(config.TTLCache)*time.Second)
					return
				}
			}
		}
	}

	if len(models.Endpoints) == 0 {
		jsonCep = config.JsonDefault
		return
	}

	results := make(chan Result, len(models.Endpoints))
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(config.TimeoutSearchCEP)*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(len(models.Endpoints))
	for _, endpoint := range models.Endpoints {
		endpoint := endpoint
		go func() {
			defer wg.Done()
			if endpoint.Source == models.SourceCorreio {
				requestCorreio(ctx, cancel, cep, endpoint.Method, endpoint.URL, endpoint.Body, results)
				return
			}
			requestProvider(ctx, cancel, cep, endpoint.Source, endpoint.Method, endpoint.URL, results)
		}()
	}
	go func() {
		wg.Wait()
		close(results)
	}()

	for {
		select {
		case result, ok := <-results:
			if !ok {
				jsonCep = config.JsonDefault
				err = nil
				return
			}
			jsonCep = string(result.Body)
			address = result.Address
			err = nil
			if isCompleteAddress(address) && config.CacheEnabled {
				gocache.SetAnyTTL(cep, cachedResult{
					JSON:    jsonCep,
					Address: address,
				}, time.Duration(config.TTLCache)*time.Second)
			}
			return

		case <-ctx.Done():
			jsonCep = config.JsonDefault
			err = nil
			return
		}
	}
}

func ValidCEP(address models.CEPAddress) bool {
	if len(address.City) == 0 &&
		len(address.StateCode) == 0 &&
		len(address.Street) == 0 &&
		len(address.Neighborhood) == 0 {
		return false
	}
	return true
}

// Deprecated: use ValidCEP.
func ValidCep(address models.CEPAddress) bool {
	return ValidCEP(address)
}
