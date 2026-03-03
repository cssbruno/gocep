package models

import "testing"

func TestSetGetEndpointsClone(t *testing.T) {
	original := GetEndpoints()
	t.Cleanup(func() {
		SetEndpoints(original)
	})

	input := []Endpoint{
		{Method: MethodGet, Source: SourceViaCep, URL: "https://example.com/%s"},
	}
	SetEndpoints(input)

	input[0].URL = "https://mutated.com/%s"
	got := GetEndpoints()
	if got[0].URL != "https://example.com/%s" {
		t.Fatalf("stored endpoint changed by caller mutation: got %q", got[0].URL)
	}

	got[0].URL = "https://mutated-again.com/%s"
	fresh := GetEndpoints()
	if fresh[0].URL != "https://example.com/%s" {
		t.Fatalf("stored endpoint changed by returned-slice mutation: got %q", fresh[0].URL)
	}
}
