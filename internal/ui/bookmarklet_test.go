package ui

import (

	"net/http/httptest"
	"strings"
	"testing"
)

func TestBookmarkletAutoRead(t *testing.T) {
	for _, needle := range []string{"new URLSearchParams(window.location.search).get('url')", "form.requestSubmit()", "ScraperGo · v" + "0.2.85"} {
		if !strings.Contains(Render("0.2.85","Build 000000"), needle) {
			t.Fatalf("missing UI feature: %s", needle)
		}
	}
}

func TestIndexKeepsManualForm(t *testing.T) {
	r := httptest.NewRequest("GET", "/?url=https%3A%2F%2Fexample.com%2Fnews", nil)
	if r.URL.Query().Get("url") != "https://example.com/news" {
		t.Fatal("query decoding failed")
	}
	if !strings.Contains(Template(), `id="form"`) {
		t.Fatal("manual entry disappeared")
	}
}
