package cep

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cssbruno/gocep/v2/models"
	"github.com/cssbruno/gocep/v2/service/gocache"
)

func testClientOptions() Options {
	return Options{
		DefaultJSON:     `{"cep":"","cidade":"","uf":"","logradouro":"","bairro":""}`,
		CacheEnabled:    false,
		CacheTTL:        time.Minute,
		SearchTimeout:   2 * time.Second,
		MaxProviderBody: 1 << 20,
	}
}

func TestClientSearchContextRespectsCallerDeadline(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	client := NewClient(
		WithHTTPClient(server.Client()),
		WithEndpoints([]models.Endpoint{
			{Method: models.MethodGet, Source: models.SourceViaCep, URL: server.URL + "/%s"},
		}),
		WithOptions(testClientOptions()),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	gotBody, gotAddress, err := client.SearchContext(ctx, "01001000")
	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("SearchContext() error = %v, want %v", err, ErrTimeout)
	}
	if gotBody != client.Options().DefaultJSON {
		t.Fatalf("SearchContext() body = %s, want %s", gotBody, client.Options().DefaultJSON)
	}
	if gotAddress != (models.CEPAddress{}) {
		t.Fatalf("SearchContext() address = %+v, want empty", gotAddress)
	}
}

func TestClientProviderPolicyOrderedFallbackWithPerProviderTimeout(t *testing.T) {
	const cepCode = "01001000"
	const expectedBody = `{"cep":"01001-000","cidade":"Sao Paulo","uf":"SP","logradouro":"Rua Rapida","bairro":"Centro"}`

	var slowCalls atomic.Int32
	var fastCalls atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/slow/" + cepCode:
			slowCalls.Add(1)
			<-r.Context().Done()
		case "/fast/" + cepCode:
			fastCalls.Add(1)
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, `{"cep":"01001-000","state":"SP","city":"Sao Paulo","neighborhood":"Centro","street":"Rua Rapida"}`)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	options := testClientOptions()
	options.SearchTimeout = time.Second
	client := NewClient(
		WithHTTPClient(server.Client()),
		WithOptions(options),
		WithEndpoints([]models.Endpoint{
			{Method: models.MethodGet, Source: models.SourceViaCep, URL: server.URL + "/slow/%s"},
			{Method: models.MethodGet, Source: models.SourceBrasilAPI, URL: server.URL + "/fast/%s"},
		}),
		WithProviderPolicy(ProviderPolicy{
			Strategy: SearchStrategyOrderedFallback,
			SourceTimeouts: map[string]time.Duration{
				models.SourceViaCep: 20 * time.Millisecond,
			},
		}),
	)

	gotBody, gotAddress, err := client.Search(cepCode)

	if err != nil {
		t.Fatalf("Search() error = %v, want nil", err)
	}
	if gotBody != expectedBody {
		t.Fatalf("Search() body = %s, want %s", gotBody, expectedBody)
	}
	if !ValidCEP(gotAddress) {
		t.Fatalf("Search() address = %+v, want valid", gotAddress)
	}
	if slowCalls.Load() != 1 || fastCalls.Load() != 1 {
		t.Fatalf("calls slow=%d fast=%d, want 1/1", slowCalls.Load(), fastCalls.Load())
	}
}

func TestClientProviderPolicyPreferredAndDisabled(t *testing.T) {
	const cepCode = "02020020"
	const viaBody = `{"cep":"02020-020","logradouro":"Rua Via","bairro":"Centro","localidade":"Sao Paulo","uf":"SP"}`
	const brBody = `{"cep":"02020-020","state":"SP","city":"Sao Paulo","neighborhood":"Centro","street":"Rua Brasil"}`
	const expectedPreferred = `{"cep":"02020-020","cidade":"Sao Paulo","uf":"SP","logradouro":"Rua Brasil","bairro":"Centro"}`

	var viaCalls atomic.Int32
	var brCalls atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/via/" + cepCode:
			viaCalls.Add(1)
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, viaBody)
		case "/br/" + cepCode:
			brCalls.Add(1)
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, brBody)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	endpoints := []models.Endpoint{
		{Method: models.MethodGet, Source: models.SourceViaCep, URL: server.URL + "/via/%s"},
		{Method: models.MethodGet, Source: models.SourceBrasilAPI, URL: server.URL + "/br/%s"},
	}

	preferredClient := NewClient(
		WithHTTPClient(server.Client()),
		WithEndpoints(endpoints),
		WithOptions(testClientOptions()),
		WithProviderPolicy(ProviderPolicy{
			Strategy:         SearchStrategyOrderedFallback,
			PreferredSources: []string{models.SourceBrasilAPI},
		}),
	)
	gotBody, _, err := preferredClient.Search(cepCode)
	if err != nil {
		t.Fatalf("preferred Search() error = %v", err)
	}
	if gotBody != expectedPreferred {
		t.Fatalf("preferred Search() body = %s, want %s", gotBody, expectedPreferred)
	}
	if viaCalls.Load() != 0 || brCalls.Load() != 1 {
		t.Fatalf("preferred calls via=%d br=%d, want 0/1", viaCalls.Load(), brCalls.Load())
	}

	viaCalls.Store(0)
	brCalls.Store(0)

	disabledClient := NewClient(
		WithHTTPClient(server.Client()),
		WithEndpoints(endpoints),
		WithOptions(testClientOptions()),
		WithProviderPolicy(ProviderPolicy{
			Strategy: SearchStrategyOrderedFallback,
			DisabledSources: map[string]bool{
				models.SourceBrasilAPI: true,
			},
		}),
	)
	gotBody, _, err = disabledClient.Search(cepCode)
	if err != nil {
		t.Fatalf("disabled Search() error = %v", err)
	}
	if gotBody == expectedPreferred {
		t.Fatalf("disabled Search() unexpectedly used disabled provider")
	}
	if viaCalls.Load() != 1 || brCalls.Load() != 0 {
		t.Fatalf("disabled calls via=%d br=%d, want 1/0", viaCalls.Load(), brCalls.Load())
	}
}

