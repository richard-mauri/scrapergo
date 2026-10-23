package main

import (
 "clean-article-reader/internal/extraction"
 "clean-article-reader/internal/pageclassify"
 "clean-article-reader/internal/structured"
 "clean-article-reader/internal/recovery"
	articlemodel "clean-article-reader/internal/article"
 sourcehtml "clean-article-reader/internal/sourcehtml"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// ScraperGo application version is defined once and shared by all presentation paths.
const appVersion = "0.2.85"
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

// inlineTextParts preserves anchor targets in their exact reading position.
// The browser renderer creates safe clickable links from these parts.
func inlineTextParts(n *Node, baseURL string) []TextPart {
	base, _ := url.Parse(baseURL)
	var parts []TextPart
	var walk func(*Node, bool, bool, string)
	walk = func(node *Node, bold, italic bool, href string) {
		if blockedTags[node.Tag] {
			return
		}
		if node.Tag == "#text" {
			t := wsRE.ReplaceAllString(html.UnescapeString(node.Text), " ")
			if t != "" {
				parts = append(parts, TextPart{Text: t, Bold: bold, Italic: italic, URL: href})
			}
			return
		}
		if node.Tag == "strong" || node.Tag == "b" {
			bold = true
		}
		if node.Tag == "em" || node.Tag == "i" {
			italic = true
		}
		if node.Tag == "a" && base != nil {
			if target, err := url.Parse(strings.TrimSpace(node.Attr["href"])); err == nil && target.String() != "" {
				resolved := base.ResolveReference(target)
				if (resolved.Scheme == "https" || resolved.Scheme == "http") && resolved.Host != "" && resolved.User == nil {
					href = resolved.String()
				}
			}
		}
		for _, child := range node.Kids {
			walk(child, bold, italic, href)
		}
	}
	walk(n, false, false, "")
	// Preserve the source punctuation and whitespace around links.
	if len(parts) > 0 {
		parts[0].Text = strings.TrimLeft(parts[0].Text, " ")
		parts[len(parts)-1].Text = strings.TrimRight(parts[len(parts)-1].Text, " ")
	}
	return parts
}

func extractBlocks(root *Node, bases ...string) []Block {
	baseURL := ""
	if len(bases) > 0 {
		baseURL = bases[0]
	}
	var blocks []Block
	var walk func(*Node)
	walk = func(n *Node) {
		if sourcehtml.Noisy(n) {
			return
		}
		if n.Tag == "img" {
			if src := articleImageURL(n, baseURL); src != "" {
				alt := normalize(n.Attr["alt"])
				if len(alt) > 180 {
					alt = alt[:180]
				}
				blocks = append(blocks, Block{Kind: "img", URL: src, Text: alt})
			}
			return
		}
		if n.Tag == "p" || n.Tag == "h2" || n.Tag == "h3" || n.Tag == "h4" || n.Tag == "li" || n.Tag == "blockquote" {
			// Recurse through paragraphs containing photographs, preserving their order.
			hasImage := false
			for _, kid := range n.Kids {
				if kid.Tag == "img" || kid.Tag == "picture" || kid.Tag == "figure" {
					hasImage = true
					break
				}
			}
			if hasImage {
				for _, kid := range n.Kids {
					walk(kid)
				}
				return
			}
			txt := textOf(n)
			if len(txt) > 0 && !extraction.IsAdvertisingText(txt) && !(n.Tag == "p" && len(txt) < 25 && len(find(n, func(child *Node) bool { return child.Tag == "a" })) == 0) {
				blocks = append(blocks, Block{Kind: n.Tag, Text: txt, Parts: inlineTextParts(n, baseURL)})
			}
			return
		}
		for _, kid := range n.Kids {
			walk(kid)
		}
	}
	walk(root)
	seen := map[string]bool{}
	out := []Block{}
	for _, b := range blocks {
		key := strings.ToLower(b.Kind + "|" + b.URL + "|" + b.Text)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, b)
	}
	return out
}

