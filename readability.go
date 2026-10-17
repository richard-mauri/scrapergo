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
				img := resolveImageURL(parsed.ImageURL(), u)
				// Some publishers only advertise the lead photograph through Open Graph.
				if img == "" {
					img = resolveImageURL(meta(parseHTML(raw), "og:image", "twitter:image"), u)
				}
				// Extract only images in the Readability-cleaned article, never in site chrome.
				var cleanHTML bytes.Buffer
				if parsed.RenderHTML(&cleanHTML) == nil {
					images := articleImages(cleanHTML.String(), u)
					for _, photo := range images {
						if photo.URL == img {
							continue
						} // avoid duplicating the hero
						blocks = append(blocks, photo)
					}
				}
				if img == "" {
					for _, photo := range blocks {
						if photo.Kind == "img" {
							img = photo.URL
							break
						}
					}
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

// resolveImageURL accepts only http(s) URLs, including relative references.
func resolveImageURL(raw string, base *url.URL) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	resolved := base.ResolveReference(u)
	if resolved.Scheme != "http" && resolved.Scheme != "https" {
		return ""
	}
	if resolved.Hostname() == "" || resolved.User != nil {
		return ""
	}
	return resolved.String()
}

// articleImages reads imagery from the already cleaned Readability output.
// Captions are kept as plain text; publisher HTML is never injected into the UI.
func articleImages(cleaned string, base *url.URL) []Block {
	root := parseHTML(cleaned)
	images := []Block{}
	seen := map[string]bool{}
	for _, n := range find(root, func(n *Node) bool { return n.Tag == "img" }) {
		src := n.Attr["src"]
		if src == "" {
			src = n.Attr["data-src"]
		}
		if src == "" {
			src = n.Attr["data-original"]
		}
		link := resolveImageURL(src, base)
		if link == "" || seen[link] {
			continue
		}
		seen[link] = true
		caption := ""
		for ancestor := n.Parent; ancestor != nil; ancestor = ancestor.Parent {
			if ancestor.Tag == "figure" {
				for _, c := range find(ancestor, func(x *Node) bool { return x.Tag == "figcaption" }) {
					caption = textOf(c)
					break
				}
				break
			}
			if ancestor.Tag == "body" {
				break
			}
		}
		if caption == "" {
			caption = normalize(n.Attr["alt"])
		}
		images = append(images, Block{Kind: "img", Text: caption, URL: link})
		if len(images) >= 12 {
			break
		}
	}
	return images
}
