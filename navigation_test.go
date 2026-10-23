package main

import (
 "clean-article-reader/internal/extraction"
	"fmt"
	"net/url"
	"strings"
	"testing"
)

func TestGenericRecoveryDropsTopicNavigationButKeepsFinalProse(t *testing.T) {
	var b strings.Builder
	b.WriteString(`<main><article><p>This extensive report provides an overview of the results and their implications for every participant.</p>`)
	for i := 1; i <= 8; i++ {
		fmt.Fprintf(&b, `<section><h2>Entry %d heading</h2><p>Detailed editorial analysis of item %d explains the outcome and underlying reasons across multiple sentences.</p></section>`, i, i)
	}
	b.WriteString(`<p>The final editorial paragraph summarizes the significance of these results and closes the report clearly.</p>`)
	b.WriteString(`<section><h3>Explore further</h3><div class="topics">`)
	for i := 0; i < 18; i++ {
		fmt.Fprintf(&b, `<a href="/topic/%d"><img src="/icon/%d.png" alt="Topic %d">Topic %d</a>`, i, i, i, i)
	}
	b.WriteString(`</div></section></article></main>`)
	base, _ := url.Parse("https://example.org/report")
	got, ok := extraction.RecoverStructuredDocument(b.String(), []Block{{Kind: "p", Text: "Entry 1 heading"}}, base)
	if !ok {
		t.Fatal("did not recover article")
	}
	joined := ""
	imgs := 0
	for _, x := range got {
		joined += "\n" + x.Text
		if x.Kind == "img" {
			imgs++
		}
	}
	if !strings.Contains(joined, "final editorial paragraph") {
		t.Fatal("lost final editorial paragraph")
	}
	if strings.Contains(joined, "Topic 12") || imgs != 0 {
		t.Fatalf("navigation leaked: images %d", imgs)
	}
}
func TestNavigationClusterPreservesEditorialLinkList(t *testing.T) {
	raw := `<ul><li><a href="/one">Resource one</a> provides a detailed technical analysis with comprehensive guidance and several specific implementation examples.</li><li><a href="/two">Resource two</a> contains an extensive review covering performance requirements and detailed practical limitations.</li><li><a href="/three">Resource three</a> provides comprehensive comparison data and background interpretation with extensive explanation.</li><li><a href="/four">Resource four</a> discusses numerous safety considerations and special operating circumstances in depth.</li><li><a href="/five">Resource five</a> explains the underlying engineering methods and discusses design constraints.</li></ul>`
	n := find(parseHTML(raw), func(n *Node) bool { return n.Tag == "ul" })[0]
	if extraction.NavigationCluster(n) {
		t.Fatal("editorial reference list misclassified")
	}
}
