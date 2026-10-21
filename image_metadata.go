package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
)

// imageMetadata invokes the ordinary article API, ensuring that image selection,
// publisher access checks and extraction policies remain identical to /api/read.
// The response is smaller, but retrieval/extraction cost is not reduced.
func imageMetadata(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	recorder := httptest.NewRecorder()
	api(recorder, r)
	result := recorder.Result()
	defer result.Body.Close()
	if result.StatusCode != http.StatusOK {
		w.WriteHeader(result.StatusCode)
		// Forward the existing API's structured error response, not a new error policy.
		var failure json.RawMessage
		if json.NewDecoder(result.Body).Decode(&failure) == nil && json.Valid(failure) {
			_, _ = w.Write(failure)
		} else {
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "article lookup failed"})
		}
		return
	}
	var article Article
	if err := json.NewDecoder(result.Body).Decode(&article); err != nil {
		http.Error(w, "invalid extraction response", http.StatusInternalServerError)
		return
	}
	candidates := articleImageCandidates(article)
	selected := ""
	if len(candidates) > 0 {
		selected = candidates[0]
	}
	_ = json.NewEncoder(w).Encode(struct {
		URL          string   `json:"url"`
		ImageURL     string   `json:"imageUrl"`
		ImagePresent bool     `json:"imagePresent"`
		Candidates   []string `json:"imageCandidates"`
	}{URL: article.URL, ImageURL: selected, ImagePresent: selected != "", Candidates: candidates})
}

// articleImageCandidates uses the same hero and ordered inline image references
// that the reader renders. Readability may intentionally omit ImageURL when a
// matching block already displays it; that must not hide it from thumbnails.
func articleImageCandidates(a Article) []string {
	var out []string
	seen := make(map[string]bool)
	add := func(raw string) {
		raw = strings.TrimSpace(raw)
		if raw != "" && !seen[raw] {
			seen[raw] = true
			out = append(out, raw)
		}
	}
	add(a.ImageURL)
	for _, b := range a.Blocks {
		if b.Kind == "img" {
			add(b.URL)
		}
	}
	return out
}
