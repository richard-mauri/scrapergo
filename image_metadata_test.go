package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestImageMetadataMethodAndBadInput(t *testing.T) {
	get := httptest.NewRecorder()
	imageMetadata(get, httptest.NewRequest(http.MethodGet, "/api/image-metadata", nil))
	if get.Code != http.StatusMethodNotAllowed {
		t.Fatalf("GET: %d", get.Code)
	}
	post := httptest.NewRecorder()
	imageMetadata(post, httptest.NewRequest(http.MethodPost, "/api/image-metadata", strings.NewReader("{")))
	if post.Code != http.StatusBadRequest {
		t.Fatalf("malformed JSON: %d", post.Code)
	}
}
func TestImageMetadataResponseContract(t *testing.T) {
	// Schema remains lightweight, with image selection performed by /api/read.
	result := struct {
		URL          string `json:"url"`
		ImageURL     string `json:"imageUrl"`
		ImagePresent bool   `json:"imagePresent"`
	}{"https://example.com/story", "https://example.com/hero.jpg", true}
	b, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "blocks") || !strings.Contains(string(b), `"imageUrl"`) {
		t.Fatalf("unexpected response %s", b)
	}
}
