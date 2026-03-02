package cep

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

const correioPayloadTemplate = `<x:Envelope xmlns:x="http://schemas.xmlsoap.org/soap/envelope/" xmlns:cli="http://cliente.bean.master.sigep.bsb.correios.com.br/">
	<x:Body>
		<cli:consultaCEP>
			<cep>%s</cep>
		</cli:consultaCEP>
	</x:Body>
</x:Envelope>`

// go test -run ^TestRequestCorreio$ -v
func TestRequestCorreio(t *testing.T) {
	tests := []struct {
		name         string
		endpoint     string
		responseBody string
		want         string
		wantResult   bool
		useServer    bool
	}{
		{
			name:         "success",
			responseBody: `<Envelope><Body><consultaCEPResponse><return><bairro>Sé</bairro><cidade>São Paulo</cidade><end>Praça da Sé</end><uf>SP</uf></return></consultaCEPResponse></Body></Envelope>`,
			want:         `{"cidade":"São Paulo","uf":"SP","logradouro":"Praça da Sé","bairro":"Sé"}`,
			wantResult:   true,
			useServer:    true,
		},
		{
			name:         "invalid_xml_no_result",
			responseBody: `invalid-xml`,
			wantResult:   false,
			useServer:    true,
		},
		{
			name:       "invalid_endpoint_no_result",
			endpoint:   "\n",
			wantResult: false,
			useServer:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotMethod atomic.Value
			var gotContentType atomic.Value

			endpoint := tt.endpoint
			if tt.useServer {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					gotMethod.Store(r.Method)
					gotContentType.Store(r.Header.Get("Content-type"))
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(tt.responseBody))
				}))
				defer server.Close()
				endpoint = server.URL

				oldClient := httpClient
				httpClient = server.Client()
				t.Cleanup(func() {
					httpClient = oldClient
				})
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			chResult := make(chan Result, 1)
			go requestCorreio(ctx, cancel, "01001000", http.MethodPost, endpoint, correioPayloadTemplate, chResult)

			if tt.wantResult {
				select {
				case got := <-chResult:
					if string(got.Body) != tt.want {
						t.Errorf("requestCorreio() = %v, want %v", string(got.Body), tt.want)
					}
				case <-time.After(time.Second):
					t.Fatalf("requestCorreio() timeout waiting for result")
				}
			} else {
				select {
				case got := <-chResult:
					t.Fatalf("requestCorreio() unexpected result: %s", string(got.Body))
				case <-time.After(200 * time.Millisecond):
				}
			}

			if tt.wantResult {
				method, _ := gotMethod.Load().(string)
				if method != http.MethodPost {
					t.Errorf("method = %q, want %q", method, http.MethodPost)
				}
				contentType, _ := gotContentType.Load().(string)
				if contentType != "text/xml; charset=utf-8" {
					t.Errorf("content-type = %q, want %q", contentType, "text/xml; charset=utf-8")
				}
			}
		})
	}
}
