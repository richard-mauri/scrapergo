package extraction

import (
 "clean-article-reader/internal/article"
	"fmt"
	"net/url"
	"strings"
	"testing"
)

func TestGenericRecoveryHandlesRankingWithoutSpecialCase(t *testing.T) {
	var raw strings.Builder
	raw.WriteString(`<main><p>Introductory discussion of this important football weekend and the remarkable set of results.</p><ul><li>Five of the top seven teams lost on Saturday in remarkable upsets.</li></ul>`)
	for n := 25; n > 0; n-- {
		fmt.Fprintf(&raw, `<div><h2>%d. University %d</h2><img src="/logo/%d.png" alt="University %d"><p>Record: 6-0</p><p>University %d had an outstanding week on the field with meaningful achievements.</p><img src="/photos/%d.jpg" alt="Action photograph"><p>Game summary and commentary describing the match in a full paragraph of prose.</p></div>`, n, n, n, n, n, n)
	}
	raw.WriteString(`</main>`)
	u, _ := url.Parse("https://example.com/sports/rankings")
	got, ok := RecoverStructuredDocument(raw.String(), []article.Block{{Kind: "p", Text: "25. University 25"}}, u)
	if !ok {
		t.Fatal("did not recover ranking")
	}
	var headings, photos, bullets int
	for _, b := range got {
		if b.Kind == "h2" {
			headings++
		}
		if b.Kind == "img" && strings.Contains(b.URL, "/photos/") {
			photos++
		}
		if b.Kind == "li" {
			bullets++
		}
	}
	if headings != 25 || photos != 25 || bullets != 1 {
		t.Fatalf("got heads %d photos %d bullets %d", headings, photos, bullets)
	}
}
