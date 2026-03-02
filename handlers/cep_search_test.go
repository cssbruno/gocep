package handlers

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cssbruno/gocep/config"
	"github.com/cssbruno/gocep/models"
	"github.com/cssbruno/gocep/service/gocache"
)

// go test -run ^TestSearchCep$ -v
func TestSearchCep(t *testing.T) {
	oldCacheEnable := config.CacheEnabled
	config.CacheEnabled = true
	t.Cleanup(func() {
		config.CacheEnabled = oldCacheEnable
	})

	gocache.SetTTL("08226021",
		`{"cidade":"São Paulo","uf":"SP","logradouro":"Rua Esperança","bairro":"Cidade Antônio Estevão de Carvalho"}`,
		time.Duration(config.TTLCache)*time.Second)
	gocache.SetTTL("00000000", config.JsonDefault, time.Duration(config.TTLCache)*time.Second)

	type args struct {
		method string
		ctype  string
		Header map[string]string
		url    string
	}
	tests := []struct {
		name     string
		args     args
		want     int //statuscode
		wantBody string
	}{
		// [GET] /v1/cep/xxxx
		{"invalid_cep_short", args{"GET", "application/json", nil, "/v1/cep/0"}, 400, `{"error":{"code":"invalid_cep","message":"cep must be in 00000000 or 00000-000 format"}}`},
		{"valid_cep_with_hyphen", args{"GET", "application/json", nil, "/v1/cep/08226-021"}, 200, `{"cidade":"São Paulo","uf":"SP","logradouro":"Rua Esperança","bairro":"Cidade Antônio Estevão de Carvalho"}`},
		{"valid_cep", args{"GET", "application/json", nil, "/v1/cep/08226021"}, 200, `{"cidade":"São Paulo","uf":"SP","logradouro":"Rua Esperança","bairro":"Cidade Antônio Estevão de Carvalho"}`},
		{"method_not_allowed", args{"POST", "application/json", nil, "/v1/cep/08226021"}, 405, `{"error":{"code":"method_not_allowed","message":"method not allowed"}}`},
		{"no_content", args{"GET", "application/json", nil, "/v1/cep/00000000"}, 204, ``},
		{"invalid_endpoint", args{"GET", "application/json", nil, "/v1/cep/00000000/"}, 302, `{"error":{"code":"invalid_endpoint","message":"invalid endpoint"}}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tt.args.method, tt.args.url, nil)
			req.Header.Add("Content-type", tt.args.ctype)
			for key, val := range tt.args.Header {
				req.Header.Add(key, val)
			}
			mux := testMux()
			mux.ServeHTTP(w, req)
			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.want {
				t.Errorf("SearchCep() out status = %v, want status %v", resp.StatusCode, tt.want)
				return
			}
			body, _ := io.ReadAll(resp.Body)
			if string(body) != tt.wantBody {
				t.Errorf("SearchCep() body = %v, want body %v", string(body), tt.wantBody)
			}
		})
	}
}

func TestSearchCep_SearchErrorPath(t *testing.T) {
	oldSearch := searchCEP
	searchCEP = func(string) (string, models.CEPAddress, error) {
		return "", models.CEPAddress{}, errors.New("forced error")
	}
	t.Cleanup(func() {
		searchCEP = oldSearch
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/cep/01001000", nil)
	rr := httptest.NewRecorder()

	mux := testMux()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
	body, _ := io.ReadAll(rr.Body)
	want := `{"error":{"code":"search_error","message":"failed to search cep"}}`
	if string(body) != want {
		t.Fatalf("body = %s, want %s", string(body), want)
	}
}
