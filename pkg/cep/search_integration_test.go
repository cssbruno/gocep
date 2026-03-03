package cep

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cssbruno/gocep/models"
	"github.com/cssbruno/gocep/service/gocache"
)

func TestSearchIgnoresInvalidStringCacheAndUsesProviderResult(t *testing.T) {
	const cepCode = "77777777"
	const expectedBody = `{"cidade":"São Paulo","uf":"SP","logradouro":"Praça da Sé","bairro":"Sé"}`

	provider := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/"+cepCode {
			t.Fatalf("provider path = %s, want /%s", r.URL.Path, cepCode)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"cep":"77777-777","logradouro":"Praça da Sé","bairro":"Sé","localidade":"São Paulo","uf":"SP"}`))
	}))
	defer provider.Close()
	useServerHTTPClient(t, provider)
	useTestCacheProvider(t)

	useTestEndpoints(t, []models.Endpoint{
		{Method: models.MethodGet, Source: models.SourceViaCep, URL: provider.URL + "/%s"},
	})

	useTestOptions(t, func(o *Options) {
		o.CacheEnabled = true
	})

	// Seed cache with invalid JSON string to simulate old/bad cached payload.
	if ok := gocache.SetTTL(cepCode, "not-json", time.Minute); !ok {
		t.Fatalf("failed to seed invalid cache")
	}

	gotBody, gotAddress, err := Search(cepCode)
	if err != nil {
		t.Fatalf("Search() error = %v, want nil", err)
	}
	if gotBody != expectedBody {
		t.Fatalf("Search() body = %s, want %s", gotBody, expectedBody)
	}
	if !ValidCEP(gotAddress) {
		t.Fatalf("Search() returned invalid normalized CEP: %+v", gotAddress)
	}
}

func TestSearchIntegrationFallbackAcrossHTTPProviders(t *testing.T) {
	const cepCode = "01001000"
	const expectedBody = `{"cidade":"Sao Paulo","uf":"SP","logradouro":"Rua Fallback","bairro":"Centro"}`

	var failCalls atomic.Int32
	var successCalls atomic.Int32
	provider := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/fail/"+cepCode:
			failCalls.Add(1)
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte(`{"error":"upstream"}`))
		case r.URL.Path == "/success/"+cepCode:
			successCalls.Add(1)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"cep":"01001-000","logradouro":"Rua Fallback","bairro":"Centro","localidade":"Sao Paulo","uf":"SP"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer provider.Close()
	useServerHTTPClient(t, provider)

	useTestEndpoints(t, []models.Endpoint{
		{Method: models.MethodGet, Source: models.SourceViaCep, URL: provider.URL + "/fail/%s"},
		{Method: models.MethodGet, Source: models.SourceViaCep, URL: provider.URL + "/success/%s"},
	})

	useTestOptions(t, func(o *Options) {
		o.CacheEnabled = false
	})

	gotBody, gotAddress, err := Search(cepCode)
	if err != nil {
		t.Fatalf("Search() error = %v, want nil", err)
	}
	if gotBody != expectedBody {
		t.Fatalf("Search() body = %s, want %s", gotBody, expectedBody)
	}
	if !ValidCEP(gotAddress) {
		t.Fatalf("Search() address = %+v, want valid", gotAddress)
	}
	if successCalls.Load() != 1 {
		t.Fatalf("success provider calls = %d, want 1", successCalls.Load())
	}
	if failCalls.Load() > 1 {
		t.Fatalf("fail provider calls = %d, want <= 1", failCalls.Load())
	}
}

func TestSearchIntegrationFallbackFromCorreioToViaCEP(t *testing.T) {
	const cepCode = "01001000"
	const expectedBody = `{"cidade":"Sao Paulo","uf":"SP","logradouro":"Rua ViaCEP","bairro":"Centro"}`

	var correioCalls atomic.Int32
	var viaCalls atomic.Int32
	provider := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/correio":
			correioCalls.Add(1)
			if r.Method != http.MethodPost {
				t.Fatalf("correio method = %s, want %s", r.Method, http.MethodPost)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`invalid-xml`))
		case r.URL.Path == "/viacep/"+cepCode:
			viaCalls.Add(1)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"cep":"01001-000","logradouro":"Rua ViaCEP","bairro":"Centro","localidade":"Sao Paulo","uf":"SP"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer provider.Close()
	useServerHTTPClient(t, provider)

	useTestEndpoints(t, []models.Endpoint{
		{
			Method: models.MethodPost,
			Source: models.SourceCorreio,
			URL:    provider.URL + "/correio",
			Body:   models.PayloadCorreio,
		},
		{Method: models.MethodGet, Source: models.SourceViaCep, URL: provider.URL + "/viacep/%s"},
	})

	useTestOptions(t, func(o *Options) {
		o.CacheEnabled = false
	})

	gotBody, gotAddress, err := Search(cepCode)
	if err != nil {
		t.Fatalf("Search() error = %v, want nil", err)
	}
	if gotBody != expectedBody {
		t.Fatalf("Search() body = %s, want %s", gotBody, expectedBody)
	}
	if !ValidCEP(gotAddress) {
		t.Fatalf("Search() address = %+v, want valid", gotAddress)
	}
	if viaCalls.Load() != 1 {
		t.Fatalf("via provider calls = %d, want 1", viaCalls.Load())
	}
	if correioCalls.Load() > 1 {
		t.Fatalf("correio provider calls = %d, want <= 1", correioCalls.Load())
	}
}

func TestSearchIgnoresIncompleteStringCacheAndUsesProviderResult(t *testing.T) {
	const cepCode = "66666666"
	const expectedBody = `{"cidade":"São Paulo","uf":"SP","logradouro":"Rua Completa","bairro":"Centro"}`

	var calls atomic.Int32
	provider := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"cep":"66666-666","logradouro":"Rua Completa","bairro":"Centro","localidade":"São Paulo","uf":"SP"}`))
	}))
	defer provider.Close()
	useServerHTTPClient(t, provider)
	useTestCacheProvider(t)

	useTestEndpoints(t, []models.Endpoint{
		{Method: models.MethodGet, Source: models.SourceViaCep, URL: provider.URL + "/%s"},
	})

	useTestOptions(t, func(o *Options) {
		o.CacheEnabled = true
	})

	// Incomplete payload must not be treated as a valid cache hit.
	if ok := gocache.SetTTL(cepCode, `{"cidade":"São Paulo","uf":"","logradouro":"","bairro":""}`, time.Minute); !ok {
		t.Fatalf("failed to seed incomplete cache")
	}

	gotBody, gotAddress, err := Search(cepCode)
	if err != nil {
		t.Fatalf("Search() error = %v, want nil", err)
	}
	if gotBody != expectedBody {
		t.Fatalf("Search() body = %s, want %s", gotBody, expectedBody)
	}
	if !ValidCEP(gotAddress) {
		t.Fatalf("Search() returned invalid normalized CEP: %+v", gotAddress)
	}
	if calls.Load() != 1 {
		t.Fatalf("provider calls = %d, want 1", calls.Load())
	}
}

