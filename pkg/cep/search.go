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
	"golang.org/x/sync/singleflight"
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

var searchSingleflight singleflight.Group

// Search concurrently looks up a CEP using providers declared in [models/endpoints.go].
// It returns the first JSON response and the normalized CEP address payload.
func Search(cep string) (jsonCep string, address models.CEPAddress, err error) {
	normalizedCEP, ok := normalizeSearchInput(cep)
	if !ok {
		return config.JsonDefault, models.CEPAddress{}, nil
	}

	if cachedJSON, cachedAddress, found := readCachedResult(normalizedCEP); found {
		return cachedJSON, cachedAddress, nil
	}

	if len(models.Endpoints) == 0 {
		return config.JsonDefault, models.CEPAddress{}, nil
	}

	return searchProvidersSingleflight(normalizedCEP)
}

func normalizeSearchInput(cep string) (string, bool) {
	if cep == "" {
		return "", false
	}

	normalizedCEP, err := util.NormalizeCEP(cep)
	if err != nil {
		return "", false
	}

	return normalizedCEP, true
}

func readCachedResult(cep string) (jsonCep string, address models.CEPAddress, found bool) {
	if !config.CacheEnabled {
		return "", models.CEPAddress{}, false
	}

	value, ok := gocache.GetAny(cep)
	if !ok {
		return "", models.CEPAddress{}, false
	}

	switch cached := value.(type) {
	case cachedResult:
		if cached.JSON == "" || !isCompleteAddress(cached.Address) {
			return "", models.CEPAddress{}, false
		}
		return cached.JSON, cached.Address, true
	case string:
		var parsedAddress models.CEPAddress
		if err := json.Unmarshal([]byte(cached), &parsedAddress); err != nil {
			return "", models.CEPAddress{}, false
		}
		if !isCompleteAddress(parsedAddress) {
			return "", models.CEPAddress{}, false
		}
		if !gocache.IsRedisBackend() {
			_ = gocache.SetAnyTTL(cep, cachedResult{
				JSON:    cached,
				Address: parsedAddress,
			}, time.Duration(config.TTLCache)*time.Second)
		}
		return cached, parsedAddress, true
	default:
		return "", models.CEPAddress{}, false
	}
}

func searchProvidersSingleflight(cep string) (string, models.CEPAddress, error) {
	value, _, _ := searchSingleflight.Do(cep, func() (any, error) {
		jsonCep, address := searchProviders(cep)
		return cachedResult{
			JSON:    jsonCep,
			Address: address,
		}, nil
	})

	result, ok := value.(cachedResult)
	if !ok {
		return config.JsonDefault, models.CEPAddress{}, nil
	}

	return result.JSON, result.Address, nil
}

func searchProviders(cep string) (jsonCep string, address models.CEPAddress) {
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
			dispatchProviderRequest(ctx, cancel, cep, endpoint, results)
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
				return config.JsonDefault, models.CEPAddress{}
			}
			jsonCep = string(result.Body)
			address = result.Address
			cacheSearchResult(cep, jsonCep, address)
			return jsonCep, address
		case <-ctx.Done():
			return config.JsonDefault, models.CEPAddress{}
		}
	}
}

func dispatchProviderRequest(ctx context.Context, cancel context.CancelFunc, cep string, endpoint models.Endpoint, results chan<- Result) {
	if endpoint.Source == models.SourceCorreio {
		requestCorreio(ctx, cancel, cep, endpoint.Method, endpoint.URL, endpoint.Body, results)
		return
	}
	requestProvider(ctx, cancel, cep, endpoint.Source, endpoint.Method, endpoint.URL, results)
}

func cacheSearchResult(cep, jsonCep string, address models.CEPAddress) {
	if !config.CacheEnabled || !isCompleteAddress(address) {
		return
	}

	gocache.SetAnyTTL(cep, cachedResult{
		JSON:    jsonCep,
		Address: address,
	}, time.Duration(config.TTLCache)*time.Second)
}

func ValidCEP(address models.CEPAddress) bool {
	return isCompleteAddress(address)
}

// Deprecated: use ValidCEP.
func ValidCep(address models.CEPAddress) bool {
	return ValidCEP(address)
}
