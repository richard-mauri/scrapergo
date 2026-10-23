package diagnostics

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
)

const MaxHTML = 5 << 20

type HTMLHandler struct { WriteExtracted func(http.ResponseWriter,string,string) }

// diagnosticExtract examines supplied HTML without fetching the URL or any assets.
// It is intentionally disabled for callers whose direct TCP peer is not loopback.
func (h HTMLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "POST required"})
		return
	}
	peer, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil || net.ParseIP(peer) == nil || !net.ParseIP(peer).IsLoopback() {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "diagnostic endpoint accepts localhost clients only"})
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, MaxHTML+8192)
	dec := json.NewDecoder(r.Body)
	var input struct {
		URL  string `json:"url"`
		HTML string `json:"html"`
	}
	if err := dec.Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid request JSON or excessive HTML size"})
		return
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "one JSON object required"})
		return
	}
	u, err := url.Parse(input.URL)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Hostname() == "" || u.User != nil || u.Fragment != "" || strings.ContainsAny(input.URL, "\r\n") || len(input.HTML) == 0 || len(input.HTML) > MaxHTML {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "valid HTTP(S) article URL and HTML (maximum 5 MiB) required"})
		return
	}
	h.WriteExtracted(w, input.HTML, input.URL)
}
