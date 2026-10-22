package main

import (
 "clean-article-reader/internal/thumbnail"
			"context"
				"fmt"
								"net/http"
	"net/url"
	"strconv"
		)

var thumbs = thumbnail.NewStore()

func articleThumbnail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		http.Error(w, "GET required", 405)
		return
	}
	width, err := thumbnail.Width(r)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	raw := r.URL.Query().Get("url")
	if len(raw) > 4096 {
		http.Error(w, "article URL too long", 400)
		return
	}
	if _, err = validatedURL(r.Context(), raw); err != nil {
		http.Error(w, "invalid public article URL", 422)
		return
	}
	key := thumbnail.Key(raw, width)
	if b, ok := thumbs.Get(key); ok {
		sendThumbnail(w, b, "HIT")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
	defer cancel()
	// Reuse the already shipped article-image delivery and extraction pipeline.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost/api/article-image?url="+url.QueryEscape(raw), nil)
	if err != nil {
		http.Error(w, "invalid request", 400)
		return
	}
	capture := newResponseCapture()
	articleImage(capture, req)
	if capture.status != http.StatusOK {
		stage := capture.header.Get("X-Image-Failure-Stage")
		if stage == "" {
			stage = "image-delivery"
		}
		w.Header().Set("X-Thumbnail-Failure-Stage", stage)
		// A missing publisher image is not a gateway failure.
		if capture.status == http.StatusNotFound {
			http.Error(w, "article has no selected image", http.StatusNotFound)
		} else {
			http.Error(w, "article image unavailable ("+stage+")", http.StatusBadGateway)
		}
		return
	}
	b, err := thumbnail.CropAndEncode([]byte(capture.body.String()), width)
	if err != nil {
		// Include only the local decoder failure; never echo an article URL.
		w.Header().Set("X-Thumbnail-Failure-Stage", "image-decode")
		http.Error(w, "image could not be resized: "+err.Error(), http.StatusUnsupportedMediaType)
		return
	}
	thumbs.Put(key, b)
	sendThumbnail(w, b, "MISS")
}
func sendThumbnail(w http.ResponseWriter, b []byte, cache string) {
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", strconv.Itoa(len(b)))
	w.Header().Set("Cache-Control", "private, max-age=300")
	w.Header().Set("X-Thumbnail-Cache", cache)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(b)
}
