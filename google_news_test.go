package main

import (
 "clean-article-reader/internal/googlenews"
	"strings"
	"testing"
)

func TestGoogleNewsID(t *testing.T) {
	id, ok, err := googlenews.ArticleID("https://news.google.com/read/CBMikAFBVV95cUxPMzlwQ0hUQkZHbXFyNWRpNjEzRU5nVXBVcXExZC1sdkJTS25TaXpKSGhyeWZsVVpDYTczTmNRSG1SbXo5LThtcnpUNkVobHo4SXgxVFVtU1NQTFFxeDdKMUk3YW9weEszSW5vNUVzb1hFY1Mxcmt1YXN3Qlc3U2NvM3ZITHo4UUo0WDA3X2J6cUjSAZYBQVVfeXFMUFhGeGswZkgtQllGWW03RlBxRGxQU1E5dGFHNV9hTzNOTUhmTEVpdUo2b01DaFoxRzFybFdhcE1idzAzdGhBYmFoMTNPQlh6eHA0enRET05aejNkTF9vT2VlcXNEUkJtMVVIdWRVUWFPSGNJM3pROGNHazdTV2NBc0o2elRpSFpQRlE3QThvdmUzYl94QV9n?hl=en-US")
	if err != nil || !ok || !strings.HasPrefix(id, "CBMikAFB") {
		t.Fatalf("id %q ok %v error %v", id, ok, err)
	}
	_, ok, err = googlenews.ArticleID("https://example.com/news")
	if err != nil || ok {
		t.Fatal("ordinary URL must bypass resolver")
	}
}
func TestNewsRPCPayload(t *testing.T) {
	v, e := googlenews.RPCPayload("CBMabcABC1234567", "1760700000", "sign-test")
	if e != nil || !strings.Contains(v.Encode(), "f.req=") {
		t.Fatalf("bad payload %v %v", v, e)
	}
}
func TestNewsRPCResponse(t *testing.T) {
	raw := []byte(")]}'\n\n[[\"Fbv4je\",\"unused\",\"[\\\"garturlres\\\",\\\"https://example.com/story\\\"]\"]]")
	u, e := googlenews.ParseRPCResponse(raw)
	if e != nil || u != "https://example.com/story" {
		t.Fatalf("got %q %v", u, e)
	}
}
func TestNoSignature(t *testing.T) {
	if len(googlenews.NewsSignatureRE.FindStringSubmatch("<html>consent</html>")) != 0 {
		t.Fatal("false signature")
	}
}

func TestParseNewsRPCStandardFrame(t *testing.T) {
	body := []byte(")]}'\n\n[[\"wrb.fr\",\"Fbv4je\",\"[null,\\\"https://example.com/news/story\\\"]\",null,null,null]]")
	got, err := googlenews.ParseRPCResponse(body)
	if err != nil || got != "https://example.com/news/story" {
		t.Fatalf("url=%q err=%v", got, err)
	}
}
func TestParseNewsRPCBlocked(t *testing.T) {
	if _, err := googlenews.ParseRPCResponse([]byte("<html>consent required</html>")); err == nil {
		t.Fatal("expected error")
	}
}
