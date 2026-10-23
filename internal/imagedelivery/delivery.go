package imagedelivery

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
 "clean-article-reader/internal/fetcher"
)

const maxDeliveredImage = 4 << 20

// articleImage serves only an image selected by the existing article extractor.
// It does not accept a caller-supplied image URL.
type Handler struct {
 Metadata http.Handler
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		http.Error(w, "GET required", 405)
		return
	}
	articleURL := r.URL.Query().Get("url")
	if len(articleURL) > 4096 {
		http.Error(w, "article URL too long", 400)
		return
	}
	if _, err := fetcher.ValidatedURL(r.Context(), articleURL); err != nil {
		http.Error(w, "invalid public article URL", 422)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
	defer cancel()
	payload, _ := json.Marshal(map[string]string{"url": articleURL})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://localhost/api/image-metadata", strings.NewReader(string(payload)))
	if err != nil {
		http.Error(w, "invalid request", 400)
		return
	}
	rec := newResponseCapture()
	h.Metadata.ServeHTTP(rec, req)
	if rec.status != http.StatusOK {
		w.Header().Set("X-Image-Failure-Stage", "metadata")
		http.Error(w, "article image lookup failed", http.StatusBadGateway)
		return
	}
	var meta struct {
		ImageURL   string   `json:"imageUrl"`
		Candidates []string `json:"imageCandidates"`
	}
	if err = json.Unmarshal([]byte(rec.body.String()), &meta); err != nil {
		w.Header().Set("X-Image-Failure-Stage", "metadata")
		http.Error(w, "invalid image metadata", http.StatusBadGateway)
		return
	}
	if len(meta.Candidates) == 0 && meta.ImageURL != "" {
		meta.Candidates = []string{meta.ImageURL}
	}
	if len(meta.Candidates) == 0 {
		w.Header().Set("X-Image-Failure-Stage", "no-selected-image")
		http.Error(w, "article has no selected image", http.StatusNotFound)
		return
	}
	var b []byte
	var typ string
	// Bound fallback attempts: the reader may show several inline images,
	// but this endpoint must never download an unbounded gallery.
	attempts := len(meta.Candidates)
	if attempts > 3 {
		attempts = 3
	}
	for _, candidate := range meta.Candidates[:attempts] {
		b, typ, err = RetrieveSelectedImage(ctx, candidate, articleURL)
		if err == nil {
			break
		}
		if ctx.Err() != nil {
			break
		}
	}
	if err != nil {
		w.Header().Set("X-Image-Failure-Stage", FailureKind(err))
		http.Error(w, "selected image unavailable", http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", typ)
	w.Header().Set("Content-Length", fmt.Sprint(len(b)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(b)
}

type responseCapture struct {
	header http.Header
	body   strings.Builder
	status int
}

func newResponseCapture() *responseCapture {
	return &responseCapture{header: make(http.Header), status: 200}
}
func (r *responseCapture) Header() http.Header         { return r.header }
func (r *responseCapture) WriteHeader(code int)        { r.status = code }
func (r *responseCapture) Write(p []byte) (int, error) { return r.body.Write(p) }

// httpsImageURL upgrades publisher-selected image references, never permitting
// a plaintext fetch. HTTPS validation and safe-host checks still run afterward.
func HTTPSImageURL(raw string) (string, error) {
 u, err := url.Parse(raw)
 if err != nil { return "", err }
 if u.Scheme != "http" && u.Scheme != "https" { return "", errors.New("image URL scheme unsupported") }
 if u.Scheme == "http" { u.Scheme = "https" }
 return u.String(), nil
}

func RetrieveSelectedImage(ctx context.Context, raw, articleURL string) ([]byte, string, error) {
	normalized, err := HTTPSImageURL(raw)
	if err != nil { return nil, "", err }
	u, err := fetcher.ValidatedURL(ctx, normalized)
	if err != nil {
		return nil, "", err
	}
	if u.Scheme != "https" {
		return nil, "", errors.New("image must use HTTPS")
	}
	article, err := url.Parse(articleURL)
	if err != nil {
		return nil, "", err
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	transport.DialContext = func(c context.Context, network, addr string) (net.Conn, error) {
		host, port, e := net.SplitHostPort(addr)
		if e != nil {
			return nil, e
		}
		if e = fetcher.SafeHost(c, host); e != nil {
			return nil, e
		}
		return (&net.Dialer{Timeout: 5 * time.Second}).DialContext(c, network, net.JoinHostPort(host, port))
	}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, Timeout: 12 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 3 {
			return errors.New("too many redirects")
		}
		v, e := fetcher.ValidatedURL(req.Context(), req.URL.String())
		if e != nil {
			return e
		}
		if v.Scheme != "https" {
			return errors.New("insecure redirect")
		}
		return nil
	}}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 Version/18.0 Safari/605.1.15")
	req.Header.Set("Accept", "image/webp,image/png,image/jpeg,image/gif,image/*;q=0.8")
	// Preserve the full publisher article as referrer for images protected by
	// normal anti-hotlink rules. The image URL is still selected by extraction,
	// and all public-host, HTTPS, redirect and byte limits remain enforced.
	req.Header.Set("Referer", ArticleReferrer(article))
	res, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, "", fmt.Errorf("upstream image status %d", res.StatusCode)
	}
	if res.ContentLength > maxDeliveredImage {
		return nil, "", errors.New("image too large")
	}
	b, err := io.ReadAll(io.LimitReader(res.Body, maxDeliveredImage+1))
	if err != nil {
		return nil, "", err
	}
	if len(b) > maxDeliveredImage {
		return nil, "", errors.New("image too large")
	}
	detected := http.DetectContentType(b)
	switch detected {
	case "image/jpeg", "image/png", "image/gif", "image/webp":
		return b, detected, nil
	}
	return nil, "", errors.New("unsupported or nonimage response")
}

// articleReferrer uses a validated publisher article URL, without credentials,
// fragment or query parameters that may contain private tracking tokens.
func ArticleReferrer(article *url.URL) string {
	copy := *article
	copy.User = nil
	copy.RawQuery = ""
	copy.ForceQuery = false
	copy.Fragment = ""
	copy.RawFragment = ""
	return copy.String()
}

// imageRetrievalFailureKind exposes a fixed diagnostic category, never a publisher
// URL, response body, hostname, or internal network error to public clients.
func FailureKind(err error) string {
	if err == nil {
		return "none"
	}
	if strings.HasPrefix(err.Error(), "upstream image status ") {
		return "upstream-status"
	}
	if strings.Contains(err.Error(), "image too large") {
		return "size-limit"
	}
	if strings.Contains(err.Error(), "unsupported or nonimage response") {
		return "unsupported-source"
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "timeout"
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return "timeout"
	}
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return "dns"
	}
	var certErr *tls.CertificateVerificationError
	if errors.As(err, &certErr) {
		return "tls-certificate"
	}
	var recordErr tls.RecordHeaderError
	if errors.As(err, &recordErr) {
		return "tls-handshake"
	}
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return "connection"
	}
	return "image-fetch"
}
