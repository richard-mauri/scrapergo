package main

import "testing"

// The NOAA article exposed a panic when an empty closing tag was tokenized.
func TestParseHTMLMalformedClosingTags(t *testing.T) {
	for _, raw := range []string{
		"<html><body><p>before</p></><p>after</p></body></html>",
		"<div>first</ >second</div>",
		"<p>before</   >after</p>",
		"<p>before</\t>after</p>",
	} {
		root := parseHTML(raw)
		if root == nil || len(root.Kids) == 0 {
			t.Fatalf("parseHTML(%q) returned no tree", raw)
		}
		if got := textOf(root); got == "" {
			t.Fatalf("parseHTML(%q) lost all content", raw)
		}
	}
}
