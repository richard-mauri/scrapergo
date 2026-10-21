package main

import (
	"net/http/httptest"
	"testing"
)

func TestArticleImageRejectsInvalidInput(t *testing.T) {
	cases := []struct {
		method, url string
		want        int
	}{
		{"POST", "/api/article-image?url=https://example.com/x", 405},
		{"GET", "/api/article-image", 422},
		{"GET", "/api/article-image?url=http://127.0.0.1/private", 422},
		{"GET", "/api/article-image?url=file:///etc/passwd", 422},
	}
	for _, tc := range cases {
		w := httptest.NewRecorder()
		articleImage(w, httptest.NewRequest(tc.method, tc.url, nil))
		if w.Code != tc.want {
			t.Errorf("%s %s: got %d want %d", tc.method, tc.url, w.Code, tc.want)
		}
	}
}
func TestSelectedImageRejectsNonPublicURL(t *testing.T) {
	if _, _, err := retrieveSelectedImage(httptest.NewRequest("GET", "/", nil).Context(), "http://127.0.0.1/image.jpg", "https://example.com/article"); err == nil {
		t.Fatal("private image URL accepted")
	}
}
