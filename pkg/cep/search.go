package cep

import (
	"context"
	"encoding/json"
	"sync"

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

// Search concurrently looks up a CEP using configured endpoints.
// It returns the first successful provider payload and normalized address.
// When no complete address is found, it returns Options.DefaultJSON and an empty address.
// The returned error is reserved for future use and is currently nil.
func Search(cep string) (jsonCep string, address models.CEPAddress, err error) {
	normalizedCEP, ok := normalizeSearchInput(cep)
	if !ok {
		return GetOptions().DefaultJSON, models.CEPAddress{}, nil
	}

	if cachedJSON, cachedAddress, found := readCachedResult(normalizedCEP); found {
		return cachedJSON, cachedAddress, nil
	}

	endpoints := models.GetEndpoints()
	if len(endpoints) == 0 {
		return GetOptions().DefaultJSON, models.CEPAddress{}, nil
	}

	return searchProvidersSingleflight(normalizedCEP, endpoints)
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
	opts := GetOptions()
	if !opts.CacheEnabled {
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
		_ = gocache.SetAnyTTL(cep, cachedResult{
			JSON:    cached,
			Address: parsedAddress,
		}, opts.CacheTTL)
		return cached, parsedAddress, true
	default:
		return "", models.CEPAddress{}, false
	}
}

func searchProvidersSingleflight(cep string, endpoints []models.Endpoint) (string, models.CEPAddress, error) {
	value, _, _ := searchSingleflight.Do(cep, func() (any, error) {
		jsonCep, address := searchProviders(cep, endpoints)
		return cachedResult{
			JSON:    jsonCep,
			Address: address,
		}, nil
	})

	result, ok := value.(cachedResult)
	if !ok {
		return GetOptions().DefaultJSON, models.CEPAddress{}, nil
	}

	return result.JSON, result.Address, nil
}

func searchProviders(cep string, endpoints []models.Endpoint) (jsonCep string, address models.CEPAddress) {
	opts := GetOptions()
	results := make(chan Result, len(endpoints))
	ctx, cancel := context.WithTimeout(context.Background(), opts.SearchTimeout)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(len(endpoints))
	for _, endpoint := range endpoints {
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
				return opts.DefaultJSON, models.CEPAddress{}
			}
			jsonCep = string(result.Body)
			address = result.Address
			cacheSearchResult(cep, jsonCep, address)
			return jsonCep, address
		case <-ctx.Done():
			return opts.DefaultJSON, models.CEPAddress{}
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
	opts := GetOptions()
	if !opts.CacheEnabled || !isCompleteAddress(address) {
		return
	}

	gocache.SetAnyTTL(cep, cachedResult{
		JSON:    jsonCep,
		Address: address,
	}, opts.CacheTTL)
}

// ValidCEP reports whether an address has all required normalized fields.
func ValidCEP(address models.CEPAddress) bool {
	return isCompleteAddress(address)
}
