package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDiagnosticRejectsRemotePeer(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/diagnostic/extract", strings.NewReader(`{"url":"https://example.com/a","html":"<html></html>"}`))
	req.RemoteAddr = "198.51.100.7:24000"
	rec := httptest.NewRecorder()
	diagnosticExtract(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("remote status %d", rec.Code)
	}
}
func TestDiagnosticRejectsMalformedAndOversized(t *testing.T) {
	for _, data := range []string{`{`, `{"url":"file:///tmp/test","html":"<html/>"}`, `{"url":"https://example.com/a","html":""}`, `{"url":"https://example.com/a","html":"<p>a</p>"} {}`} {
		req := httptest.NewRequest(http.MethodPost, "/api/diagnostic/extract", strings.NewReader(data))
		req.RemoteAddr = "127.0.0.1:3333"
		rec := httptest.NewRecorder()
		diagnosticExtract(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("invalid %q -> %d", data, rec.Code)
		}
	}
}
func TestDiagnosticRequiresPost(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/diagnostic/extract", nil)
	req.RemoteAddr = "127.0.0.1:3333"
	rec := httptest.NewRecorder()
	diagnosticExtract(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status %d", rec.Code)
	}
}
