package cep

import (
	"context"
	"encoding/json"
	"time"

	"github.com/cssbruno/gocep/config"
	"github.com/cssbruno/gocep/models"
	"github.com/cssbruno/gocep/service/gocache"
)

// Result represents one provider response payload.
type Result struct {
	Body  []byte
	WeCep models.WeCep
}

type cachedResult struct {
	JSON  string
	WeCep models.WeCep
}

// Search concurrently looks up a CEP using providers declared in [models/endpoints.go].
// It returns the first JSON response and the normalized WeCep payload.
func Search(cep string) (jsonCep string, wecep models.WeCep, err error) {
	if len(cep) == 0 {
		jsonCep = config.JsonDefault
		return
	}

	if config.CacheEnabled {
		if value, found := gocache.GetAny(cep); found {
			switch v := value.(type) {
			case cachedResult:
				return v.JSON, v.WeCep, nil
			case string:
				jsonCep = v
				if err = json.Unmarshal([]byte(jsonCep), &wecep); err == nil {
					_ = gocache.SetAnyTTL(cep, cachedResult{
						JSON:  jsonCep,
						WeCep: wecep,
					}, time.Duration(config.TTLCache)*time.Second)
					return
				}
			}
		}
	}

	results := make(chan Result, len(models.Endpoints))
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(config.TimeoutSearchCEP)*time.Second)
	defer cancel()

	for _, endpoint := range models.Endpoints {
		endpoint := endpoint
		go func() {
			if endpoint.Source == models.SourceCorreio {
				requestCorreio(ctx, cancel, cep, endpoint.Method, endpoint.URL, endpoint.Body, results)
				return
			}
			requestProvider(ctx, cancel, cep, endpoint.Source, endpoint.Method, endpoint.URL, results)
		}()
	}

	select {
	case result := <-results:
		jsonCep = string(result.Body)
		wecep = result.WeCep
		if wecep.City != "" && wecep.Street != "" && wecep.StateCode != "" && wecep.Neighborhood != "" {
			if config.CacheEnabled {
				gocache.SetAnyTTL(cep, cachedResult{
					JSON:  jsonCep,
					WeCep: wecep,
				}, time.Duration(config.TTLCache)*time.Second)
			}
		}
		return

	case <-ctx.Done():
	}
	jsonCep = config.JsonDefault
	return
}

func ValidCEP(wecep models.WeCep) bool {
	if len(wecep.City) == 0 &&
		len(wecep.StateCode) == 0 &&
		len(wecep.Street) == 0 &&
		len(wecep.Neighborhood) == 0 {
		return false
	}
	return true
}

// Deprecated: use ValidCEP.
func ValidCep(wecep models.WeCep) bool {
	return ValidCEP(wecep)
}
