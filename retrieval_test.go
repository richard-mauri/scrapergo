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

func TestReadeckLikeHeaders(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "https://example.com/", nil)
	if err != nil {
		t.Fatal(err)
	}
	setReadeckLikeHeaders(req)
	if req.Header.Get("Sec-CH-UA-Platform") != `"Windows"` {
		t.Fatal("incorrect client hint platform")
	}
	if req.Header.Get("Sec-CH-UA-Mobile") != "?0" {
		t.Fatal("incorrect mobile hint")
	}
	if req.Header.Get("Sec-Fetch-Mode") != "navigate" {
		t.Fatal("incorrect fetch mode")
	}
	if req.Header.Get("User-Agent") == "" {
		t.Fatal("missing user agent")
	}
	if !alternateHeaderRetry(http.StatusForbidden) || alternateHeaderRetry(http.StatusUnauthorized) || alternateHeaderRetry(http.StatusOK) {
		t.Fatal("retry must be limited to explicit HTTP 403")
	}
}