func TestClientAllProviderTimeoutsReturnErrTimeout(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	options := testClientOptions()
	options.SearchTimeout = 2 * time.Second

	client := NewClient(
		WithHTTPClient(server.Client()),
		WithOptions(options),
		WithEndpoints([]models.Endpoint{
			{Method: models.MethodGet, Source: models.SourceViaCep, URL: server.URL + "/%s"},
		}),
		WithProviderPolicy(ProviderPolicy{
			Strategy: SearchStrategyOrderedFallback,
			SourceTimeouts: map[string]time.Duration{
				models.SourceViaCep: 15 * time.Millisecond,
			},
		}),
	)

	gotBody, gotAddress, err := client.Search("01001000")
	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("Search() error = %v, want %v", err, ErrTimeout)
	}
	if gotBody != options.DefaultJSON {
		t.Fatalf("Search() body = %s, want %s", gotBody, options.DefaultJSON)
	}
	if gotAddress != (models.CEPAddress{}) {
		t.Fatalf("Search() address = %+v, want empty", gotAddress)
	}
}

func TestClientHooksCacheAndProviderEvents(t *testing.T) {
	cepCode := "88888888"
	cache := &testCacheProvider{}
	_ = cache.SetAnyTTL(cepCode, cachedResult{
		JSON: `{"cidade":"Rio de Janeiro","uf":"RJ","logradouro":"Rua B","bairro":"Centro"}`,
		Address: models.CEPAddress{
			CEP:          "88888-888",
			City:         "Rio de Janeiro",
			StateCode:    "RJ",
			Street:       "Rua B",
			Neighborhood: "Centro",
		},
	}, time.Minute)

	var cacheHits atomic.Int32
	var providerEvents atomic.Int32
	opts := testClientOptions()
	opts.CacheEnabled = true
	client := NewClient(
		WithOptions(opts),
		WithCacheProvider(cache),
		WithHooks(Hooks{
			OnCacheEvent: func(event CacheEvent) {
				if event.Operation == "read" && event.Hit {
					cacheHits.Add(1)
				}
			},
			OnProviderEvent: func(ProviderResultEvent) {
				providerEvents.Add(1)
			},
		}),
	)

	gotBody, gotAddress, err := client.Search(cepCode)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if gotBody == "" || !ValidCEP(gotAddress) {
		t.Fatalf("Search() unexpected result body=%q address=%+v", gotBody, gotAddress)
	}
	if cacheHits.Load() == 0 {
		t.Fatalf("expected cache hook to report read hit")
	}
	if providerEvents.Load() != 0 {
		t.Fatalf("expected no provider hook events on cache hit, got %d", providerEvents.Load())
	}
}

func TestClientConfigurationIsolation(t *testing.T) {
	old := GetOptions()
	t.Cleanup(func() {
		SetOptions(old)
	})

	cli := NewClient(WithOptions(Options{
		DefaultJSON:     "{}",
		CacheEnabled:    false,
		CacheTTL:        time.Minute,
		SearchTimeout:   time.Second,
		MaxProviderBody: 128,
	}))
	cli.SetOptions(Options{
		DefaultJSON:     "{}",
		CacheEnabled:    false,
		CacheTTL:        2 * time.Minute,
		SearchTimeout:   3 * time.Second,
		MaxProviderBody: 256,
	})

	global := GetOptions()
	if global.SearchTimeout != old.SearchTimeout || global.MaxProviderBody != old.MaxProviderBody {
		t.Fatalf("global options changed unexpectedly: got %+v want %+v", global, old)
	}
}

