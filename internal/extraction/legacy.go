package extraction

import (
 "clean-article-reader/internal/article"
 "clean-article-reader/internal/sourcehtml"
 "encoding/json"
 "fmt"
 "html"
 "net/url"
 "regexp"
 "sort"
 "strings"
)

var legacyWhitespace = regexp.MustCompile(`\s+`)

// inlineTextParts preserves anchor targets in their exact reading position.
// The browser renderer creates safe clickable links from these parts.
func InlineTextParts(n *sourcehtml.Node, baseURL string) []article.TextPart {
	base, _ := url.Parse(baseURL)
	var parts []article.TextPart
	var walk func(*sourcehtml.Node, bool, bool, string)
	walk = func(node *sourcehtml.Node, bold, italic bool, href string) {
		if sourcehtml.BlockedTags[node.Tag] {
			return
		}
		if node.Tag == "#text" {
			t := legacyWhitespace.ReplaceAllString(html.UnescapeString(node.Text), " ")
			if t != "" {
				parts = append(parts, article.TextPart{Text: t, Bold: bold, Italic: italic, URL: href})
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

func ExtractBlocks(root *sourcehtml.Node, bases ...string) []article.Block {
	baseURL := ""
	if len(bases) > 0 {
		baseURL = bases[0]
	}
	var blocks []article.Block
	var walk func(*sourcehtml.Node)
	walk = func(n *sourcehtml.Node) {
		if sourcehtml.Noisy(n) {
			return
		}
		if n.Tag == "img" {
			if src := sourcehtml.ImageURL(n, baseURL); src != "" {
				alt := sourcehtml.Normalize(n.Attr["alt"])
				if len(alt) > 180 {
					alt = alt[:180]
				}
				blocks = append(blocks, article.Block{Kind: "img", URL: src, Text: alt})
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
			txt := sourcehtml.Text(n)
			if len(txt) > 0 && !IsAdvertisingText(txt) && !(n.Tag == "p" && len(txt) < 25 && len(sourcehtml.Find(n, func(child *sourcehtml.Node) bool { return child.Tag == "a" })) == 0) {
				blocks = append(blocks, article.Block{Kind: n.Tag, Text: txt, Parts: InlineTextParts(n, baseURL)})
			}
			return
		}
		for _, kid := range n.Kids {
			walk(kid)
		}
	}
	walk(root)
	seen := map[string]bool{}
	out := []article.Block{}
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

func ChooseArticle(root *sourcehtml.Node) *sourcehtml.Node {
	candidates := sourcehtml.Find(root, func(n *sourcehtml.Node) bool {
		return n.Tag == "article" || n.Tag == "main" || n.Tag == "div" || n.Tag == "section"
	})
	type scored struct {
		n *sourcehtml.Node
		s int
	}
	var ranked []scored
	for _, n := range candidates {
		if sourcehtml.Noisy(n) {
			continue
		}
		blocks := ExtractBlocks(n)
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

func JSONLDArticle(raw string) []article.Block {
	for _, m := range ldScriptRE.FindAllStringSubmatch(raw, -1) {
		if !strings.Contains(strings.ToLower(m[1]), "application/ld+json") {
			continue
		}
		var v any
		if json.Unmarshal([]byte(m[2]), &v) != nil {
			continue
		}
		if body := FindLDArticleBody(v); len(body) >= 250 {
			parts := regexp.MustCompile(`\n\s*\n+`).Split(strings.TrimSpace(body), -1)
			out := make([]article.Block, 0, len(parts))
			for _, part := range parts {
				txt := sourcehtml.Normalize(regexp.MustCompile(`<[^>]*>`).ReplaceAllString(part, " "))
				if len(txt) > 25 && !IsAdvertisingText(txt) {
					out = append(out, article.Block{Kind: "p", Text: txt})
				}
			}
			if len(out) > 0 {
				return out
			}
		}
	}
	return nil
}
func FindLDArticleBody(v any) string {
	switch x := v.(type) {
	case []any:
		for _, item := range x {
			if s := FindLDArticleBody(item); s != "" {
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
			if s := FindLDArticleBody(x[key]); s != "" {
				return s
			}
		}
	}
	return ""
}
func LegacyExtract(raw, src string) (article.Article, error) {
	root := sourcehtml.Parse(raw)
	a := article.Article{URL: src, Title: sourcehtml.Meta(root, "og:title", "twitter:title"), Byline: sourcehtml.Meta(root, "author", "article:author"), Published: sourcehtml.Meta(root, "article:published_time", "datepublished", "pubdate"), Site: sourcehtml.Meta(root, "og:site_name")}
	if a.Title == "" {
		titles := sourcehtml.Find(root, func(n *sourcehtml.Node) bool { return n.Tag == "h1" })
		if len(titles) > 0 {
			a.Title = sourcehtml.Text(titles[0])
		}
	}
	if a.Title == "" {
		titles := sourcehtml.Find(root, func(n *sourcehtml.Node) bool { return n.Tag == "title" })
		if len(titles) > 0 {
			a.Title = sourcehtml.Text(titles[0])
		}
	}
	if a.Title == "" {
		a.Title = "Untitled article"
	}
	target := ChooseArticle(root)
	a.Blocks = ExtractBlocks(target, src)
	if len(a.Blocks) < 3 {
		if ld := JSONLDArticle(raw); len(ld) > 0 {
			a.Blocks = ld
		}
	}
	// Discard obvious footer-like text even inside article containers.
	cleaned := []article.Block{}
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
