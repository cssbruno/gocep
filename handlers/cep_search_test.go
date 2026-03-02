package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/cssbruno/gocep/config"
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
		{"test_searchcep_", args{"GET", "application/json", nil, "/v1/cep/0"}, 400, `{"error":{"code":"invalid_cep","message":"cep must contain exactly 8 digits"}}`},
		{"test_searchcep_", args{"GET", "application/json", nil, "/v1/cep/08226-021"}, 400, `{"error":{"code":"invalid_cep","message":"cep must contain exactly 8 digits"}}`},
		{"test_searchcep_", args{"GET", "application/json", nil, "/v1/cep/08226021"}, 200, `{"cidade":"São Paulo","uf":"SP","logradouro":"Rua Esperança","bairro":"Cidade Antônio Estevão de Carvalho"}`},
		{"test_searchcep_", args{"POST", "application/json", nil, "/v1/cep/08226021"}, 405, `{"error":{"code":"method_not_allowed","message":"method not allowed"}}`},
		{"test_searchcep_", args{"GET", "application/json", nil, "/v1/cep/00000000"}, 204, ``},
		{"test_searchcep_", args{"GET", "application/json", nil, "/v1/cep/00000000/"}, 302, `{"error":{"code":"invalid_endpoint","message":"invalid endpoint"}}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tt.args.method, tt.args.url, nil)
			req.Header.Add("Content-type", tt.args.ctype)
			for key, val := range tt.args.Header {
				req.Header.Add(key, val)
			}
			mux := http.NewServeMux()
			mux.HandleFunc("/v1/cep/{cep...}", SearchCep)
			mux.HandleFunc("/v1/cep", NotFound)
			mux.HandleFunc("/", NotFound)
			mux.ServeHTTP(w, req)
			resp := w.Result()
			defer resp.Body.Close()

			if !reflect.DeepEqual(resp.StatusCode, tt.want) {
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