func TestClientSearchContextDoesNotShareCancelableContext(t *testing.T) {
	const cepCode = "01001000"
	const expectedBody = `{"cep":"01001-000","cidade":"Sao Paulo","uf":"SP","logradouro":"Rua Rapida","bairro":"Centro"}`

	firstStarted := make(chan struct{})
	var calls atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			close(firstStarted)
			<-r.Context().Done()
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"cep":"01001-000","logradouro":"Rua Rapida","bairro":"Centro","localidade":"Sao Paulo","uf":"SP"}`)
	}))
	defer server.Close()

	client := NewClient(
		WithHTTPClient(server.Client()),
		WithOptions(testClientOptions()),
		WithEndpoints([]models.Endpoint{
			{Method: models.MethodGet, Source: models.SourceViaCep, URL: server.URL + "/%s"},
		}),
	)

	firstCtx, firstCancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer firstCancel()
	firstErr := make(chan error, 1)
	go func() {
		_, _, err := client.SearchContext(firstCtx, cepCode)
		firstErr <- err
	}()

	waitCtx, waitCancel := context.WithTimeout(context.Background(), time.Second)
	defer waitCancel()
	select {
	case <-firstStarted:
	case <-waitCtx.Done():
		t.Fatalf("timed out waiting for first request")
	}

	secondCtx, secondCancel := context.WithTimeout(context.Background(), time.Second)
	defer secondCancel()
	gotBody, gotAddress, err := client.SearchContext(secondCtx, cepCode)
	if err != nil {
		t.Fatalf("second SearchContext() error = %v, want nil", err)
	}
	if gotBody != expectedBody {
		t.Fatalf("second SearchContext() body = %s, want %s", gotBody, expectedBody)
	}
	if !ValidCEP(gotAddress) {
		t.Fatalf("second SearchContext() address = %+v, want valid", gotAddress)
	}

	if err := <-firstErr; !errors.Is(err, ErrTimeout) {
		t.Fatalf("first SearchContext() error = %v, want %v", err, ErrTimeout)
	}
	if calls.Load() != 2 {
		t.Fatalf("provider calls = %d, want 2", calls.Load())
	}
}

func TestClientRejectsOversizedProviderBody(t *testing.T) {
	tests := []struct {
		name     string
		endpoint models.Endpoint
	}{
		{
			name: "json",
			endpoint: models.Endpoint{
				Method: models.MethodGet,
				Source: models.SourceViaCep,
			},
		},
		{
			name: "correio",
			endpoint: models.Endpoint{
				Method: models.MethodPost,
				Source: models.SourceCorreio,
				Body:   models.PayloadCorreio,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = io.WriteString(w, strings.Repeat("x", 32))
			}))
			defer server.Close()

			opts := testClientOptions()
			opts.CacheEnabled = true
			opts.MaxProviderBody = 8
			cache := &testCacheProvider{}
			tt.endpoint.URL = server.URL + "/%s"
			if tt.endpoint.Source == models.SourceCorreio {
				tt.endpoint.URL = server.URL
			}

			client := NewClient(
				WithHTTPClient(server.Client()),
				WithOptions(opts),
				WithCacheProvider(cache),
				WithEndpoints([]models.Endpoint{tt.endpoint}),
			)

			gotBody, gotAddress, err := client.Search("01001000")
			if !errors.Is(err, ErrNotFound) {
				t.Fatalf("Search() error = %v, want %v", err, ErrNotFound)
			}
			if gotBody != opts.DefaultJSON {
				t.Fatalf("Search() body = %s, want %s", gotBody, opts.DefaultJSON)
			}
			if gotAddress != (models.CEPAddress{}) {
				t.Fatalf("Search() address = %+v, want empty", gotAddress)
			}
			if _, found := cache.GetAny("01001000"); found {
				t.Fatalf("oversized response was cached")
			}
		})
	}
}

func TestNewClientDoesNotInheritGlobalCacheProvider(t *testing.T) {
	gocache.SetProvider(&testCacheProvider{})
	t.Cleanup(func() {
		gocache.SetProvider(nil)
	})
	_ = gocache.SetAnyTTL("01001000", cachedResult{
		JSON: `{"cep":"01001-000","cidade":"Sao Paulo","uf":"SP","logradouro":"Rua Cache","bairro":"Centro"}`,
		Address: models.CEPAddress{
			CEP:          "01001-000",
			City:         "Sao Paulo",
			StateCode:    "SP",
			Street:       "Rua Cache",
			Neighborhood: "Centro",
		},
	}, time.Minute)

	opts := testClientOptions()
	opts.CacheEnabled = true
	client := NewClient(
		WithOptions(opts),
		WithEndpoints(nil),
	)

	gotBody, gotAddress, err := client.Search("01001000")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Search() error = %v, want %v", err, ErrNotFound)
	}
	if gotBody != opts.DefaultJSON {
		t.Fatalf("Search() body = %s, want %s", gotBody, opts.DefaultJSON)
	}
	if gotAddress != (models.CEPAddress{}) {
		t.Fatalf("Search() address = %+v, want empty", gotAddress)
	}
}
