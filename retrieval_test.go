package main

import (
	"net/http"
	"testing"
)

func TestBrowserHeaders(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "https://example.com/", nil)
	if err != nil {
		t.Fatal(err)
	}
	setBrowserHeaders(req)
	for _, k := range []string{"User-Agent", "Accept", "Accept-Language", "Sec-Fetch-Mode"} {
		if req.Header.Get(k) == "" {
			t.Errorf("missing header %q", k)
		}
	}
}
