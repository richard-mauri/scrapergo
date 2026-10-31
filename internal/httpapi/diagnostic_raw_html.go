package httpapi

import (
    "context"
    "encoding/json"
    "net"
    "net/http"
    "net/url"
    "strings"
)

// RawHTMLDiagnosticHandler returns the exact HTML bytes from the standard
// validated publisher fetcher. Only direct loopback clients may use it.
// Responses are attachments to prevent HTML execution under this origin.
func RawHTMLDiagnosticHandler(download func(context.Context, string) (string, string, error)) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Cache-Control", "no-store")
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("Content-Security-Policy", "sandbox")
        if r.Method != http.MethodPost {
            w.Header().Set("Allow", "POST")
            http.Error(w, "POST required", http.StatusMethodNotAllowed)
            return
        }
        peer, _, err := net.SplitHostPort(r.RemoteAddr)
        if err != nil || net.ParseIP(peer) == nil || !net.ParseIP(peer).IsLoopback() {
            http.Error(w, "localhost clients only", http.StatusForbidden)
            return
        }
        r.Body = http.MaxBytesReader(w, r.Body, 4096)
        var input struct { URL string `json:"url"` }
        if err := json.NewDecoder(r.Body).Decode(&input); err != nil || strings.TrimSpace(input.URL) == "" {
            http.Error(w, "invalid URL JSON", http.StatusBadRequest)
            return
        }
        u, err := url.Parse(input.URL)
        if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" || u.User != nil {
            http.Error(w, "invalid HTTP(S) URL", http.StatusBadRequest)
            return
        }
        // The injected production downloader applies URL, DNS, redirect,
        // response-type, timeout, and five-MiB response-size restrictions.
        raw, finalURL, err := download(r.Context(), input.URL)
        if err != nil {
            http.Error(w, "publisher download failed: "+err.Error(), http.StatusUnprocessableEntity)
            return
        }
        w.Header().Set("Content-Type", "text/html; charset=utf-8")
        w.Header().Set("Content-Disposition", `attachment; filename="scrapergo-source.html"`)
        w.Header().Set("X-ScraperGo-Final-URL", finalURL)
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte(raw))
    }
}
