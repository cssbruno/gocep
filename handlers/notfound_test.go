package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// go test -run ^TestNotFound'$ -v
func TestNotFound(t *testing.T) {
	tests := []struct {
		name string
		want int //statuscode
		body string
	}{
		{"test_not_found_", http.StatusNotFound, `{"error":{"code":"not_found","message":"resource not found"}}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)

			rr := httptest.NewRecorder()

			NotFound(rr, req)

			if rr.Code != tt.want {
				t.Errorf("NotFound() handler returned wrong status code: got %v want %v", rr.Code, tt.want)
			}

			body, _ := io.ReadAll(rr.Body)
			if string(body) != tt.body {
				t.Errorf("NotFound() handler returned wrong body: got %v want %v", string(body), tt.body)
			}
		})
	}
}
