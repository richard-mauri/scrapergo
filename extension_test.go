package main

import (
	"os"
	"strings"
	"testing"
)

func TestSafariFallbackUI(t *testing.T) {
	if !strings.Contains(indexHTML, "Open original article in Safari") {
		t.Fatal("fallback missing")
	}
}
func TestSafariExtensionArtifacts(t *testing.T) {
	for _, p := range []string{"safari-extension/manifest.json", "safari-extension/background.js", "safari-extension/options.html"} {
		if _, err := os.Stat(p); err != nil {
			t.Fatal(err)
		}
	}
}

func TestSafariDirectOpenNoPublisherTab(t *testing.T) {
	contents, err := os.ReadFile("safari-extension/background.js")
	if err != nil {
		t.Fatal(err)
	}
	script := string(contents)
	if !strings.Contains(script, "openReader(info.linkUrl)") {
		t.Fatal("context menu must open the reader directly")
	}
	if strings.Contains(script, "active: false") || strings.Contains(script, "SCRAPERGO_EXTRACT") {
		t.Fatal("background worker must not open a temporary publisher tab")
	}
}

func TestSafariReusesExistingReaderTab(t *testing.T) {
	b, err := os.ReadFile("safari-extension/background.js")
	if err != nil {
		t.Fatal(err)
	}
	js := string(b)
	for _, phrase := range []string{"browser.tabs.query({})", "browser.tabs.update(reader.id", "browser.tabs.create({url: destination", "pendingOpen"} {
		if !strings.Contains(js, phrase) {
			t.Errorf("missing tab reuse behavior: %s", phrase)
		}
	}
}
