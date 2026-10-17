package main

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"

	readability "codeberg.org/readeck/go-readability/v2"
)

// extractWithReadability parses already-retrieved HTML; it makes no additional
// network requests. This preserves the existing SSRF protections and paywall gate.
func extractWithReadability(raw, source string) (Article, error) {
	u, err := url.Parse(source)
	if err != nil {
		return Article{}, err
	}
	parsed, parseErr := readability.FromReader(strings.NewReader(raw), u)
	if parseErr == nil {
		var output bytes.Buffer
		if err = parsed.RenderText(&output); err == nil {
			text := strings.TrimSpace(output.String())
			chunks := strings.Split(text, "\n\n")
			blocks := make([]Block, 0, len(chunks))
			count := 0
			for _, chunk := range chunks {
				line := normalize(chunk)
				if line == "" || sentenceAdRE.MatchString(line) {
					continue
				}
				blocks = append(blocks, Block{Kind: "p", Text: line})
				count += len(strings.Fields(line))
			}
			if count >= 60 {
				title := parsed.Title()
				if title == "" {
					title = "Untitled article"
				}
				img := parsed.ImageURL()
				if imageURL, e := url.Parse(img); e != nil || (imageURL.Scheme != "https" && imageURL.Scheme != "http") {
					img = ""
				}
				return Article{Title: title, Byline: parsed.Byline(), Site: parsed.SiteName(), URL: source, Blocks: blocks, WordCount: count, ImageURL: img, Method: "Go Readability v2"}, nil
			}
		}
	}
	fallback, fallbackErr := extract(raw, source)
	if fallbackErr == nil {
		fallback.Method = "Legacy extractor fallback"
		return fallback, nil
	}
	if parseErr != nil {
		return fallback, fmt.Errorf("Readability: %v; legacy extractor: %w", parseErr, fallbackErr)
	}
	return fallback, fallbackErr
}
