package main

import (
 "clean-article-reader/internal/reader"
 "clean-article-reader/internal/extraction"
 "clean-article-reader/internal/app"
 "clean-article-reader/internal/diagnostics"
 "clean-article-reader/internal/httpapi"
 "clean-article-reader/internal/fetcher"
 "clean-article-reader/internal/imagedelivery"
 "clean-article-reader/internal/pageclassify"
 "clean-article-reader/internal/structured"
	articlemodel "clean-article-reader/internal/article"
 sourcehtml "clean-article-reader/internal/sourcehtml"
	"context"
	_ "embed"
	"encoding/json"
	"flag"
	"log"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

// ScraperGo application version is defined once and shared by all presentation paths.
const appVersion = "0.2.85"
var appBuild = "Build 000000"

func appDisplayVersion() string { return "ScraperGo v" + appVersion + " · " + appBuild }
const sampleURL = "https://247wallst.com/investing/2026/10/16/taiwan-is-planning-4-billion-industrial-parks-in-texas-for-nvidias-suppliers/"

//go:embed tanner.jpg
var tannerJPG []byte

const maxHTML = 5 << 20

// Compatibility aliases keep existing extraction and rendering callers unchanged
// as the canonical article model moves behind an importable package boundary.
type InlineLink = articlemodel.InlineLink
type TextPart = articlemodel.TextPart
type Block = articlemodel.Block
type Article = articlemodel.Article

type Node = sourcehtml.Node

var blockedTags = sourcehtml.BlockedTags
var wsRE = regexp.MustCompile(`\s+`)

func normalize(s string) string { return sourcehtml.Normalize(s) }
func parseHTML(raw string) *Node { return sourcehtml.Parse(raw) }
func find(root *Node, test func(*Node) bool) []*Node { return sourcehtml.Find(root,test) }
func textOf(n *Node) string { return sourcehtml.Text(n) }
func meta(root *Node, keys ...string) string { return sourcehtml.Meta(root, keys...) }
func articleImageURL(n *Node, baseURL string) string { return sourcehtml.ImageURL(n,baseURL) }

// Legacy entry points delegate all request validation and HTML retrieval to internal/fetcher.
func safeHost(ctx context.Context, host string) error { return fetcher.SafeHost(ctx, host) }
func validatedURL(ctx context.Context, address string) (*url.URL, error) { return fetcher.ValidatedURL(ctx,address) }
func setBrowserHeaders(req *http.Request) { fetcher.SetBrowserHeaders(req) }
func setReadeckLikeHeaders(req *http.Request) { fetcher.SetReadeckLikeHeaders(req) }
func alternateHeaderRetry(code int) bool { return fetcher.AlternateHeaderRetry(code) }
func download(ctx context.Context, address string) (string,string,error) { return fetcher.Download(ctx,address) }
type publisherHTTPError = fetcher.PublisherHTTPError
func publisherAccessMessage(err error) (string,bool) { return fetcher.PublisherAccessMessage(err) }

func api(w http.ResponseWriter, r *http.Request) {
 newArticleReadHandler().ServeHTTP(w,r)
}

// writeExtractedArticle is shared by normal retrieval and diagnostic HTML submission.
func writeExtractedArticle(w http.ResponseWriter, raw, finalURL string) {
	logSlateRetrievalDiagnostic(raw, finalURL)
	// A navigation index is not a single article. Classify conservatively before
	// Readability can flatten its menu and story cards into a false article.
	if (pageclassify.IsIndex(raw) || pageclassify.IsLanding(raw)) && !extraction.HasPartialArticleEvidence(raw, finalURL) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"kind":  "page_index",
			"error": "This appears to be a homepage or index, not a single article. Open the original website to browse its links.",
			"url":   finalURL,
		})
		return
	}
	// Extract before deciding whether restricted pages offer readable public text.
	// Restriction metadata alone does not establish that no usable preview exists.
	kind := accessRestriction(raw)
	cleanedArticleHTML := structured.StripEmbeddedMedia(raw)
	extracted, extractionErr := extractWithEvidence(cleanedArticleHTML, finalURL)
	article := extracted.Article
	if extractionErr == nil {
		article = reader.DefaultProcessor().Process(extracted, cleanedArticleHTML)
		logSlateExtractionDiagnostic(article.Blocks, finalURL)
	}
	if kind != "" {
		log.Printf("restriction publisher=%s kind=%s evidence=%s html_bytes=%d public_words=%d", mustHostname(finalURL), kind, restrictionEvidence(raw), len(raw), article.WordCount)
		if extractionErr == nil && article.WordCount >= 80 {
			article.AccessType = kind
			article.AccessNote = accessNotice(kind)
			_ = json.NewEncoder(w).Encode(article)
			return
		}
		w.WriteHeader(422)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": accessNotice(kind), "kind": kind, "publisher": mustHostname(finalURL), "evidence": restrictionEvidence(raw)})
		return
	}
	if (extractionErr == nil && (pageclassify.IsThinPortal(raw, article.WordCount) || pageclassify.IsSparseRootLanding(raw, finalURL, article.WordCount)) || extractionErr != nil && pageclassify.IsSearchPortal(raw)) && !extraction.HasPartialArticleEvidence(raw, finalURL) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"kind":  "page_index",
			"error": "This appears to be a homepage or index, not a single article. Open the original website to browse its links.",
			"url":   finalURL,
		})
		return
	}
	if extractionErr != nil {
		w.WriteHeader(422)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "ScraperGo could not identify a readable article on this page. Open the original website to continue.", "kind": "extraction_error", "url": finalURL})
		return
	}
	_ = json.NewEncoder(w).Encode(article)
}

