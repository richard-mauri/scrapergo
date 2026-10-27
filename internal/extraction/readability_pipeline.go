package extraction

import (
 "clean-article-reader/internal/article"
 "clean-article-reader/internal/sourcehtml"
	"bytes"
	"fmt"
	"net/url"
	"strings"

	readability "codeberg.org/readeck/go-readability/v2"
)

// extractWithReadability parses already-retrieved HTML; it makes no additional
// network requests. This preserves the existing SSRF protections and paywall gate.
func ExtractWithReadability(raw, source string, legacy func(string,string)(article.Article,error), inlineParts func(*sourcehtml.Node,string) []article.TextPart) (article.Article, error) {
	u, err := url.Parse(source)
	if err != nil {
		return article.Article{}, err
	}
	parsed, parseErr := readability.FromReader(strings.NewReader(raw), u)
	if parseErr == nil {
		var output bytes.Buffer
		if err = parsed.RenderText(&output); err == nil {
			text := strings.TrimSpace(output.String())
			chunks := strings.Split(text, "\n\n")
			blocks := make([]article.Block, 0, len(chunks))
			count := 0
			for _, chunk := range chunks {
				line := sourcehtml.Normalize(chunk)
				if line == "" || IsAdvertisingText(line) {
					continue
				}
				blocks = append(blocks, article.Block{Kind: "p", Text: line})
				count += len(strings.Fields(line))
			}
			structuredRecovered := false
			if complete, ok := RecoverStructuredDocument(raw, blocks, u); ok {
				blocks = complete
				count = CountWords(blocks)
				structuredRecovered = true
			}
			// Restore missing opening paragraphs before accepting an otherwise plausible
			// Readability selection. Do not shift the remainder of the article.
			if !structuredRecovered {
				blocks = RecoverLeadingParagraphs(raw, blocks, u, inlineParts)
				count = CountWords(blocks)
			}
			if count >= 60 {\n                blocks = RestoreSourceTables(blocks,raw,source)\n                count = CountWords(blocks)
				title := parsed.Title()
				if title == "" {
					title = "Untitled article"
				}
				img := sourcehtml.ResolveURL(parsed.ImageURL(), u)
				if img == "" {
					img = sourcehtml.ResolveURL(sourcehtml.Meta(sourcehtml.Parse(raw), "og:image", "twitter:image"), u)
				}
				var cleanHTML bytes.Buffer
				if parsed.RenderHTML(&cleanHTML) == nil {
					cleaned := cleanHTML.String()
					PreserveSelectedHTMLLinks(blocks, cleaned, u, inlineParts)
					if !structuredRecovered {
						ordered := PositionArticleFigures(blocks, cleaned, u)
						if strings.Contains(u.Path, "/gallery/") {
							ordered = SupplementGalleryFigures(ordered, raw, u)
						}
						blocks = ordered
					}
				}
				if !structuredRecovered {
					blocks = RecoverEditorialContinuity(raw, blocks, u)
					count = CountWords(blocks)
				}
				if img == "" {
					for _, photo := range blocks {
						if photo.Kind == "img" {
							img = photo.URL
							break
						}
					}
				}
				for _, photo := range blocks {
					if photo.Kind == "img" && photo.URL == img {
						img = ""
						break
					}
				}
				return article.Article{Title: title, Byline: parsed.Byline(), Site: parsed.SiteName(), URL: source, Blocks: blocks, WordCount: count, ImageURL: img, Method: "Go Readability v2"}, nil
			}
		}
	}
	fallback, fallbackErr := legacy(raw, source)
	if fallbackErr == nil {
		fallback.Method = "Legacy extractor fallback"
		return fallback, nil
	}
	if partial, ok := PartialHTMLArticle(raw, source); ok {
		return partial, nil
	}
	if parseErr != nil {
		return fallback, fmt.Errorf("Readability: %v; legacy extractor: %w", parseErr, fallbackErr)
	}
	return fallback, fallbackErr
}
