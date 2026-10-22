package main

import (
    "net/url"
    "strings"
)

// partialHTMLArticle recognizes a *short, server-delivered* article excerpt.
// It is not a publisher adapter: it requires corroborating page metadata,
// a matching headline, and coherent prose within an article element.
// It never fetches additional content or attempts to defeat access controls.
func partialHTMLArticle(raw, source string) (Article, bool) {
    base, err := url.Parse(source)
    if err != nil || base.Hostname() == "" || (base.Scheme != "https" && base.Scheme != "http") {
        return Article{}, false
    }
    root := parseHTML(raw)
    if !strings.EqualFold(strings.TrimSpace(meta(root, "og:type")), "article") {
        return Article{}, false
    }
    title := normalize(meta(root, "og:title", "twitter:title"))
    if len(title) < 12 {
        return Article{}, false
    }
    canonicalMatches := false
    for _, n := range find(root, func(n *Node) bool { return n.Tag == "link" }) {
        if !strings.EqualFold(n.Attr["rel"], "canonical") {
            continue
        }
        candidate, err := url.Parse(strings.TrimSpace(n.Attr["href"]))
        if err == nil {
            candidate = base.ResolveReference(candidate)
            canonicalMatches = strings.EqualFold(candidate.Hostname(), base.Hostname()) &&
                strings.TrimRight(candidate.EscapedPath(), "/") == strings.TrimRight(base.EscapedPath(), "/")
        }
        break
    }
    if !canonicalMatches {
        return Article{}, false
    }
    headlineMatches := false
    for _, h := range find(root, func(n *Node) bool { return n.Tag == "h1" }) {
        headline := strings.ToLower(normalize(textOf(h)))
        titleKey := strings.ToLower(title)
        if len(headline) >= 12 && (strings.Contains(titleKey, headline) || strings.Contains(headline, titleKey)) {
            headlineMatches = true
            break
        }
    }
    if !headlineMatches {
        return Article{}, false
    }
    var best []Block
    bestCount := 0
    for _, article := range find(root, func(n *Node) bool { return n.Tag == "article" }) {
        blocks := make([]Block, 0, 8)
        words := 0
        for _, p := range find(article, func(n *Node) bool { return n.Tag == "p" }) {
            excluded := false
            for ancestor := p.Parent; ancestor != nil && ancestor != article; ancestor = ancestor.Parent {
                if ancestor.Tag == "figure" || ancestor.Tag == "figcaption" ||
                    ancestor.Tag == "nav" || ancestor.Tag == "footer" || ancestor.Tag == "aside" {
                    excluded = true
                    break
                }
            }
            if excluded {
                continue
            }
            line := normalize(textOf(p))
            if len(line) < 12 || sentenceAdRE.MatchString(line) {
                continue
            }
            blocks = append(blocks, Block{Kind: "p", Text: line})
            words += len(strings.Fields(line))
            if len(blocks) >= 32 {
                break
            }
        }
        if words > bestCount {
            best, bestCount = blocks, words
        }
    }
    // This path is deliberately limited to substantial *partial* excerpts.
    // Full articles continue to use Readability/structured recovery.
    if bestCount < 30 || bestCount >= 60 {
        return Article{}, false
    }
    img := resolveImageURL(meta(root, "og:image", "twitter:image"), base)
    return Article{
        Title: title, URL: source, Blocks: best, WordCount: bestCount,
        ImageURL: img, Method: "Partial HTML excerpt",
        AccessNote: "Only an excerpt was available to ScraperGo. The original website may require cookie consent, registration, or a subscription to display the complete article. Open the original article to continue reading.",
    }, true
}

// hasPartialArticleEvidence is shared by the pre- and post-extraction index gates.
func hasPartialArticleEvidence(raw, source string) bool {
    _, ok := partialHTMLArticle(raw, source)
    return ok
}