func TestSearchConcurrentCallsAreDeduplicated(t *testing.T) {
	const cepCode = "02020020"
	const expectedBody = `{"cidade":"São Paulo","uf":"SP","logradouro":"Rua Dedupe","bairro":"Centro"}`

	var calls atomic.Int32
	provider := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"cep":"02020-020","logradouro":"Rua Dedupe","bairro":"Centro","localidade":"São Paulo","uf":"SP"}`))
	}))
	defer provider.Close()
	useServerHTTPClient(t, provider)

	useTestEndpoints(t, []models.Endpoint{
		{Method: models.MethodGet, Source: models.SourceViaCep, URL: provider.URL + "/%s"},
	})

	useTestOptions(t, func(o *Options) {
		o.CacheEnabled = false
	})

	const workers = 20
	var wg sync.WaitGroup
	wg.Add(workers)

	errs := make(chan string, workers)
	for range workers {
		go func() {
			defer wg.Done()
			gotBody, gotAddress, err := Search(cepCode)
			if err != nil {
				errs <- err.Error()
				return
			}
			if gotBody != expectedBody {
				errs <- "unexpected body"
				return
			}
			if !ValidCEP(gotAddress) {
				errs <- "unexpected invalid address"
			}
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Fatalf("concurrent Search() error: %s", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("provider calls = %d, want 1", calls.Load())
	}
}
