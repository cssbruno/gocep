package cep

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cssbruno/gocep/models"
)

type closeTracker struct {
	closed bool
}

func (c *closeTracker) Read(_ []byte) (int, error) { return 0, io.EOF }
func (c *closeTracker) Close() error {
	c.closed = true
	return nil
}

func TestExecuteRequest_ErrorReturnsFalse(t *testing.T) {
	oldClient := getHTTPClient()
	SetHTTPClient(&http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			return nil, errors.New("forced do error")
		}),
	})
	t.Cleanup(func() {
		SetHTTPClient(oldClient)
	})

	req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}

	resp, ok := executeRequest(req)
	if ok {
		t.Fatalf("executeRequest() ok = true, want false")
	}
	if resp != nil {
		t.Fatalf("executeRequest() response = %v, want nil", resp)
	}
}

func TestExecuteRequest_Non200ClosesBody(t *testing.T) {
	tracker := &closeTracker{}
	oldClient := getHTTPClient()
	SetHTTPClient(&http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusBadGateway,
				Body:       tracker,
				Header:     make(http.Header),
				Request:    r,
			}, nil
		}),
	})
	t.Cleanup(func() {
		SetHTTPClient(oldClient)
	})

	req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}

	resp, ok := executeRequest(req)
	if ok {
		t.Fatalf("executeRequest() ok = true, want false")
	}
	if resp != nil {
		t.Fatalf("executeRequest() response = %v, want nil", resp)
	}
	if !tracker.closed {
		t.Fatalf("expected response body to be closed on non-200")
	}
}

func TestExecuteRequest_Success(t *testing.T) {
	oldClient := getHTTPClient()
	SetHTTPClient(&http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("ok")),
				Header:     make(http.Header),
				Request:    r,
			}, nil
		}),
	})
	t.Cleanup(func() {
		SetHTTPClient(oldClient)
	})

	req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}

	resp, ok := executeRequest(req)
	if !ok {
		t.Fatalf("executeRequest() ok = false, want true")
	}
	if resp == nil {
		t.Fatalf("executeRequest() response = nil, want non-nil")
	}
	_ = resp.Body.Close()
}

func TestExecuteRequest_NonTLSURLRejected(t *testing.T) {
	oldClient := getHTTPClient()
	SetHTTPClient(&http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			t.Fatalf("transport should not be called for non-TLS URL")
			return nil, nil
		}),
	})
	t.Cleanup(func() {
		SetHTTPClient(oldClient)
	})

	req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}

	resp, ok := executeRequest(req)
	if ok {
		t.Fatalf("executeRequest() ok = true, want false")
	}
	if resp != nil {
		t.Fatalf("executeRequest() response = %v, want nil", resp)
	}
}

func TestExecuteRequest_RedirectErrorReturnsFalse(t *testing.T) {
	redirectServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://example.com/next", http.StatusFound)
	}))
	defer redirectServer.Close()

	oldClient := getHTTPClient()
	client := redirectServer.Client()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return errors.New("stop redirect")
	}
	SetHTTPClient(client)
	t.Cleanup(func() {
		SetHTTPClient(oldClient)
	})

	req, err := http.NewRequest(http.MethodGet, redirectServer.URL, nil)
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}

	resp, ok := executeRequest(req)
	if ok {
		t.Fatalf("executeRequest() ok = true, want false")
	}
	if resp != nil {
		t.Fatalf("executeRequest() response = %v, want nil", resp)
	}
}

func TestSendAddressResult_SendsAndCancels(t *testing.T) {
	oldMarshal := marshalAddressJSON
	marshalAddressJSON = func(v any) ([]byte, error) {
		return oldMarshal(v)
	}
	t.Cleanup(func() {
		marshalAddressJSON = oldMarshal
	})

	ctx := context.Background()
	calledCancel := false
	chResult := make(chan Result, 1)
	address := models.CEPAddress{
		CEP:          "01001-000",
		City:         "Sao Paulo",
		StateCode:    "SP",
		Street:       "Street",
		Neighborhood: "District",
	}

	sendAddressResult(ctx, func() { calledCancel = true }, chResult, address)

	if !calledCancel {
		t.Fatalf("expected cancel to be called after sending result")
	}

	select {
	case got := <-chResult:
		if got.Address != address {
			t.Fatalf("address = %+v, want %+v", got.Address, address)
		}
	default:
		t.Fatalf("expected result to be sent")
	}
}

func TestSendAddressResult_ContextDoneNoSend(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	cancelCtx()

	calledCancel := false
	chResult := make(chan Result)
	address := models.CEPAddress{
		CEP:          "01001-000",
		City:         "Sao Paulo",
		StateCode:    "SP",
		Street:       "Street",
		Neighborhood: "District",
	}

	sendAddressResult(ctx, func() { calledCancel = true }, chResult, address)

	if calledCancel {
		t.Fatalf("cancel should not be called when context is already done")
	}

	select {
	case <-chResult:
		t.Fatalf("did not expect any result to be sent")
	default:
	}
}

func TestSendAddressResult_MarshalErrorNoSend(t *testing.T) {
	oldMarshal := marshalAddressJSON
	marshalAddressJSON = func(any) ([]byte, error) {
		return nil, errors.New("forced marshal error")
	}
	t.Cleanup(func() {
		marshalAddressJSON = oldMarshal
	})

	ctx := context.Background()
	calledCancel := false
	chResult := make(chan Result, 1)
	address := models.CEPAddress{
		CEP:          "01001-000",
		City:         "Sao Paulo",
		StateCode:    "SP",
		Street:       "Street",
		Neighborhood: "District",
	}

	sendAddressResult(ctx, func() { calledCancel = true }, chResult, address)

	if calledCancel {
		t.Fatalf("cancel should not be called on marshal error")
	}

	select {
	case <-chResult:
		t.Fatalf("did not expect any result to be sent")
	default:
	}
}

func TestSendAddressResult_IncompleteAddressNoSend(t *testing.T) {
	ctx := context.Background()
	calledCancel := false
	chResult := make(chan Result, 1)

	sendAddressResult(ctx, func() { calledCancel = true }, chResult, models.CEPAddress{
		City: "Sao Paulo",
	})

	if calledCancel {
		t.Fatalf("cancel should not be called for incomplete address")
	}

	select {
	case <-chResult:
		t.Fatalf("did not expect any result to be sent")
	default:
	}
}
