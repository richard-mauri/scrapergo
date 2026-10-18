package main

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBookmarkletAutoRead(t *testing.T) {
	for _, needle := range []string{"new URLSearchParams(window.location.search).get('url')", "form.requestSubmit()", "ScraperGo · v0.2.40"} {
		if !strings.Contains(indexHTML, needle) {
			t.Fatalf("missing UI feature: %s", needle)
		}
	}
}

func TestIndexKeepsManualForm(t *testing.T) {
	r := httptest.NewRequest("GET", "/?url=https%3A%2F%2Fexample.com%2Fnews", nil)
	if r.URL.Query().Get("url") != "https://example.com/news" {
		t.Fatal("query decoding failed")
	}
	if !strings.Contains(indexHTML, `id="form"`) {
		t.Fatal("manual entry disappeared")
	}
}
