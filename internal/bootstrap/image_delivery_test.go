package bootstrap

import (
 "clean-article-reader/internal/imagedelivery"
	"net/http/httptest"
 "net/http"
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
		(imagedelivery.Handler{Metadata:NewArticleMetadataHandler(NewArticleReadHandler())}).ServeHTTP(w, httptest.NewRequest(tc.method, tc.url, nil))
		if w.Code != tc.want {
			t.Errorf("%s %s: got %d want %d", tc.method, tc.url, w.Code, tc.want)
		}
	}
}
func TestSelectedImageRejectsNonPublicURL(t *testing.T) {
	if _, _, err := imagedelivery.RetrieveSelectedImage(httptest.NewRequest("GET", "/", nil).Context(), "http://127.0.0.1/image.jpg", "https://example.com/article"); err == nil {
		t.Fatal("private image URL accepted")
	}
}

func TestArticleImageRejectsOutOfRangeCandidateIndex(t *testing.T) {
 metadata := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "application/json")
  _, _ = w.Write([]byte(`{"imageUrl":"https://example.com/photo.jpg","imageCandidates":["https://example.com/photo.jpg"]}`))
 })
 for _, index := range []string{"-1", "abc", "1"} {
  w := httptest.NewRecorder()
  req := httptest.NewRequest(http.MethodGet, "/api/article-image?url=https%3A%2F%2Fexample.com%2Fstory&index="+index,nil)
  (imagedelivery.Handler{Metadata:metadata}).ServeHTTP(w,req)
  if w.Code != http.StatusBadRequest {
   t.Errorf("candidate index %q: got %d, want 400",index,w.Code)
  }
 }
}

func TestArticleImageRejectsDifferentSelectedPhoto(t *testing.T) {
 metadata := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "application/json")
  _, _ = w.Write([]byte(`{"imageCandidates":["https://example.com/portrait.jpg","https://example.com/football.jpg"]}`))
 })
 w := httptest.NewRecorder()
 req := httptest.NewRequest(http.MethodGet, "/api/article-image?url=https%3A%2F%2Fexample.com%2Fstory&image=https%3A%2F%2Fexample.com%2Fwrong.jpg",nil)
 (imagedelivery.Handler{Metadata:metadata}).ServeHTTP(w,req)
 if w.Code != http.StatusNotFound {
  t.Fatalf("unselected photo: got %d, want 404",w.Code)
 }
}