// accessNotice deliberately avoids claiming that a displayed preview is complete.
func accessNotice(kind string) string {
	switch kind {
	case "registration":
		return "Free registration may be required to continue reading. The text below may be only a preview."
	case "paywall":
		return "A paid subscription may be required for the full article. The text below may be only a preview."
	default:
		return "The publisher marks this article as access-restricted. The text below may be incomplete."
	}
}

// restrictionSource removes executable JavaScript and CSS before inspecting
// human-facing access requirements. UI code can contain paywall phrases as
// messages or examples without the requested page actually being restricted.
// JSON-LD is retained because isAccessibleForFree is legitimate access metadata.
var scriptBlockRE = regexp.MustCompile(`(?is)<script\b([^>]*)>(.*?)</script\s*>`)
var styleBlockRE = regexp.MustCompile(`(?is)<style\b[^>]*>.*?</style\s*>`)
var htmlCommentRE = regexp.MustCompile(`(?s)<!--.*?-->`)

func restrictionSource(raw string) string {
	kept := make([]string, 0, 2)
	cleaned := scriptBlockRE.ReplaceAllStringFunc(raw, func(block string) string {
		prefixEnd := strings.IndexByte(block, '>')
		if prefixEnd >= 0 {
			startTag := strings.ToLower(block[:prefixEnd+1])
			body := strings.TrimSpace(strings.TrimSuffix(block[prefixEnd+1:], "</script>"))
			// Some publishers emit a bare JSON object in a script element
			// without a type attribute. Only accept syntactically valid JSON
			// containing access-control properties; never scan JS source text.
			var metadata map[string]json.RawMessage
			bareAccessJSON := !strings.Contains(startTag, "type=") &&
				json.Unmarshal([]byte(body), &metadata) == nil &&
				(metadata["isAccessibleForFree"] != nil || metadata["isaccessibleforfree"] != nil || metadata["paywall"] != nil)
			if strings.Contains(startTag, "application/ld+json") || bareAccessJSON {
				kept = append(kept, body)
			}
		}
		return " "
	})
	cleaned = styleBlockRE.ReplaceAllString(cleaned, " ")
	cleaned = htmlCommentRE.ReplaceAllString(cleaned, " ")
	return strings.ToLower(cleaned + " " + strings.Join(kept, " "))
}