func chooseArticle(root *Node) *Node {
	candidates := find(root, func(n *Node) bool {
		return n.Tag == "article" || n.Tag == "main" || n.Tag == "div" || n.Tag == "section"
	})
	type scored struct {
		n *Node
		s int
	}
	var ranked []scored
	for _, n := range candidates {
		if sourcehtml.Noisy(n) {
			continue
		}
		blocks := extractBlocks(n)
		score := 0
		pars := 0
		for _, b := range blocks {
			if b.Kind == "p" {
				pars++
				score += min(len(b.Text), 750)
			} else if strings.HasPrefix(b.Kind, "h") {
				score += 50
			}
		}
		if pars < 2 {
			continue
		}
		score += pars * 20
		if n.Tag == "article" {
			score += 1700
		}
		if n.Tag == "main" {
			score += 300
		}
		if strings.Contains(strings.ToLower(n.Attr["class"]+" "+n.Attr["id"]), "article") {
			score += 300
		}
		ranked = append(ranked, scored{n, score})
	}
	if len(ranked) == 0 {
		return root
	}
	sort.SliceStable(ranked, func(i, j int) bool { return ranked[i].s > ranked[j].s })
	return ranked[0].n
}

// Compatibility entrypoint for extraction callers during the main-package migration.
func recoverArticlePhotographs(a *Article, raw, baseURL string) { recovery.RecoverArticlePhotographs(a,raw,baseURL) }

// JSON-LD often contains a clean articleBody even if visible markup is wrapped
// in advertisement-related containers. Prefer the normal page body when viable.
var ldScriptRE = regexp.MustCompile(`(?is)<script\b([^>]*)>(.*?)</script\s*>`)

func jsonLDArticle(raw string) []Block {
	for _, m := range ldScriptRE.FindAllStringSubmatch(raw, -1) {
		if !strings.Contains(strings.ToLower(m[1]), "application/ld+json") {
			continue
		}
		var v any
		if json.Unmarshal([]byte(m[2]), &v) != nil {
			continue
		}
		if body := findLDArticleBody(v); len(body) >= 250 {
			parts := regexp.MustCompile(`\n\s*\n+`).Split(strings.TrimSpace(body), -1)
			out := make([]Block, 0, len(parts))
			for _, part := range parts {
				txt := normalize(regexp.MustCompile(`<[^>]*>`).ReplaceAllString(part, " "))
				if len(txt) > 25 && !extraction.IsAdvertisingText(txt) {
					out = append(out, Block{Kind: "p", Text: txt})
				}
			}
			if len(out) > 0 {
				return out
			}
		}
	}
	return nil
}
func findLDArticleBody(v any) string {
	switch x := v.(type) {
	case []any:
		for _, item := range x {
			if s := findLDArticleBody(item); s != "" {
				return s
			}
		}
	case map[string]any:
		ty := fmt.Sprint(x["@type"])
		if strings.Contains(strings.ToLower(ty), "article") {
			if body, ok := x["articleBody"].(string); ok {
				return body
			}
		}
		for _, key := range []string{"@graph", "mainEntity", "itemListElement"} {
			if s := findLDArticleBody(x[key]); s != "" {
				return s
			}
		}
	}
	return ""
}
func extract(raw, src string) (Article, error) {
	root := parseHTML(raw)
	a := Article{URL: src, Title: meta(root, "og:title", "twitter:title"), Byline: meta(root, "author", "article:author"), Published: meta(root, "article:published_time", "datepublished", "pubdate"), Site: meta(root, "og:site_name")}
	if a.Title == "" {
		titles := find(root, func(n *Node) bool { return n.Tag == "h1" })
		if len(titles) > 0 {
			a.Title = textOf(titles[0])
		}
	}
	if a.Title == "" {
		titles := find(root, func(n *Node) bool { return n.Tag == "title" })
		if len(titles) > 0 {
			a.Title = textOf(titles[0])
		}
	}
	if a.Title == "" {
		a.Title = "Untitled article"
	}
	target := chooseArticle(root)
	a.Blocks = extractBlocks(target, src)
	if len(a.Blocks) < 3 {
		if ld := jsonLDArticle(raw); len(ld) > 0 {
			a.Blocks = ld
		}
	}
	// Discard obvious footer-like text even inside article containers.
	cleaned := []Block{}
	for _, b := range a.Blocks {
		if strings.HasPrefix(strings.ToLower(b.Text), "contact ") || strings.HasPrefix(strings.ToLower(b.Text), "all articles →") {
			continue
		}
		cleaned = append(cleaned, b)
		a.WordCount += len(strings.Fields(b.Text))
	}
	a.Blocks = cleaned
	if a.WordCount < 60 {
		return a, fmt.Errorf("Found only %d article words in %d bytes of HTML. The page may be blocking automated access, loading content with JavaScript, or using unsupported markup", a.WordCount, len(raw))
	}
	return a, nil
}
func safeHost(ctx context.Context, host string) error {
	if host == "" {
		return errors.New("missing hostname")
	}
	if strings.EqualFold(host, "localhost") || strings.HasSuffix(strings.ToLower(host), ".localhost") {
		return errors.New("local addresses are not allowed")
	}
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return fmt.Errorf("DNS lookup failed: %w", err)
	}
	for _, ip := range ips {
		v := ip.IP
		if !v.IsGlobalUnicast() || v.IsPrivate() || v.IsLoopback() || v.IsLinkLocalUnicast() || v.IsLinkLocalMulticast() {
			return errors.New("private or reserved network addresses are not allowed")
		}
	}
	return nil
}
func validatedURL(ctx context.Context, s string) (*url.URL, error) {
	u, err := url.Parse(strings.TrimSpace(s))
	if err != nil {
		return nil, err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, errors.New("only HTTP and HTTPS URLs are supported")
	}
	if u.User != nil {
		return nil, errors.New("URL credentials are not supported")
	}
	if err := safeHost(ctx, u.Hostname()); err != nil {
		return nil, err
	}
	return u, nil
}

