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
 return ExtractWithReadabilityTraced(raw, source, legacy, inlineParts, nil)
}

// ExtractWithReadabilityTraced uses the exact same extraction path as the
// standard entry point; tracing only observes stage outcomes.
func ExtractWithReadabilityTraced(raw, source string, legacy func(string,string)(article.Article,error), inlineParts func(*sourcehtml.Node,string) []article.TextPart, trace *RecoveryTrace) (article.Article, error) {
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
			// Record the initial RenderText representation before any recovery
			// stage can insert, replace or restructure article blocks.
			trace.Observe("readability-render-text", blocks)
			structuredRecovered := false
			if trace == nil {
                if complete, ok := RecoverStructuredDocument(raw, blocks, u); ok {
                    blocks = complete
                    count = CountWords(blocks)
                    structuredRecovered = true
                }
            } else {
                blocks = trace.Run("structured-document", blocks, func() []article.Block {
                    if complete, ok := RecoverStructuredDocument(raw, blocks, u); ok {
                        structuredRecovered = true
                        return complete
                    }
                    return blocks
                })
                if structuredRecovered { count = CountWords(blocks) }
            }
			// Restore missing opening paragraphs before accepting an otherwise plausible
			// Readability selection. Do not shift the remainder of the article.
			if !structuredRecovered {
				blocks = trace.Run("leading-paragraphs", blocks, func() []article.Block {
                    return RecoverLeadingParagraphs(raw, blocks, u, inlineParts)
                })
				count = CountWords(blocks)
			}
			if count >= 60 {
				blocks = trace.Run("source-tables", blocks, func() []article.Block {
                    return RestoreSourceTables(blocks, raw, source)
                })
				count = CountWords(blocks)
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
					// Reconcile flattened figure captions only when the selected HTML
					// identifies the original image and caption semantics.
					if !structuredRecovered {
						blocks = trace.Run("caption-reconciliation", blocks, func() []article.Block {
							return reconcileImageCaptionCreditTriples(removeFlattenedFigureCaptions(blocks, cleaned), raw, cleaned)
						})
					}
					blocks = trace.Run("selected-html-links", blocks, func() []article.Block {
                        PreserveSelectedHTMLLinks(blocks, cleaned, u, inlineParts)
                        return blocks
                    })
					if !structuredRecovered {
						ordered := trace.Run("position-figures", blocks, func() []article.Block {
                            return PositionArticleFigures(blocks, cleaned, u)
                        })
						if strings.Contains(u.Path, "/gallery/") {
							ordered = SupplementGalleryFigures(ordered, raw, u)
						}
						blocks = ordered
					}
				}
				if !structuredRecovered {
					blocks = trace.Run("editorial-continuity", blocks, func() []article.Block {
                    return RecoverEditorialContinuity(raw, blocks, u)
                })
				}
				blocks = trace.Run("institutional-tail", blocks, func() []article.Block {
                    return RecoverInstitutionalTail(raw, blocks, u)
                })
				blocks = trace.Run("heading-anchored-prose", blocks, func() []article.Block {
                    return RestoreHeadingAnchoredProse(raw, blocks, u)
                })
				// Readability's selected HTML retains heading/time structure that
				// RenderText flattens into undifferentiated paragraphs.
				if cleanHTML.Len() > 0 {
					blocks = trace.Run("readability-structure", blocks, func() []article.Block {
                        return PreserveReadabilityStructure(blocks, cleanHTML.String())
                    })
				}
				count = CountWords(blocks)
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
	if trace != nil { trace.SetOutcome("legacy-fallback", "executed_unchanged") }
    fallback, fallbackErr := legacy(raw, source)
	if fallbackErr == nil {
		if trace != nil { trace.SetOutcome("legacy-fallback", "fallback_selected") }
        fallback.Method = "Legacy extractor fallback"
		return fallback, nil
	}
	if trace != nil { trace.SetOutcome("partial-html-fallback", "executed_unchanged") }
    if partial, ok := PartialHTMLArticle(raw, source); ok {
        if trace != nil { trace.SetOutcome("partial-html-fallback", "fallback_selected") }
        return partial, nil
    }
	if parseErr != nil {
		return fallback, fmt.Errorf("Readability: %v; legacy extractor: %w", parseErr, fallbackErr)
	}
	return fallback, fallbackErr
}
