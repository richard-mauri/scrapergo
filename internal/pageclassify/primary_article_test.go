package pageclassify

import (
 "fmt"
 "strings"
 "testing"
)

func TestArticleWithManyRelatedStoriesIsNotIndex(t *testing.T) {
 paragraph := strings.Repeat("This is substantive editorial analysis concerning the company's latest financial results and developments. ", 3)
 var b strings.Builder
 b.WriteString("<html><main><article><h1>Why is Truist Financial stock gaining today?</h1>")
 for i:=0;i<5;i++ {b.WriteString("<p>"+paragraph+"</p>")}
 b.WriteString("</article><section><h2>Most Popular Articles</h2>")
 for i:=0;i<32;i++ {b.WriteString(fmt.Sprintf("<a href='/news/article-%d'><h3>Market news and financial results today number %d</h3></a>",i,i))}
 b.WriteString("</section></main></html>")
 html:=b.String()
 if !HasPrimaryArticleEvidence(html) {t.Fatal("primary article evidence not recognized")}
 if IsIndex(html) {t.Fatal("related article links overrode primary article")}
 if IsLanding(html) {t.Fatal("related article links misclassified editorial article as landing page")}
}

func TestHeadlineTeasersAreNotPrimaryArticle(t *testing.T) {
 var b strings.Builder
 b.WriteString("<html><main>")
 for i:=0;i<28;i++ {b.WriteString(fmt.Sprintf("<article><h2>Market story number %d</h2><a href='/news/%d'>Read details about market movers</a><p>Brief teaser only.</p></article>",i,i))}
 b.WriteString("</main></html>")
 if HasPrimaryArticleEvidence(b.String()) {t.Fatal("teaser cards should not qualify")}
}

func TestMainArticleWithoutSemanticArticleWrapper(t *testing.T) {
 paragraph := strings.Repeat("Substantial editorial reporting about Truist Financial earnings, investor reaction and bank performance. ", 2)
 var b strings.Builder
 b.WriteString("<html><main><h1>Why is Truist Financial stock gaining today?</h1>")
 for i:=0;i<4;i++ {b.WriteString("<p>"+paragraph+"</p>")}
 b.WriteString("<section><h2>Most Popular Articles</h2>")
 for i:=0;i<35;i++ {b.WriteString(fmt.Sprintf("<a href='/news/%d'><h3>Another distinct market article and financial story %d</h3></a>",i,i))}
 b.WriteString("</section></main></html>")
 if IsIndex(b.String()) {t.Fatal("main editorial article incorrectly rejected as index")}
}

func TestRelatedLinkOnlyMainStillRejected(t *testing.T) {
 var b strings.Builder
 b.WriteString("<html><main><h1>Market news today</h1>")
 for i:=0;i<35;i++ {b.WriteString(fmt.Sprintf("<a href='/news/%d'><h3>Another distinct market article and financial story %d</h3></a>",i,i))}
 b.WriteString("</main></html>")
 if !IsIndex(b.String()) {t.Fatal("real link-only news index accepted as article")}
}