// setBrowserHeaders uses ordinary Chrome-style navigation headers. These can
// improve compatibility, but cannot reproduce Safari's TLS fingerprint or cookies.
func setBrowserHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.8")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
}

// setReadeckLikeHeaders provides a second, internally consistent Chromium
// request profile. This is an independently implemented approximation, not
// Readeck code, and does not reproduce its TLS fingerprint.
func setReadeckLikeHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/jpeg,image/png,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Language", "en-US,en;q=0.8")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-CH-UA", `"Google Chrome";v="144", "Chromium";v="144"`)
	req.Header.Set("Sec-CH-UA-Mobile", "?0")
	req.Header.Set("Sec-CH-UA-Platform", `"Windows"`)
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
}

// alternateHeaderRetry applies only to a publisher's explicit refusal.
// A second 403 is returned normally rather than classified as a paywall.
func alternateHeaderRetry(status int) bool { return status == http.StatusForbidden }

func download(ctx context.Context, address string) (string, string, error) {
	u, err := validatedURL(ctx, address)
	if err != nil {
		return "", "", err
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}
		if err := safeHost(ctx, host); err != nil {
			return nil, err
		}
		return (&net.Dialer{Timeout: 8 * time.Second}).DialContext(ctx, network, net.JoinHostPort(host, port))
	}
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Timeout: 18 * time.Second, Transport: transport, Jar: jar, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return errors.New("too many redirects")
		}
		_, err := validatedURL(req.Context(), req.URL.String())
		return err
	}}
	// The initial attempt preserves v0.2.14 behavior. Only a 403 triggers
	// one alternative request, with a fresh cookie jar and no publisher tab.
	started := time.Now()
	var res *http.Response
	for attempt := 0; attempt < 2; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			return "", "", err
		}
		profile := "existing"
		if attempt == 0 {
			setBrowserHeaders(req)
		} else {
			profile = "readeck-like"
			setReadeckLikeHeaders(req)
		}
		res, err = client.Do(req)
		if err != nil {
			log.Printf("fetch host=%s profile=%s error=%v elapsed=%s", u.Hostname(), profile, err, time.Since(started))
			return "", "", err
		}
		log.Printf("fetch host=%s profile=%s status=%d redirect_host=%s content_type=%q elapsed=%s",
			u.Hostname(), profile, res.StatusCode, res.Request.URL.Hostname(), res.Header.Get("Content-Type"), time.Since(started))
		if attempt == 0 && alternateHeaderRetry(res.StatusCode) {
			_ = res.Body.Close()
			// A retry should not carry over cookies from the rejected response.
			freshJar, _ := cookiejar.New(nil)
			client.Jar = freshJar
			continue
		}
		break
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", "", &publisherHTTPError{Host: res.Request.URL.Hostname(), StatusCode: res.StatusCode}
	}
	if c := res.Header.Get("Content-Type"); c != "" && !strings.Contains(strings.ToLower(c), "html") {
		return "", "", errors.New("URL did not return an HTML page")
	}
	data, err := io.ReadAll(io.LimitReader(res.Body, maxHTML+1))
	if err != nil {
		return "", "", err
	}
	if len(data) > maxHTML {
		return "", "", errors.New("page exceeds 5 MB limit")
	}
	return string(data), res.Request.URL.String(), nil
}

