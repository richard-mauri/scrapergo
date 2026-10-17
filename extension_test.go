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
