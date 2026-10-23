package main

import (
 "clean-article-reader/internal/imagedelivery"
 "clean-article-reader/internal/httpapi"
 "clean-article-reader/internal/fetcher"
 "net/http"
 "net/http/httptest"
 "strings"
 "testing"
)

func TestThumbnailMethodRejectionStillWorks(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/thumbnail", strings.NewReader(""))
	w := httptest.NewRecorder()
	httpapi.NewThumbnailHandler(fetcher.ValidatedURL,imagedelivery.Handler{Metadata:newArticleMetadataHandler(newArticleReadHandler())}).ServeHTTP(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status %d", w.Code)
	}
}
