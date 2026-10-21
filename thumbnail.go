package main

import (
	"bytes"
	"container/list"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

const (
	thumbnailWidthDefault = 256
	thumbnailWidthMinimum = 64
	thumbnailWidthMaximum = 640
	thumbnailMaxPixels    = 16000000
	thumbnailCacheEntries = 128
	thumbnailCacheTTL     = 30 * time.Minute
)

// Completed thumbnails only; bounded LRU and TTL. No raw publisher bytes are stored.
type thumbEntry struct {
	key   string
	data  []byte
	until time.Time
}
type thumbnailStore struct {
	mu    sync.Mutex
	items map[string]*list.Element
	order *list.List
}

var thumbs = thumbnailStore{items: make(map[string]*list.Element), order: list.New()}

func (s *thumbnailStore) get(key string) ([]byte, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e := s.items[key]
	if e == nil {
		return nil, false
	}
	v := e.Value.(thumbEntry)
	if time.Now().After(v.until) {
		s.order.Remove(e)
		delete(s.items, key)
		return nil, false
	}
	s.order.MoveToFront(e)
	return v.data, true
}
func (s *thumbnailStore) put(key string, data []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if e := s.items[key]; e != nil {
		s.order.Remove(e)
		delete(s.items, key)
	}
	s.items[key] = s.order.PushFront(thumbEntry{key: key, data: data, until: time.Now().Add(thumbnailCacheTTL)})
	for s.order.Len() > thumbnailCacheEntries {
		e := s.order.Back()
		v := e.Value.(thumbEntry)
		delete(s.items, v.key)
		s.order.Remove(e)
	}
}
func thumbKey(raw string, width int) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%d:%s", width, raw)))
	return hex.EncodeToString(sum[:])
}
func thumbnailWidth(r *http.Request) (int, error) {
	raw := r.URL.Query().Get("width")
	if raw == "" {
		return thumbnailWidthDefault, nil
	}
	n, e := strconv.Atoi(raw)
	if e != nil || n < thumbnailWidthMinimum || n > thumbnailWidthMaximum {
		return 0, errors.New("width must be 64..640")
	}
	return n, nil
}

// cropAndEncodeThumbnail performs center-cropped 16:11 rendering using only Go's
// standard library. It avoids unbounded image allocations before decoding.
func cropAndEncodeThumbnail(data []byte, width int) ([]byte, error) {
	cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	if format != "jpeg" && format != "png" && format != "gif" {
		return nil, errors.New("unsupported thumbnail source format")
	}
	if cfg.Width < 1 || cfg.Height < 1 || int64(cfg.Width)*int64(cfg.Height) > thumbnailMaxPixels {
		return nil, errors.New("image dimensions exceed limit")
	}
	src, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	height := width * 11 / 16
	bounds := src.Bounds()
	sw, sh := bounds.Dx(), bounds.Dy()
	// Crop input to 16:11, then sample the center of each output pixel.
	cropW, cropH := sw, sh
	if int64(sw)*11 > int64(sh)*16 {
		cropW = sh * 16 / 11
	} else {
		cropH = sw * 11 / 16
	}
	if cropW < 1 || cropH < 1 {
		return nil, errors.New("invalid crop dimensions")
	}
	startX := bounds.Min.X + (sw-cropW)/2
	startY := bounds.Min.Y + (sh-cropH)/2
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	// Fill entire rectangle, including opaque alpha for transparent PNG/GIF.
	draw.Draw(dst, dst.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
	for y := 0; y < height; y++ {
		sy := startY + (2*y+1)*cropH/(2*height)
		for x := 0; x < width; x++ {
			sx := startX + (2*x+1)*cropW/(2*width)
			dst.Set(x, y, src.At(sx, sy))
		}
	}
	var out bytes.Buffer
	if err := jpeg.Encode(&out, dst, &jpeg.Options{Quality: 82}); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func articleThumbnail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		http.Error(w, "GET required", 405)
		return
	}
	width, err := thumbnailWidth(r)
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
	key := thumbKey(raw, width)
	if b, ok := thumbs.get(key); ok {
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
		http.Error(w, "article image unavailable", http.StatusBadGateway)
		return
	}
	b, err := cropAndEncodeThumbnail([]byte(capture.body.String()), width)
	if err != nil {
		http.Error(w, "image could not be resized", http.StatusUnsupportedMediaType)
		return
	}
	thumbs.put(key, b)
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
