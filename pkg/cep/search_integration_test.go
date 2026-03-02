package cep

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cssbruno/gocep/config"
	"github.com/cssbruno/gocep/models"
	"github.com/cssbruno/gocep/service/gocache"
)

func TestSearchIgnoresInvalidStringCacheAndUsesProviderResult(t *testing.T) {
	const cepCode = "77777777"
	const expectedBody = `{"cidade":"São Paulo","uf":"SP","logradouro":"Praça da Sé","bairro":"Sé"}`

	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/"+cepCode {
			t.Fatalf("provider path = %s, want /%s", r.URL.Path, cepCode)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"cep":"77777-777","logradouro":"Praça da Sé","bairro":"Sé","localidade":"São Paulo","uf":"SP"}`))
	}))
	defer provider.Close()

	oldEndpoints := models.Endpoints
	models.Endpoints = []models.Endpoint{
		{Method: models.MethodGet, Source: models.SourceViaCep, URL: provider.URL + "/%s"},
	}
	t.Cleanup(func() {
		models.Endpoints = oldEndpoints
	})

	oldCacheEnabled := config.CacheEnabled
	config.CacheEnabled = true
	t.Cleanup(func() {
		config.CacheEnabled = oldCacheEnabled
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
	failProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		failCalls.Add(1)
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"error":"upstream"}`))
	}))
	defer failProvider.Close()

	var successCalls atomic.Int32
	successProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		successCalls.Add(1)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"cep":"01001-000","logradouro":"Rua Fallback","bairro":"Centro","localidade":"Sao Paulo","uf":"SP"}`))
	}))
	defer successProvider.Close()

	oldEndpoints := models.Endpoints
	models.Endpoints = []models.Endpoint{
		{Method: models.MethodGet, Source: models.SourceViaCep, URL: failProvider.URL + "/%s"},
		{Method: models.MethodGet, Source: models.SourceViaCep, URL: successProvider.URL + "/%s"},
	}
	t.Cleanup(func() {
		models.Endpoints = oldEndpoints
	})

	oldCacheEnabled := config.CacheEnabled
	config.CacheEnabled = false
	t.Cleanup(func() {
		config.CacheEnabled = oldCacheEnabled
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
	correioProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correioCalls.Add(1)
		if r.Method != http.MethodPost {
			t.Fatalf("correio method = %s, want %s", r.Method, http.MethodPost)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`invalid-xml`))
	}))
	defer correioProvider.Close()

	var viaCalls atomic.Int32
	viaProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		viaCalls.Add(1)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"cep":"01001-000","logradouro":"Rua ViaCEP","bairro":"Centro","localidade":"Sao Paulo","uf":"SP"}`))
	}))
	defer viaProvider.Close()

	oldEndpoints := models.Endpoints
	models.Endpoints = []models.Endpoint{
		{
			Method: models.MethodPost,
			Source: models.SourceCorreio,
			URL:    correioProvider.URL,
			Body:   models.PayloadCorreio,
		},
		{Method: models.MethodGet, Source: models.SourceViaCep, URL: viaProvider.URL + "/%s"},
	}
	t.Cleanup(func() {
		models.Endpoints = oldEndpoints
	})

	oldCacheEnabled := config.CacheEnabled
	config.CacheEnabled = false
	t.Cleanup(func() {
		config.CacheEnabled = oldCacheEnabled
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
