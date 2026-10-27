package extraction

import (
 "clean-article-reader/internal/article"
 "net/url"
 "strings"
 "testing"
)

func TestRestoreHeadingAnchoredProse(t *testing.T) {
 first:="This long editorial paragraph belongs to the first section and contains important technical details that should not be lost during article extraction."
 second:="A second section explains meaningful environmental improvements and manufacturing information that readers need to see in the correct document order."
 raw:="<main><article><h2>Performance</h2><p>"+first+"</p><h2>Environment</h2><p>"+second+"</p><h2>Pricing</h2><ul><li>Pricing details remain intact.</li></ul></article></main>"
 u,_:=url.Parse("https://example.org/release")
 in:=[]article.Block{{Kind:"h2",Text:"Performance"},{Kind:"h2",Text:"Environment"},{Kind:"h2",Text:"Pricing"},{Kind:"li",Text:"Pricing details remain intact."}}
 got:=RestoreHeadingAnchoredProse(raw,in,u)
 if len(got)!=6 || got[1].Text!=first || got[3].Text!=second || got[5].Kind!="li" {
  t.Fatalf("paragraph recovery/order incorrect: %#v",got)
 }
 again:=RestoreHeadingAnchoredProse(raw,got,u)
 if len(again)!=len(got) {t.Fatal("non-idempotent recovery")}
}

func TestRestoreHeadingAnchoredProseRequiresCorroboration(t *testing.T) {
 raw:="<main><h2>Performance</h2><p>A long promotional paragraph from a page widget about special products and unrelated services.</p><h2>Environment</h2><h2>Pricing</h2></main>"
 u,_:=url.Parse("https://example.org")
 got:=RestoreHeadingAnchoredProse(raw,[]article.Block{{Kind:"h2",Text:"Performance"},{Kind:"h2",Text:"Pricing"}},u)
 if len(got)!=2 {t.Fatalf("weak anchors must not recover content: %#v",got)}
}

func TestRestoreHeadingAnchoredProseSuppressesCopyWidget(t *testing.T) {
 raw:="<main><article><h2>Performance</h2><h2>Environment</h2><h2>Pricing</h2></article></main>"
 u,_:=url.Parse("https://example.org")
 giant:="Text of this article"+strings.Repeat(" repeated full source paragraph",400)
 in:=[]article.Block{{Kind:"h2",Text:"Performance"},{Kind:"h2",Text:"Environment"},{Kind:"h2",Text:"Pricing"},{Kind:"li",Text:giant},{Kind:"li",Text:"Authentic feature list item"}}
 got:=RestoreHeadingAnchoredProse(raw,in,u)
 if len(got)!=4 || got[3].Text!="Authentic feature list item" {t.Fatalf("copy widget retained or real list removed: %#v",got)}
}
