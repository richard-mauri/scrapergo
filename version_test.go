package main

import (
	"strings"
	"testing"
)

func TestVersionSingleSource(t *testing.T) {
	html := renderedIndexHTML()
	if !strings.Contains(html, "ScraperGo · v"+appVersion) {
		t.Fatalf("header version differs from appVersion %s", appVersion)
	}
	if strings.Contains(html, "__APP_VERSION__") {
		t.Fatal("unexpanded version placeholder")
	}
	if strings.Contains(indexHTML, "ScraperGo · v0.") {
		t.Fatal("hardcoded version in template")
	}
}
