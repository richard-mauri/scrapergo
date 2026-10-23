package main

import (
 "clean-article-reader/internal/httpapi"
 "clean-article-reader/internal/fetcher"
 "net/http"
 "clean-article-reader/internal/thumbnail"
	"bytes"
		"image"
	"image/color"
	"image/jpeg"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestThumbnailResizeDimensions(t *testing.T) {
	source := image.NewRGBA(image.Rect(0, 0, 1920, 1080))
	for y := 0; y < 1080; y++ {
		for x := 0; x < 1920; x++ {
			source.Set(x, y, color.RGBA{R: 100, G: 140, B: 200, A: 255})
		}
	}
	var in bytes.Buffer
	if err := jpeg.Encode(&in, source, nil); err != nil {
		t.Fatal(err)
	}
	out, err := thumbnail.CropAndEncode(in.Bytes(), 256)
	if err != nil {
		t.Fatal(err)
	}
	cfg, format, err := image.DecodeConfig(bytes.NewReader(out))
	if err != nil {
		t.Fatal(err)
	}
	if format != "jpeg" || cfg.Width != 256 || cfg.Height != 176 {
		t.Fatalf("wrong thumbnail: %s %dx%d", format, cfg.Width, cfg.Height)
	}
}
func TestThumbnailValidation(t *testing.T) {
	for _, url := range []string{"/api/thumbnail?url=https://example.com/article&width=1", "/api/thumbnail?url=https://example.com/article&width=99999"} {
		w := httptest.NewRecorder()
		httpapi.NewThumbnailHandler(fetcher.ValidatedURL,http.HandlerFunc(articleImage)).ServeHTTP(w, httptest.NewRequest("GET", url, nil))
		if w.Code != 400 {
			t.Fatalf("%s: %d", url, w.Code)
		}
	}
	w := httptest.NewRecorder()
	httpapi.NewThumbnailHandler(fetcher.ValidatedURL,http.HandlerFunc(articleImage)).ServeHTTP(w, httptest.NewRequest("POST", "/api/thumbnail", nil))
	if w.Code != 405 {
		t.Fatalf("POST status %d", w.Code)
	}
	w = httptest.NewRecorder()
	httpapi.NewThumbnailHandler(fetcher.ValidatedURL,http.HandlerFunc(articleImage)).ServeHTTP(w, httptest.NewRequest("GET", "/api/thumbnail?url=http://127.0.0.1/private", nil))
	if w.Code != 422 {
		t.Fatalf("private URL status %d", w.Code)
	}
}
func TestThumbnailCache(t *testing.T) {
	store := thumbnail.NewStore()
	store.Put("one", []byte("valid"))
	v, ok := store.Get("one")
	if !ok || string(v) != "valid" {
		t.Fatal("cache round-trip")
	}
	for i := 0; i < thumbnail.CacheEntries+2; i++ {
		store.Put(strconv.Itoa(i), []byte{1})
	}
	if store.Len() != thumbnail.CacheEntries {
		t.Fatal("cache capacity exceeded")
	}
}

func TestThumbnailSourceType(t *testing.T) {
	cases := []struct {
		data []byte
		want string
	}{
		{[]byte("RIFF\x00\x00\x00\x00WEBPVP8 "), "webp"},
		{[]byte("\x00\x00\x00\x18ftypavif"), "avif"},
		{[]byte("nonsense"), "unknown"},
	}
	for _, tc := range cases {
		if got := thumbnail.SourceType(tc.data); got != tc.want {
			t.Fatalf("got %q; want %q", got, tc.want)
		}
	}
}
func TestUnsupportedThumbnailFormatExplained(t *testing.T) {
	_, err := thumbnail.CropAndEncode([]byte("\x00\x00\x00\x18ftypavif"), 256)
	if err == nil || !strings.Contains(err.Error(), "avif") {
		t.Fatalf("expected AVIF diagnostic, got %v", err)
	}
}
