package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cssbruno/gocep/config"
	"github.com/cssbruno/gocep/models"
	"github.com/cssbruno/gocep/service/gocache"
)

func testMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/cep/{cep...}", SearchCep)
	mux.HandleFunc("/v1/cep", NotFound)
	mux.HandleFunc("/", NotFound)
	return mux
}

func TestSearchCepIntegrationViaCEP(t *testing.T) {
	var calls atomic.Int32
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.Method != http.MethodGet {
			t.Errorf("provider method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/01001000" {
			t.Errorf("provider path = %s, want /01001000", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"cep":"01001-000","logradouro":"Praça da Sé","bairro":"Sé","localidade":"São Paulo","uf":"SP"}`))
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
	config.CacheEnabled = false
	t.Cleanup(func() {
		config.CacheEnabled = oldCacheEnabled
	})

	mux := testMux()
	req := httptest.NewRequest(http.MethodGet, "/v1/cep/01001000", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	body, _ := io.ReadAll(rr.Body)
	wantBody := `{"cidade":"São Paulo","uf":"SP","logradouro":"Praça da Sé","bairro":"Sé"}`
	if string(body) != wantBody {
		t.Fatalf("body = %s, want %s", string(body), wantBody)
	}

	if calls.Load() != 1 {
		t.Fatalf("provider calls = %d, want 1", calls.Load())
	}
}

func TestSearchCepIntegrationCorreio(t *testing.T) {
	var calls atomic.Int32
	var gotMethod atomic.Value
	var gotContentType atomic.Value

	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		gotMethod.Store(r.Method)
		gotContentType.Store(r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<Envelope><Body><consultaCEPResponse><return><bairro>Sé</bairro><cidade>São Paulo</cidade><end>Praça da Sé</end><uf>SP</uf></return></consultaCEPResponse></Body></Envelope>`))
	}))
	defer provider.Close()

	oldEndpoints := models.Endpoints
	models.Endpoints = []models.Endpoint{
		{
			Method: models.MethodPost,
			Source: models.SourceCorreio,
			URL:    provider.URL,
			Body:   models.PayloadCorreio,
		},
	}
	t.Cleanup(func() {
		models.Endpoints = oldEndpoints
	})

	oldCacheEnabled := config.CacheEnabled
	config.CacheEnabled = false
	t.Cleanup(func() {
		config.CacheEnabled = oldCacheEnabled
	})

	mux := testMux()
	req := httptest.NewRequest(http.MethodGet, "/v1/cep/01001000", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	body, _ := io.ReadAll(rr.Body)
	wantBody := `{"cidade":"São Paulo","uf":"SP","logradouro":"Praça da Sé","bairro":"Sé"}`
	if string(body) != wantBody {
		t.Fatalf("body = %s, want %s", string(body), wantBody)
	}

	method, _ := gotMethod.Load().(string)
	if method != http.MethodPost {
		t.Fatalf("provider method = %s, want %s", method, http.MethodPost)
	}

	contentType, _ := gotContentType.Load().(string)
	if contentType != "text/xml; charset=utf-8" {
		t.Fatalf("provider content-type = %s, want text/xml; charset=utf-8", contentType)
	}

	if calls.Load() != 1 {
		t.Fatalf("provider calls = %d, want 1", calls.Load())
	}
}

func TestSearchCepIntegrationUpstreamFailureReturnsNoContent(t *testing.T) {
	var calls atomic.Int32
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"error":"upstream"}`))
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
	config.CacheEnabled = false
	t.Cleanup(func() {
		config.CacheEnabled = oldCacheEnabled
	})

	mux := testMux()
	req := httptest.NewRequest(http.MethodGet, "/v1/cep/01001000", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNoContent)
	}
	if rr.Body.Len() != 0 {
		t.Fatalf("body len = %d, want 0", rr.Body.Len())
	}
	if calls.Load() != 1 {
		t.Fatalf("provider calls = %d, want 1", calls.Load())
	}
}

func TestSearchCepIntegrationCacheHitSkipsProvider(t *testing.T) {
	const cepCode = "55555555"
	const cachedPayload = `{"cidade":"Sao Paulo","uf":"SP","logradouro":"Rua A","bairro":"Centro"}`

	var calls atomic.Int32
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"cep":"55555-555","logradouro":"Rua Provider","bairro":"Centro","localidade":"Sao Paulo","uf":"SP"}`))
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

	if ok := gocache.SetTTL(cepCode, cachedPayload, time.Minute); !ok {
		t.Fatalf("failed to seed cache")
	}

	mux := testMux()
	req := httptest.NewRequest(http.MethodGet, "/v1/cep/"+cepCode, nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	body, _ := io.ReadAll(rr.Body)
	if string(body) != cachedPayload {
		t.Fatalf("body = %s, want %s", string(body), cachedPayload)
	}
	if calls.Load() != 0 {
		t.Fatalf("provider calls = %d, want 0 (cache hit)", calls.Load())
	}
}

func TestSearchCepIntegrationInvalidEndpoint(t *testing.T) {
	mux := testMux()
	req := httptest.NewRequest(http.MethodGet, "/v1/cep/01001000/", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusFound)
	}
	body, _ := io.ReadAll(rr.Body)
	want := `{"error":{"code":"invalid_endpoint","message":"invalid endpoint"}}`
	if string(body) != want {
		t.Fatalf("body = %s, want %s", string(body), want)
	}
}

func TestSearchCepIntegrationNotFoundRoute(t *testing.T) {
	mux := testMux()
	req := httptest.NewRequest(http.MethodGet, "/unknown", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
	body, _ := io.ReadAll(rr.Body)
	want := `{"error":{"code":"not_found","message":"resource not found"}}`
	if string(body) != want {
		t.Fatalf("body = %s, want %s", string(body), want)
	}
}
