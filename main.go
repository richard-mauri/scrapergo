package main

import (
	"context"
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

// ScraperGo v0.2.21 — browser-style retrieval and diagnostics.
const sampleURL = "https://247wallst.com/investing/2026/10/16/taiwan-is-planning-4-billion-industrial-parks-in-texas-for-nvidias-suppliers/"
const maxHTML = 5 << 20

type InlineLink struct {
	Text string `json:"text"`
	URL  string `json:"url"`
}
type Block struct {
	Kind  string       `json:"kind"`
	Links []InlineLink `json:"links,omitempty"`
	Text  string       `json:"text"`
	URL   string       `json:"url,omitempty"`
}
type Article struct {
	Title      string  `json:"title"`
	Byline     string  `json:"byline,omitempty"`
	Published  string  `json:"published,omitempty"`
	Site       string  `json:"site,omitempty"`
	URL        string  `json:"url"`
	Blocks     []Block `json:"blocks"`
	WordCount  int     `json:"wordCount"`
	ImageURL   string  `json:"imageUrl,omitempty"`
	Method     string  `json:"method,omitempty"`
	AccessNote string  `json:"accessNote,omitempty"`
	AccessType string  `json:"accessType,omitempty"`
}
type Node struct {
	Tag    string
	Attr   map[string]string
	Parent *Node
	Kids   []*Node
	Text   string
}

var tagRE = regexp.MustCompile(`(?s)<!--.*?-->|<![^>]*>|<[^>]*>`)
var attrRE = regexp.MustCompile(`(?i)([a-zA-Z_:][-a-zA-Z0-9_:.]*)\s*=\s*(?:"([^"]*)"|'([^']*)'|([^\s>]+))`)
var wsRE = regexp.MustCompile(`\s+`)
var promotRE = regexp.MustCompile(`(?i)(advertis|sponsor|promo|newsletter|signup|sign-up|subscribe|paywall|related|recommended|trending|most-read|share[-_ ]|social|ticker|stock[-_ ]?quote|market[-_ ]?data|author[-_ ]?bio|affiliate|outbrain|taboola|comments|breadcrumb|sidebar|sticky|modal|popup|cookie|footer|navigation|nav-menu|more-content|continue-reading|biggest-winner|biggest-loser)`)
var sentenceAdRE = regexp.MustCompile(`(?i)(enter your email|free report|stocks to buy right now|our analysts combed|click here to|grab your free|sign up for|subscribe to|advertisement|read more:|get our free|limited.time offer)`)
var voidTags = map[string]bool{"meta": true, "link": true, "img": true, "input": true, "br": true, "hr": true, "source": true, "area": true, "base": true, "embed": true, "wbr": true}
var blockedTags = map[string]bool{"script": true, "style": true, "noscript": true, "iframe": true, "svg": true, "form": true, "button": true, "nav": true, "footer": true, "aside": true, "header": true}

func normalize(s string) string {
	return strings.TrimSpace(wsRE.ReplaceAllString(html.UnescapeString(s), " "))
}
func parseHTML(raw string) *Node {
	root := &Node{Tag: "root"}
	stack := []*Node{root}
	last := 0
	for _, loc := range tagRE.FindAllStringIndex(raw, -1) {
		if loc[0] > last {
			stack[len(stack)-1].Kids = append(stack[len(stack)-1].Kids, &Node{Tag: "#text", Text: raw[last:loc[0]], Parent: stack[len(stack)-1]})
		}
		tok := raw[loc[0]:loc[1]]
		last = loc[1]
		if strings.HasPrefix(tok, "<!--") || strings.HasPrefix(tok, "<!") {
			continue
		}
		if strings.HasPrefix(tok, "</") {
			tag := strings.ToLower(strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(tok, "</"), ">")))
			tag = strings.Fields(tag)[0]
			for j := len(stack) - 1; j > 0; j-- {
				if stack[j].Tag == tag {
					stack = stack[:j]
					break
				}
			}
			continue
		}
		body := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(tok, "<"), ">"))
		if body == "" {
			continue
		}
		parts := strings.Fields(body)
		if len(parts) == 0 {
			continue
		}
		tag := strings.ToLower(strings.TrimRight(parts[0], "/"))
		if !regexp.MustCompile(`^[a-z][a-z0-9:-]*$`).MatchString(tag) {
			continue
		}
		p := stack[len(stack)-1]
		node := &Node{Tag: tag, Parent: p, Attr: map[string]string{}}
		for _, m := range attrRE.FindAllStringSubmatch(body, -1) {
			val := m[2]
			if val == "" {
				val = m[3]
			}
			if val == "" {
				val = m[4]
			}
			node.Attr[strings.ToLower(m[1])] = html.UnescapeString(val)
		}
		p.Kids = append(p.Kids, node)
		if !voidTags[tag] && !strings.HasSuffix(body, "/") {
			stack = append(stack, node)
		}
	}
	if last < len(raw) {
		p := stack[len(stack)-1]
		p.Kids = append(p.Kids, &Node{Tag: "#text", Text: raw[last:], Parent: p})
	}
	return root
}
func find(root *Node, test func(*Node) bool) []*Node {
	var out []*Node
	var walk func(*Node)
	walk = func(n *Node) {
		if test(n) {
			out = append(out, n)
		}
		for _, k := range n.Kids {
			walk(k)
		}
	}
	walk(root)
	return out
}
func textOf(n *Node) string {
	if n == nil {
		return ""
	}
	if n.Tag == "#text" {
		return n.Text
	}
	if blockedTags[n.Tag] {
		return ""
	}
	var b strings.Builder
	for _, c := range n.Kids {
		b.WriteString(textOf(c))
		b.WriteByte(' ')
	}
	return normalize(b.String())
}
func meta(root *Node, keys ...string) string {
	for _, n := range find(root, func(n *Node) bool { return n.Tag == "meta" }) {
		key := strings.ToLower(n.Attr["property"])
		if key == "" {
			key = strings.ToLower(n.Attr["name"])
		}
		for _, wanted := range keys {
			if key == wanted && n.Attr["content"] != "" {
				return normalize(n.Attr["content"])
			}
		}
	}
	return ""
}
func noisy(n *Node) bool {
	if blockedTags[n.Tag] {
		return true
	}
	s := n.Attr["class"] + " " + n.Attr["id"] + " " + n.Attr["role"] + " " + n.Attr["aria-label"]
	return promotRE.MatchString(s)
}
func inNoisy(n, stop *Node) bool {
	for p := n; p != nil && p != stop; p = p.Parent {
		if noisy(p) {
			return true
		}
	}
	return false
}
func extractBlocks(root *Node) []Block {
	var blocks []Block
	var walk func(*Node)
	walk = func(n *Node) {
		if noisy(n) {
			return
		}
		if n.Tag == "p" || n.Tag == "h2" || n.Tag == "h3" || n.Tag == "h4" || n.Tag == "li" || n.Tag == "blockquote" {
			s := textOf(n)
			kind := n.Tag
			if len(s) > 0 && !sentenceAdRE.MatchString(s) && !(n.Tag == "p" && len(s) < 25) {
				blocks = append(blocks, Block{Kind: kind, Text: s})
			}
			return
		}
		for _, k := range n.Kids {
			walk(k)
		}
	}
	walk(root)
	// Deduplicate common mirrored intro and marketing blocks.
	seen := map[string]bool{}
	out := []Block{}
	for _, b := range blocks {
		key := strings.ToLower(b.Text)
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
		if noisy(n) {
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
				if len(txt) > 25 && !sentenceAdRE.MatchString(txt) {
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
	a.Blocks = extractBlocks(target)
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
		return "", "", fmt.Errorf("article request to %s returned HTTP %d (%s)", res.Request.URL.Hostname(), res.StatusCode, http.StatusText(res.StatusCode))
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
func api(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "POST required", 405)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	var in struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "Invalid request JSON", 400)
		return
	}
	address, err := resolveGoogleNewsURL(r.Context(), in.URL)
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(422)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	log.Printf("resolve publisher=%s google_news=%t", mustHostname(address), strings.Contains(in.URL, "news.google.com/"))
	raw, finalURL, err := download(r.Context(), address)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	if err != nil {
		w.WriteHeader(422)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error(), "kind": "access_error"})
		return
	}
	// Extract before deciding whether restricted pages offer readable public text.
	// Restriction metadata alone does not establish that no usable preview exists.
	kind := accessRestriction(raw)
	article, extractionErr := extractWithReadability(raw, finalURL)
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
	if extractionErr != nil {
		w.WriteHeader(422)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": extractionErr.Error(), "kind": "extraction_error"})
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

// accessRestriction classifies explicit restriction evidence, never HTTP status alone.
// Free registration wins over generic subscription metadata, as in Axios newsletters.
func accessRestriction(raw string) string {
	lower := strings.ToLower(raw)
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
	lower := strings.ToLower(raw)
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
	mux.HandleFunc("/api/read", api)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' https: data:; style-src 'unsafe-inline'; script-src 'unsafe-inline'; base-uri 'none'; object-src 'none'")
		fmt.Fprint(w, indexHTML)
	})
	server := &http.Server{Addr: ":" + port, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	log.Printf("ScraperGo v0.2.21 listening on http://localhost:%s", port)
	if err := server.ListenAndServe(); err != nil {
		if errors.Is(err, syscall.EADDRINUSE) {
			log.Printf("Port %s is already in use. Try a different port:\n  go run . --port 8081\n  ./dist/scrapergo-darwin-amd64 --port 8081\nOr set PORT=8081 before launching.", port)
		} else {
			log.Printf("HTTP server failed: %v", err)
		}
		os.Exit(1)
	}
}
