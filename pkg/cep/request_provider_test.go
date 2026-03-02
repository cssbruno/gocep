package cep

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cssbruno/gocep/models"
)

// go test -run ^TestRequestProvider$ -v
func TestRequestProvider(t *testing.T) {
	tests := []struct {
		name         string
		cep          string
		source       string
		statusCode   int
		responseBody string
		want         string
		wantPath     string
		wantResult   bool
	}{
		{
			name:         "cdnapicep_success",
			cep:          "01001000",
			source:       models.SourceCdnApiCep,
			statusCode:   http.StatusOK,
			responseBody: `{"status":200,"code":"01001-000","state":"SP","city":"São Paulo","district":"Sé","address":"Praça da Sé - lado ímpar"}`,
			want:         `{"cidade":"São Paulo","uf":"SP","logradouro":"Praça da Sé - lado ímpar","bairro":"Sé"}`,
			wantPath:     "/01001-000",
			wantResult:   true,
		},
		{
			name:         "githubjeffotoni_success",
			cep:          "01001000",
			source:       models.SourceGitHubJeffotoni,
			statusCode:   http.StatusOK,
			responseBody: `{"cep":"01001-000","logradouro":"da Sé","bairro":"Sé","uf":"SP","cidade":"São Paulo"}`,
			want:         `{"cidade":"São Paulo","uf":"SP","logradouro":"da Sé","bairro":"Sé"}`,
			wantPath:     "/01001000",
			wantResult:   true,
		},
		{
			name:         "viacep_success",
			cep:          "01001000",
			source:       models.SourceViaCep,
			statusCode:   http.StatusOK,
			responseBody: `{"cep":"01001-000","logradouro":"Praça da Sé","bairro":"Sé","localidade":"São Paulo","uf":"SP"}`,
			want:         `{"cidade":"São Paulo","uf":"SP","logradouro":"Praça da Sé","bairro":"Sé"}`,
			wantPath:     "/01001000",
			wantResult:   true,
		},
		{
			name:         "postmon_success",
			cep:          "01001000",
			source:       models.SourcePostmon,
			statusCode:   http.StatusOK,
			responseBody: `{"bairro":"Sé","cidade":"São Paulo","logradouro":"Praça da Sé","estado":"SP"}`,
			want:         `{"cidade":"São Paulo","uf":"SP","logradouro":"Praça da Sé","bairro":"Sé"}`,
			wantPath:     "/01001000",
			wantResult:   true,
		},
		{
			name:         "republicavirtual_success",
			cep:          "01001000",
			source:       models.SourceRepublicaVirtual,
			statusCode:   http.StatusOK,
			responseBody: `{"uf":"SP","cidade":"São Paulo","bairro":"Sé","logradouro":"da Sé"}`,
			want:         `{"cidade":"São Paulo","uf":"SP","logradouro":"da Sé","bairro":"Sé"}`,
			wantPath:     "/01001000",
			wantResult:   true,
		},
		{
			name:         "brasilapi_success",
			cep:          "01001000",
			source:       models.SourceBrasilAPI,
			statusCode:   http.StatusOK,
			responseBody: `{"cep":"01001-000","state":"SP","city":"São Paulo","neighborhood":"Sé","street":"Praça da Sé"}`,
			want:         `{"cidade":"São Paulo","uf":"SP","logradouro":"Praça da Sé","bairro":"Sé"}`,
			wantPath:     "/01001000",
			wantResult:   true,
		},
		{
			name:         "invalid_json_no_result",
			cep:          "01001000",
			source:       models.SourceViaCep,
			statusCode:   http.StatusOK,
			responseBody: `invalid-json`,
			wantPath:     "/01001000",
			wantResult:   false,
		},
		{
			name:         "non_200_no_result",
			cep:          "01001000",
			source:       models.SourceViaCep,
			statusCode:   http.StatusBadGateway,
			responseBody: `{"error":"upstream"}`,
			wantPath:     "/01001000",
			wantResult:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPath atomic.Value
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath.Store(r.URL.Path)
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			oldClient := httpClient
			httpClient = server.Client()
			t.Cleanup(func() {
				httpClient = oldClient
			})

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			chResult := make(chan Result, 1)
			go requestProvider(ctx, cancel, tt.cep, tt.source, http.MethodGet, server.URL+"/%s", chResult)

			if tt.wantResult {
				select {
				case got := <-chResult:
					if string(got.Body) != tt.want {
						t.Errorf("requestProvider() = %v, want %v", string(got.Body), tt.want)
					}
				case <-time.After(time.Second):
					t.Fatalf("requestProvider() timeout waiting for result")
				}
			} else {
				select {
				case got := <-chResult:
					t.Fatalf("requestProvider() unexpected result: %s", string(got.Body))
				case <-time.After(200 * time.Millisecond):
				}
			}

			path, _ := gotPath.Load().(string)
			if path != tt.wantPath {
				t.Errorf("path = %q, want %q", path, tt.wantPath)
			}
		})
	}
}

// go test -run ^TestAddHyphen$ -v
func TestAddHyphen(t *testing.T) {
	tests := []struct {
		name string
		cep  string
		want string
	}{
		{
			name: "add_hyphen_valid",
			cep:  "08226024",
			want: "08226-024",
		},
		{
			name: "add_hyphen_short",
			cep:  "0822",
			want: "0822",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := addHyphen(tt.cep)
			if got != tt.want {
				t.Errorf("addHyphen() = %v, want %v", got, tt.want)
			}
		})
	}
}
