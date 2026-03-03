package cep

import (
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cssbruno/gocep/models"
	"github.com/cssbruno/gocep/service/gocache"
)

// go test -run ^TestSearch$ -v
func TestSearch(t *testing.T) {
	useTestOptions(t, func(o *Options) {
		o.CacheEnabled = true
	})
	useTestCacheProvider(t)

	gocache.SetTTL("08226024",
		`{"cidade":"São Paulo","uf":"SP","logradouro":"Rua Esperança","bairro":"Cidade Antônio Estevão de Carvalho"}`,
		GetOptions().CacheTTL)
	gocache.SetTTL("01001000", `{"cidade":"São Paulo","uf":"SP","logradouro":"da Sé","bairro":"Sé"}`, GetOptions().CacheTTL)

	type args struct {
		cep string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantCep bool
		wantErr bool
	}{
		{
			name:    "test_search_1",
			args:    args{"08226024"},
			want:    `{"cidade":"São Paulo","uf":"SP","logradouro":"Rua Esperança","bairro":"Cidade Antônio Estevão de Carvalho"}`,
			wantCep: true,
			wantErr: false, // (err != nil) => false
		},
		{
			name:    "test_search_2",
			args:    args{"01001000"},
			want:    `{"cidade":"São Paulo","uf":"SP","logradouro":"da Sé","bairro":"Sé"}`,
			wantCep: true,
			wantErr: false,
		},
		{
			name:    "test_search_hyphenated_cep",
			args:    args{"01001-000"},
			want:    `{"cidade":"São Paulo","uf":"SP","logradouro":"da Sé","bairro":"Sé"}`,
			wantCep: true,
			wantErr: false,
		},
		{
			name:    "test_search_3",
			args:    args{""},
			want:    `{"cidade":"","uf":"","logradouro":"","bairro":""}`,
			wantCep: false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, address, err := Search(tt.args.cep)
			if (err != nil) != tt.wantErr {
				t.Errorf("Search() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Search() = %v, want %v", got, tt.want)
			}

			if ValidCEP(address) != tt.wantCep {
				t.Errorf("ValidCEP() = %v, want %v", ValidCEP(address), tt.wantCep)
			}
		})
	}
}

