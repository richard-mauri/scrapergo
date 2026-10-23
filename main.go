package main

import (
 "clean-article-reader/internal/accesspolicy"
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
	kind := accesspolicy.Restriction(raw)
	cleanedArticleHTML := structured.StripEmbeddedMedia(raw)
	extracted, extractionErr := extractWithEvidence(cleanedArticleHTML, finalURL)
	article := extracted.Article
	if extractionErr == nil {
		article = reader.DefaultProcessor().Process(extracted, cleanedArticleHTML)
		logSlateExtractionDiagnostic(article.Blocks, finalURL)
	}
	if kind != "" {
		log.Printf("restriction publisher=%s kind=%s evidence=%s html_bytes=%d public_words=%d", accesspolicy.Hostname(finalURL), kind, accesspolicy.Evidence(raw), len(raw), article.WordCount)
		if extractionErr == nil && article.WordCount >= 80 {
			article.AccessType = kind
			article.AccessNote = accesspolicy.Notice(kind)
			_ = json.NewEncoder(w).Encode(article)
			return
		}
		w.WriteHeader(422)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": accesspolicy.Notice(kind), "kind": kind, "publisher": accesspolicy.Hostname(finalURL), "evidence": accesspolicy.Evidence(raw)})
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
