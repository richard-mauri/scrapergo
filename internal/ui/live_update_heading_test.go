package ui

import (
	"strings"
	"testing"
)

// Only confirmed source-backed live headings receive update-divider styling.
// Ordinary article subheadings retain the established reader layout.
func TestLiveUpdateHeadingPresentation(t *testing.T) {
	if !strings.Contains(readerCSS, ".paper.live-feed .content h3{") {
		t.Fatal("missing scoped live-feed heading presentation")
	}
	if !strings.Contains(readerJS, "classList.toggle('live-feed'") {
		t.Fatal("reader does not identify live-feed paths")
	}
	if !strings.Contains(readerCSS, ".live-update-time{") || !strings.Contains(readerCSS, ".live-update-title{") {
		t.Fatal("missing standalone live update timestamp or title styling")
	}
}