// go test -run ^TestValidCEP$ -v
func TestValidCEP(t *testing.T) {
	type args struct {
		address models.CEPAddress
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test_valid_cep_",
			args: args{models.CEPAddress{City: "São Paulo", StateCode: "SP", Street: "Rua Esperança", Neighborhood: "Cidade Antônio Estevão de Carvalho"}},
			want: true,
		},
		{
			name: "test_valid_cep_",
			args: args{models.CEPAddress{}},
			want: false,
		},
		{
			name: "test_partial_cep_is_invalid",
			args: args{models.CEPAddress{City: "São Paulo"}},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidCEP(tt.args.address)
			if got != tt.want {
				t.Errorf("ValidCEP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkSearchCacheHit(b *testing.B) {
	SetOptions(Options{
		DefaultJSON:     `{"cidade":"","uf":"","logradouro":"","bairro":""}`,
		CacheEnabled:    true,
		CacheTTL:        48 * time.Hour,
		SearchTimeout:   15 * time.Second,
		MaxProviderBody: 1 << 20,
	})
	gocache.SetProvider(&testCacheProvider{})
	defer gocache.SetProvider(nil)

	cep := "08226024"
	payload := `{"cidade":"São Paulo","uf":"SP","logradouro":"Rua Esperança","bairro":"Cidade Antônio Estevão de Carvalho"}`
	address := models.CEPAddress{
		City:         "São Paulo",
		StateCode:    "SP",
		Street:       "Rua Esperança",
		Neighborhood: "Cidade Antônio Estevão de Carvalho",
	}
	gocache.SetAnyTTL(cep, cachedResult{JSON: payload, Address: address}, GetOptions().CacheTTL)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := Search(cep)
		if err != nil {
			b.Fatalf("Search() error = %v", err)
		}
	}
}

func TestSearchTypedCacheHit(t *testing.T) {
	useTestOptions(t, func(o *Options) {
		o.CacheEnabled = true
	})
	useTestCacheProvider(t)

	cepCode := "99999999"
	expectedBody := `{"cidade":"São Paulo","uf":"SP","logradouro":"Rua A","bairro":"Centro"}`
	expectedAddress := models.CEPAddress{
		City:         "São Paulo",
		StateCode:    "SP",
		Street:       "Rua A",
		Neighborhood: "Centro",
	}
	_ = gocache.SetAnyTTL(cepCode, cachedResult{
		JSON:    expectedBody,
		Address: expectedAddress,
	}, time.Minute)

	gotBody, gotAddress, err := Search(cepCode)
	if err != nil {
		t.Fatalf("Search() error = %v, want nil", err)
	}
	if gotBody != expectedBody {
		t.Fatalf("Search() body = %s, want %s", gotBody, expectedBody)
	}
	if gotAddress != expectedAddress {
		t.Fatalf("Search() address = %+v, want %+v", gotAddress, expectedAddress)
	}
}

func TestSearchStringCacheRehydratesTypedCache(t *testing.T) {
	useTestOptions(t, func(o *Options) {
		o.CacheEnabled = true
	})
	useTestCacheProvider(t)

	cepCode := "88888888"
	payload := `{"cidade":"Rio de Janeiro","uf":"RJ","logradouro":"Rua B","bairro":"Centro"}`
	if ok := gocache.SetTTL(cepCode, payload, time.Minute); !ok {
		t.Fatalf("SetTTL() = false, want true")
	}

	gotBody, gotAddress, err := Search(cepCode)
	if err != nil {
		t.Fatalf("Search() error = %v, want nil", err)
	}
	if gotBody != payload {
		t.Fatalf("Search() body = %s, want %s", gotBody, payload)
	}
	if gotAddress.City != "Rio de Janeiro" || gotAddress.StateCode != "RJ" || gotAddress.Street != "Rua B" || gotAddress.Neighborhood != "Centro" {
		t.Fatalf("Search() address = %+v, unexpected", gotAddress)
	}

	typed, found := gocache.GetAny(cepCode)
	if !found {
		t.Fatalf("GetAny() found = false, want true")
	}
	if _, ok := typed.(cachedResult); !ok {
		t.Fatalf("GetAny() type = %T, want cachedResult", typed)
	}
}

func TestSearchTimeoutReturnsDefault(t *testing.T) {
	useTestOptions(t, func(o *Options) {
		o.SearchTimeout = 0
	})

	useTestEndpoints(t, []models.Endpoint{
		{Method: models.MethodGet, Source: models.SourceViaCep, URL: "\n"},
	})

	gotBody, gotAddress, err := Search("12345678")
	if err != nil {
		t.Fatalf("Search() error = %v, want nil", err)
	}
	if gotBody != GetOptions().DefaultJSON {
		t.Fatalf("Search() body = %s, want %s", gotBody, GetOptions().DefaultJSON)
	}
	if gotAddress != (models.CEPAddress{}) {
		t.Fatalf("Search() address = %+v, want empty", gotAddress)
	}
}

func TestSearchNoEndpointsReturnsDefault(t *testing.T) {
	useTestOptions(t, func(o *Options) {
		o.CacheEnabled = false
		o.SearchTimeout = 5 * time.Second
	})

	useTestEndpoints(t, nil)

	start := time.Now()
	gotBody, gotAddress, err := Search("12345678")
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Search() error = %v, want nil", err)
	}
	if gotBody != GetOptions().DefaultJSON {
		t.Fatalf("Search() body = %s, want %s", gotBody, GetOptions().DefaultJSON)
	}
	if gotAddress != (models.CEPAddress{}) {
		t.Fatalf("Search() address = %+v, want empty", gotAddress)
	}
	if elapsed > 250*time.Millisecond {
		t.Fatalf("Search() took %s with no endpoints, expected fast return", elapsed)
	}
}

func TestSearchInvalidCEPSkipsProviderRequests(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"cep":"01001-000","logradouro":"Praça da Sé","bairro":"Sé","localidade":"São Paulo","uf":"SP"}`)
	}))
	defer server.Close()
	useServerHTTPClient(t, server)

	useTestOptions(t, func(o *Options) {
		o.CacheEnabled = false
		o.SearchTimeout = 5 * time.Second
	})

	useTestEndpoints(t, []models.Endpoint{
		{Method: models.MethodGet, Source: models.SourceViaCep, URL: server.URL + "/%s"},
	})

	gotBody, gotAddress, err := Search("abc")
	if err != nil {
		t.Fatalf("Search() error = %v, want nil", err)
	}
	if gotBody != GetOptions().DefaultJSON {
		t.Fatalf("Search() body = %s, want %s", gotBody, GetOptions().DefaultJSON)
	}
	if gotAddress != (models.CEPAddress{}) {
		t.Fatalf("Search() address = %+v, want empty", gotAddress)
	}
	if calls.Load() != 0 {
		t.Fatalf("provider calls = %d, want 0", calls.Load())
	}
}

func TestSearchAllProvidersDoneReturnsBeforeTimeout(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()
	useServerHTTPClient(t, server)

	useTestOptions(t, func(o *Options) {
		o.CacheEnabled = false
		o.SearchTimeout = 5 * time.Second
	})

	useTestEndpoints(t, []models.Endpoint{
		{Method: models.MethodGet, Source: models.SourceViaCep, URL: server.URL + "/%s"},
		{Method: models.MethodGet, Source: models.SourceBrasilAPI, URL: server.URL + "/%s"},
	})

	start := time.Now()
	gotBody, gotAddress, err := Search("01001000")
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Search() error = %v, want nil", err)
	}
	if gotBody != GetOptions().DefaultJSON {
		t.Fatalf("Search() body = %s, want %s", gotBody, GetOptions().DefaultJSON)
	}
	if gotAddress != (models.CEPAddress{}) {
		t.Fatalf("Search() address = %+v, want empty", gotAddress)
	}
	if elapsed > time.Second {
		t.Fatalf("Search() took %s, expected to return before timeout", elapsed)
	}
}

func TestSearch_CorreioEndpointBranch(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want %s", r.Method, http.MethodPost)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `<Envelope><Body><consultaCEPResponse><return><bairro>Sé</bairro><cidade>São Paulo</cidade><end>Praça da Sé</end><uf>SP</uf></return></consultaCEPResponse></Body></Envelope>`)
	}))
	defer server.Close()
	useServerHTTPClient(t, server)

	useTestOptions(t, func(o *Options) {
		o.CacheEnabled = false
	})

	useTestEndpoints(t, []models.Endpoint{
		{
			Method: models.MethodPost,
			Source: models.SourceCorreio,
			URL:    server.URL,
			Body:   models.PayloadCorreio,
		},
	})

	gotBody, gotAddress, err := Search("01001000")
	if err != nil {
		t.Fatalf("Search() error = %v, want nil", err)
	}
	wantBody := `{"cidade":"São Paulo","uf":"SP","logradouro":"Praça da Sé","bairro":"Sé"}`
	if gotBody != wantBody {
		t.Fatalf("Search() body = %s, want %s", gotBody, wantBody)
	}
	if gotAddress.City != "São Paulo" || gotAddress.StateCode != "SP" || gotAddress.Street != "Praça da Sé" || gotAddress.Neighborhood != "Sé" {
		t.Fatalf("Search() address unexpected: %+v", gotAddress)
	}
}
