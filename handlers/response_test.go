package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteAPIError_MarshalFailure(t *testing.T) {
	oldMarshal := marshalAPIError
	marshalAPIError = func(v any) ([]byte, error) {
		return nil, errors.New("forced marshal error")
	}
	t.Cleanup(func() {
		marshalAPIError = oldMarshal
	})

	rr := httptest.NewRecorder()
	writeAPIError(rr, http.StatusBadRequest, "code", "message")

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
	if rr.Body.Len() != 0 {
		t.Fatalf("body len = %d, want 0", rr.Body.Len())
	}
}
