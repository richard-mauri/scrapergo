package main

import (
	"net/url"
	"regexp"
	"strings"
)

var imageSizeSuffix = regexp.MustCompile(`(?i)-[0-9]{2,5}x[0-9]{2,5}(\.[a-z0-9]+)$`)

// imageIdentity strips non-content delivery parameters and common image-size
// suffixes. Distinct photographs retain distinct paths and filenames.
func imageIdentity(raw string) string {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Hostname() == "" {
		return ""
	}
	path := strings.ToLower(u.EscapedPath())
	path = imageSizeSuffix.ReplaceAllString(path, "$1")
	return strings.ToLower(u.Hostname()) + path
}

// deduplicateArticleImages keeps the hero image and the first occurrence of
// each distinct inline photograph. Text, captions and credits are preserved.
func deduplicateArticleImages(a *Article) {
	if a == nil {
		return
	}
	seen := map[string]bool{}
	if key := imageIdentity(a.ImageURL); key != "" {
		seen[key] = true
	}
	blocks := make([]Block, 0, len(a.Blocks))
	for _, b := range a.Blocks {
		if b.Kind == "img" {
			key := imageIdentity(b.URL)
			if key != "" && seen[key] {
				continue
			}
			if key != "" {
				seen[key] = true
			}
		}
		blocks = append(blocks, b)
	}
	a.Blocks = blocks
}