// publisherHTTPError preserves the HTTP status without assuming why access was denied.
type publisherHTTPError struct {
	Host       string
	StatusCode int
}

func (e *publisherHTTPError) Error() string {
	return fmt.Sprintf("article request to %s returned HTTP %d (%s)", e.Host, e.StatusCode, http.StatusText(e.StatusCode))
}

// A 401 or 403 is an access denial, not proof of a subscription requirement.
func publisherAccessMessage(err error) (string, bool) {
	var denied *publisherHTTPError
	if errors.As(err, &denied) && (denied.StatusCode == http.StatusUnauthorized || denied.StatusCode == http.StatusForbidden) {
		return "Publisher access restricted. The publisher declined ScraperGo's request. Open the original website to view any available preview, sign-in, or subscription options.", true
	}
	return "", false
}

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
		article = assembleArticleDocument(extracted, cleanedArticleHTML)
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
func resolvePort(args []string, envPort string, output io.Writer) (string, error) {
	fs := flag.NewFlagSet("scrapergo", flag.ContinueOnError)
	fs.SetOutput(output)
	fs.Usage = func() {
		fmt.Fprintln(output, "Usage: scrapergo [--port PORT] [--help]")
		fmt.Fprintln(output, "\nOptions:")
		fs.PrintDefaults()
		fmt.Fprintln(output, "\nExamples:")
		fmt.Fprintln(output, "  scrapergo --port 8081")
		fmt.Fprintln(output, "  ./dist/scrapergo-darwin-arm64 --port 8081")
		fmt.Fprintln(output, "  PORT=8081 scrapergo")
		fmt.Fprintln(output, "\n--port overrides the PORT environment variable. Default: 8080.")
	}
	cliPort := fs.String("port", "", "HTTP listening port (1-65535)")
	if err := fs.Parse(args); err != nil {
		return "", err
	}
	if fs.NArg() != 0 {
		return "", fmt.Errorf("unexpected argument %q (use --help)", fs.Arg(0))
	}
	port := "8080"
	if envPort != "" {
		port = envPort
	}
	if *cliPort != "" {
		port = *cliPort
	}
	n, err := strconv.Atoi(port)
	if err != nil || n < 1 || n > 65535 {
		return "", fmt.Errorf("invalid port %q: specify a number from 1 to 65535 (example: --port 8081)", port)
	}
	return strconv.Itoa(n), nil
}

func main() {
	port, err := resolvePort(os.Args[1:], os.Getenv("PORT"), os.Stderr)
	if err == flag.ErrHelp {
		return
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "ScraperGo:", err)
		os.Exit(2)
	}
	mux := http.NewServeMux()
	readHandler := newArticleReadHandler()
	mux.Handle("/api/read", readHandler)
	mux.HandleFunc("/api/classify", classifyArticle)
	mux.Handle("/api/image-metadata", newArticleMetadataHandler(readHandler))
	mux.HandleFunc("/api/article-image", articleImage)
	mux.HandleFunc("/api/thumbnail", articleThumbnail)
	mux.HandleFunc("/api/diagnostic/extract", diagnosticExtract)
	mux.HandleFunc("/api/diagnostic/image-trace", diagnosticImageTrace)
	mux.HandleFunc("/tanner.jpg", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		w.Write(tannerJPG)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' https: data:; style-src 'unsafe-inline'; script-src 'unsafe-inline'; base-uri 'none'; object-src 'none'")
		fmt.Fprint(w, renderedIndexHTML())
	})
	server := &http.Server{Addr: ":" + port, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	log.Printf("ScraperGo v%s listening on http://localhost:%s", appVersion, port)
	if err := server.ListenAndServe(); err != nil {
		if errors.Is(err, syscall.EADDRINUSE) {
			log.Printf("Port %s is already in use. Try a different port:\n  go run . --port 8081\n  ./dist/scrapergo-darwin-amd64 --port 8081\nOr set PORT=8081 before launching.", port)
		} else {
			log.Printf("HTTP server failed: %v", err)
		}
		os.Exit(1)
	}
}
