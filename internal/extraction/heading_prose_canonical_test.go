package extraction

import (
 "net/url"
 "strings"
 "testing"

 "clean-article-reader/internal/article"
)

// HTML descendant text inserts spacing between inline element nodes, while
// TextBlock preserves actual punctuation. Heading recovery must check the
// canonical block text against already retained content, not just the HTML
// descendant serialization, or the same article paragraph gets reinserted.
func TestHeadingAnchoredProseDoesNotReinsertStyledCanonicalParagraph(t *testing.T) {
 const introduction="Research teams described their field observations, collected throughout the season with independent equipment, and explained their implications for future surveys."
 const middle="The investigators presented numerical results, supported by separate validation experiments, in a complete technical account for readers."
 raw:=`<main><article>`+
  `<p>Research teams described their field <em>observations</em>, collected throughout the season with independent equipment, and explained their implications for future surveys.</p>`+
  `<h2>Introduction</h2>`+
  `<p>The investigators presented numerical <strong>results</strong>, supported by separate validation experiments, in a complete technical account for readers.</p>`+
  `<h2>Method</h2><h2>Findings</h2></article></main>`
 u,_:=url.Parse("https://example.org/research")
 selected:=[]article.Block{
  {Kind:"p",Text:introduction},
  {Kind:"h2",Text:"Introduction"},
  {Kind:"p",Text:middle},
  {Kind:"h2",Text:"Method"},
  {Kind:"h2",Text:"Findings"},
 }
 got:=RestoreHeadingAnchoredProse(raw,selected,u)
 if len(got)!=len(selected) {t.Fatalf("reinserted already selected styled paragraph: %#v",got)}
 for _,want:=range []string{introduction,middle} {
  count:=0
  for _,b:=range got {if b.Kind=="p" && b.Text==want {count++}}
  if count!=1 {t.Fatalf("expected one retained paragraph, got %d for %q",count,want)}
 }
 if !strings.Contains(raw,"<em>") || !strings.Contains(raw,"<strong>") {t.Fatal("fixture must contain actual inline HTML")}
}
