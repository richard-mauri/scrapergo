package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
 "net/url"
	"net/http/httptest"
	"strings"
	"sync"
	"time"
)

type Classifier struct {
 Read http.Handler
 ValidateURL func(context.Context,string) (*url.URL,error)
}

// Classification follows /api/read. Confirmed extraction failures are rejected;
// transient upstream and transport failures are unavailable, never accepted.
type classificationResult struct {
	Classification  string `json:"classification"`
	ReadableArticle bool   `json:"readableArticle"`
	Reason          string `json:"reason,omitempty"`
	ImageAvailable  bool   `json:"imageAvailable"`
 ContentStatus string `json:"contentStatus,omitempty"`
 ContentReason string `json:"contentReason,omitempty"`
}
type classificationCacheEntry struct {
	result  classificationResult
	expires time.Time
}

var classificationCache = struct {
	sync.Mutex
	entries map[string]classificationCacheEntry
}{entries: make(map[string]classificationCacheEntry)}
var classificationSlots = make(chan struct{}, 6)

func (h *Classifier) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "POST required", 405)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	var input struct {
		URL string `json:"url"`
	}
	if json.NewDecoder(r.Body).Decode(&input) != nil || len(input.URL) > 4096 || strings.TrimSpace(input.URL) == "" {
		http.Error(w, "invalid article URL", 400)
		return
	}
	// Preserve all the SSRF validation in /api/read, even for cached results.
	// Do not cache invalid URL responses.
	if _, err := h.ValidateURL(r.Context(), input.URL); err != nil {
		http.Error(w, "invalid public article URL", 422)
		return
	}
	classificationCache.Lock()
	entry, ok := classificationCache.entries[input.URL]
	classificationCache.Unlock()
	if ok && time.Now().Before(entry.expires) {
		_ = json.NewEncoder(w).Encode(entry.result)
		return
	}
	select {
	case classificationSlots <- struct{}{}:
		defer func() { <-classificationSlots }()
	case <-r.Context().Done():
		_ = json.NewEncoder(w).Encode(classificationResult{Classification: "unknown", Reason: "canceled"})
		return
	}
	classificationCache.Lock()
	entry, ok = classificationCache.entries[input.URL]
	classificationCache.Unlock()
	if ok && time.Now().Before(entry.expires) {
		_ = json.NewEncoder(w).Encode(entry.result)
		return
	}
	// Reader request uses the same extraction and classification rules as normal reading.
	ctx, cancel := context.WithTimeout(r.Context(), 12*time.Second)
	defer cancel()
	payload, _ := json.Marshal(input)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, "http://localhost/api/read", strings.NewReader(string(payload)))
	rec := httptest.NewRecorder()
	h.Read.ServeHTTP(rec, req)
	result := classificationResult{Classification: "unknown", Reason: "extraction_unavailable"}
	if ctx.Err() != nil {
		result.Reason = "timeout"
	} else if rec.Code == http.StatusOK {
		var article struct {
			ImageURL string `json:"imageUrl"`
 AccessType string `json:"accessType"`
 AccessNote string `json:"accessNote"`
		}
		if json.Unmarshal(rec.Body.Bytes(), &article) == nil {
			status, reason := classifyContentStatus(article.AccessType, article.AccessNote)
            if rec.Header().Get("X-ScraperGo-Content-Status") == "partial" {status, reason = "partial", "publisher_explicit_preview"}
            result = classificationResult{Classification: "article", ReadableArticle: true, ImageAvailable: article.ImageURL != "", ContentStatus: status, ContentReason: reason}
		}
	} else {
		var issue struct {
			Kind string `json:"kind"`
		}
		if json.Unmarshal(rec.Body.Bytes(), &issue) == nil {
			switch issue.Kind {
			case "page_index":
				result = classificationResult{Classification: "homepage", Reason: "homepage_or_index"}
			case "extraction_error":
				result = classificationResult{Classification: "unreadable", Reason: "no_readable_article"}
			case "publisher_access_denied", "access_error":
				result = classificationResult{Classification: "unavailable", Reason: "publisher_unavailable"}
			}
		}
	}
	ttl := time.Minute
	if result.Classification == "article" || result.Classification == "homepage" || result.Classification == "unreadable" {
		ttl = 30 * time.Minute
	}
	if ctx.Err() == nil {
		classificationCache.Lock()
		if len(classificationCache.entries) >= 512 {
			for k := range classificationCache.entries {
				delete(classificationCache.entries, k)
				break
			}
		}
		classificationCache.entries[input.URL] = classificationCacheEntry{result, time.Now().Add(ttl)}
		classificationCache.Unlock()
	}
	_ = json.NewEncoder(w).Encode(result)
}
