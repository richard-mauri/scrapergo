package httpapi_test

import (
 "clean-article-reader/internal/bootstrap"
 "clean-article-reader/internal/imagedelivery"
 "clean-article-reader/internal/httpapi"
 "clean-article-reader/internal/fetcher"
 "net/http/httptest"
 "testing"
)

func TestThumbnailValidation(t *testing.T) {
	for _, url := range []string{"/api/thumbnail?url=https://example.com/article&width=1", "/api/thumbnail?url=https://example.com/article&width=99999"} {
		w := httptest.NewRecorder()
		httpapi.NewThumbnailHandler(fetcher.ValidatedURL,imagedelivery.Handler{Metadata:bootstrap.NewArticleMetadataHandler(bootstrap.NewArticleReadHandler())}).ServeHTTP(w, httptest.NewRequest("GET", url, nil))
		if w.Code != 400 {
			t.Fatalf("%s: %d", url, w.Code)
		}
	}
	w := httptest.NewRecorder()
	httpapi.NewThumbnailHandler(fetcher.ValidatedURL,imagedelivery.Handler{Metadata:bootstrap.NewArticleMetadataHandler(bootstrap.NewArticleReadHandler())}).ServeHTTP(w, httptest.NewRequest("POST", "/api/thumbnail", nil))
	if w.Code != 405 {
		t.Fatalf("POST status %d", w.Code)
	}
	w = httptest.NewRecorder()
	httpapi.NewThumbnailHandler(fetcher.ValidatedURL,imagedelivery.Handler{Metadata:bootstrap.NewArticleMetadataHandler(bootstrap.NewArticleReadHandler())}).ServeHTTP(w, httptest.NewRequest("GET", "/api/thumbnail?url=http://127.0.0.1/private", nil))
	if w.Code != 422 {
		t.Fatalf("private URL status %d", w.Code)
	}
}
