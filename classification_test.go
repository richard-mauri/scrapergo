package main

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClassifyRejectsBadInput(t *testing.T) {
	for _, tc := range []struct {
		method, body string
		want         int
	}{{"GET", "", 405}, {"POST", "{}", 400}, {"POST", `{"url":"http://127.0.0.1/private"}`, 422}} {
		r := httptest.NewRequest(tc.method, "/api/classify", strings.NewReader(tc.body))
		w := httptest.NewRecorder()
		classifyArticle(w, r)
		if w.Code != tc.want {
			t.Fatalf("%s: got %d want %d", tc.body, w.Code, tc.want)
		}
	}
}

// A temporary publisher failure never qualifies an article for display.
func TestClassificationUnavailableIsNotReadable(t *testing.T) {
	result := classificationResult{Classification: "unavailable", Reason: "publisher_unavailable"}
	if result.ReadableArticle {
		t.Fatal("unavailable publisher cannot be readable")
	}
}
