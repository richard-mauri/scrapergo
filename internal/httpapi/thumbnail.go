package httpapi

import (
 "clean-article-reader/internal/thumbnail"
 "net/http/httptest"
			"context"
 "time"
												"net/http"
	"net/url"
	"strconv"
		)

type ThumbnailHandler struct {
 ValidateURL func(context.Context,string)(*url.URL,error)
 ArticleImage http.Handler
 Store *thumbnail.Store
}

func NewThumbnailHandler(validate func(context.Context,string)(*url.URL,error), image http.Handler) *ThumbnailHandler {
 return &ThumbnailHandler{ValidateURL:validate,ArticleImage:image,Store:thumbnail.NewStore()}
}

func (h *ThumbnailHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	if _, err = h.ValidateURL(r.Context(), raw); err != nil {
		http.Error(w, "invalid public article URL", 422)
		return
	}
	key := thumbnail.Key(raw, width)
	if b, ok := h.Store.Get(key); ok {
		SendThumbnail(w, b, "HIT")
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
	capture := httptest.NewRecorder()
	h.ArticleImage.ServeHTTP(capture, req)
	if capture.Code != http.StatusOK {
		stage := capture.Header().Get("X-Image-Failure-Stage")
		if stage == "" {
			stage = "image-delivery"
		}
		w.Header().Set("X-Thumbnail-Failure-Stage", stage)
		// A missing publisher image is not a gateway failure.
		if capture.Code == http.StatusNotFound {
			http.Error(w, "article has no selected image", http.StatusNotFound)
		} else {
			http.Error(w, "article image unavailable ("+stage+")", http.StatusBadGateway)
		}
		return
	}
	b, err := thumbnail.CropAndEncode(capture.Body.Bytes(), width)
	if err != nil {
		// Include only the local decoder failure; never echo an article URL.
		w.Header().Set("X-Thumbnail-Failure-Stage", "image-decode")
		http.Error(w, "image could not be resized: "+err.Error(), http.StatusUnsupportedMediaType)
		return
	}
	h.Store.Put(key, b)
	SendThumbnail(w, b, "MISS")
}
func SendThumbnail(w http.ResponseWriter, b []byte, cache string) {
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", strconv.Itoa(len(b)))
	w.Header().Set("Cache-Control", "private, max-age=300")
	w.Header().Set("X-Thumbnail-Cache", cache)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(b)
}
