package main

import (
	"fmt"
	"net/url"
	"strings"
	"testing"
)

func TestContextualTopicIndexAndPromo(t *testing.T) {
	var b strings.Builder
	b.WriteString(`<main><article><p>Long introduction to the editorial report with context and explanation for the reader.</p>`)
	for i := 1; i <= 8; i++ {
		fmt.Fprintf(&b, `<section><h2>Report section %d</h2><p>The complete editorial assessment for section %d gives readers the facts and context needed for this study.</p></section>`, i, i)
	}
	b.WriteString(`<p>Final editorial analysis describes the principal result and provides closing context for the reader.</p><section><h3>In this topic</h3><div class="chips">`)
	for i := 0; i < 9; i++ {
		fmt.Fprintf(&b, `<a href="/topics/%d"><img src="/icons/%d.png" alt="Topic %d">Topic %d</a>`, i, i, i, i)
	}
	b.WriteString(`</div></section><section class="next-story"><h3>Next Story</h3><img src="/teaser.jpg" alt="Next Story Image"><h2>Promotional headline is not the original article</h2></section></article></main>`)
	base, _ := url.Parse("https://example.org/story")
	blocks, ok := recoverStructuredDocument(b.String(), []Block{{Kind: "p", Text: "intro"}}, base)
	if !ok {
		t.Fatal("recovery failed")
	}
	title, links, noise, ending := 0, 0, 0, 0
	for _, v := range blocks {
		switch v.Kind {
		case "topic-title":
			title++
		case "topic-link":
			links++
		}
		if strings.Contains(v.Text, "Promotional headline") || strings.Contains(v.Text, "Next Story") || strings.Contains(v.URL, "teaser.jpg") {
			noise++
		}
		if strings.Contains(v.Text, "Final editorial analysis") {
			ending++
		}
	}
	if title != 1 || links != 9 || noise != 0 || ending != 1 {
		t.Fatalf("title=%d links=%d noise=%d final=%d", title, links, noise, ending)
	}
}