// accessRestriction classifies explicit restriction evidence, never HTTP status alone.
// Free registration wins over generic subscription metadata, as in Axios newsletters.
func accessRestriction(raw string) string {
	lower := restrictionSource(raw)
	for _, marker := range []string{
		"sign up for", "register for free to continue reading", "create a free account to continue reading",
	} {
		if marker == "sign up for" {
			// Avoid calling unrelated newsletter advertisements an article registration wall.
			if strings.Contains(lower, "to continue reading for free") || strings.Contains(lower, "to read the rest for free") {
				return "registration"
			}
			continue
		}
		if strings.Contains(lower, marker) {
			return "registration"
		}
	}
	for _, marker := range []string{"paid subscription required", "subscribe for $", "become a paying subscriber"} {
		if strings.Contains(lower, marker) {
			return "paywall"
		}
	}
	for _, marker := range []string{
		"subscribe to continue reading", "subscribe to read the full article",
		"this article is for subscribers only", "subscriber-only article",
		"subscription required to read", `"isaccessibleforfree":false`,
		`"isaccessibleforfree": false`, `"paywall":true`, `"paywall": true`,
	} {
		if strings.Contains(lower, marker) {
			// A generic metadata or subscribe marker does not establish that payment is required.
			return "restricted"
		}
	}
	return ""
}

// restrictionEvidence explains which predefined phrase caused a restriction decision.
// It returns marker names, not article text, to keep server logs predictable.
func restrictionEvidence(raw string) string {
	lower := restrictionSource(raw)
	for _, rule := range []struct{ phrase, label string }{
		{"to continue reading for free", "free-reading-phrase"},
		{"to read the rest for free", "free-reading-phrase"},
		{"register for free to continue reading", "free-registration"},
		{"create a free account to continue reading", "free-account"},
		{"paid subscription required", "explicit-paid-subscription"},
		{"subscribe for $", "explicit-price"},
		{"become a paying subscriber", "explicit-paid-subscriber"},
		{"subscribe to continue reading", "subscription-prompt"},
		{"subscribe to read the full article", "subscription-prompt"},
		{"this article is for subscribers only", "subscriber-only"},
		{"subscriber-only article", "subscriber-only"},
		{"subscription required to read", "subscription-required"},
		{`"isaccessibleforfree":false`, "structured-paywall-metadata"},
		{`"isaccessibleforfree": false`, "structured-paywall-metadata"},
		{`"paywall":true`, "paywall-metadata"},
		{`"paywall": true`, "paywall-metadata"},
	} {
		if strings.Contains(lower, rule.phrase) {
			return rule.label
		}
	}
	return "unidentified"
}

// Backward compatibility for tests and callers expecting a boolean restriction signal.
func paywallEvidence(raw string) bool { return accessRestriction(raw) != "" }
func mustHostname(address string) string {
	u, err := url.Parse(address)
	if err != nil {
		return "publisher"
	}
	return u.Hostname()
}

// resolvePort selects --port / -port over PORT, with 8080 as the default.
// Using a dedicated FlagSet allows precise usage errors and unit testing.
func main() {
	if len(os.Args) == 2 && (os.Args[1] == "--version" || os.Args[1] == "-version") {
		fmt.Println(appDisplayVersion())
		return
	}
	port, err := app.ResolvePort(os.Args[1:], os.Getenv("PORT"), os.Stderr)
	if err == flag.ErrHelp {
		return
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "ScraperGo:", err)
		os.Exit(2)
	}
 readHandler:=newArticleReadHandler()
 err=app.Run(app.Config{Port:port,Version:appVersion,Build:appBuild},app.Handlers{
  Read:readHandler,
  Classify:(&httpapi.Classifier{Read:readHandler,ValidateURL:fetcher.ValidatedURL}).ServeHTTP,
  ImageMetadata:newArticleMetadataHandler(readHandler),
  ArticleImage:(imagedelivery.Handler{Metadata:newArticleMetadataHandler(readHandler)}).ServeHTTP,
  Thumbnail:httpapi.NewThumbnailHandler(fetcher.ValidatedURL,imagedelivery.Handler{Metadata:newArticleMetadataHandler(readHandler)}).ServeHTTP,
  DiagnosticExtract:(diagnostics.HTMLHandler{WriteExtracted:writeExtractedArticle}).ServeHTTP,
  DiagnosticImageTrace:(diagnostics.ImageTraceHandler{Extract:extractWithEvidence}).ServeHTTP,
  TannerPhoto:tannerJPG,
 })
 if err!=nil {os.Exit(1)}
}
