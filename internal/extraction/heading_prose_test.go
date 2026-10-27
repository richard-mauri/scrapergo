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

func TestRestoreHeadingAnchoredProseRecoversIntroduction(t *testing.T) {
 intro1:="The company announced a new professional computer with improved performance and longer battery life, bringing substantial benefits to creative professionals across many workflows."
 intro2:="The new system introduces more capable processing, faster memory, and enhancements that make it appealing to developers and customers working with demanding creative applications."
 first:="The performance section includes important technical details and concrete improvements that should remain adjacent to the section heading."
 raw:="<div><p>Unrelated page introduction and promotional description outside the article container must never be included by recovery.</p><article><p>"+intro1+"</p><p>"+intro2+"</p><h2>Performance</h2><p>"+first+"</p><h2>Environment</h2><h2>Pricing</h2></article></div>"
 u,_:=url.Parse("https://example.org/newsroom/release")
 blocks:=[]article.Block{{Kind:"h2",Text:"Performance"},{Kind:"h2",Text:"Environment"},{Kind:"h2",Text:"Pricing"}}
 got:=RestoreHeadingAnchoredProse(raw,blocks,u)
 if len(got)!=6 || got[0].Text!=intro1 || got[1].Text!=intro2 || got[2].Text!="Performance" || got[3].Text!=first {
  t.Fatalf("missing or misplaced introduction: %#v",got)
 }
 repeat:=RestoreHeadingAnchoredProse(raw,got,u)
 if len(repeat)!=len(got) {t.Fatalf("intro recovery duplicated text: %#v",repeat)}
}

func TestRestoreHeadingAnchoredProseDoesNotImportPageWideIntro(t *testing.T) {
 outside:="The promotional landing page includes a lengthy introduction, unrelated to this article, that must not be recovered into the article reader."
 raw:="<main><p>"+outside+"</p><h2>Performance</h2><h2>Environment</h2><h2>Pricing</h2></main>"
 u,_:=url.Parse("https://example.org/news")
 in:=[]article.Block{{Kind:"h2",Text:"Performance"},{Kind:"h2",Text:"Environment"},{Kind:"h2",Text:"Pricing"}}
 got:=RestoreHeadingAnchoredProse(raw,in,u)
 if len(got)!=len(in) {t.Fatalf("page-wide introduction was imported: %#v",got)}
}

func TestRestoreAppleStyleDivCopy(t *testing.T) {
 intro:="The company announced a new professional computer with major performance improvements, additional capabilities and a significant upgrade for customers working in demanding environments."
 middle:="The new processor offers faster performance for creative workflows, local artificial intelligence models and productivity applications used by professionals worldwide."
 end:="The updated operating system introduces additional tools, reliability improvements and new features to make computing more efficient across several applications."
 raw:="<main><article class='article'><div class='pagebody text component'><div class='component-content'><div class='pagebody-copy'>"+intro+"</div></div></div>"+
  "<div class='pagebody text component'><div class='component-content'><h2>Performance</h2><div class='pagebody-copy'>"+middle+"</div></div></div>"+
  "<div class='pagebody text component'><div class='component-content'><h2>Environment</h2><div class='pagebody-copy'>"+end+"</div></div></div>"+
  "<div class='pagebody text component'><div class='component-content'><h2>Pricing</h2></div></div>"+
  "<div class='docsanddownloads'><ul><li>Text of this article "+strings.Repeat("duplicate contents ",600)+"</li></ul></div></article></main>"
 u,_:=url.Parse("https://example.org/newsroom/article")
 selected:=[]article.Block{{Kind:"h2",Text:"Performance"},{Kind:"h2",Text:"Environment"},{Kind:"h2",Text:"Pricing"},{Kind:"li",Text:"Text of this article "+strings.Repeat("duplicate contents ",600)}}
 got:=RestoreHeadingAnchoredProse(raw,selected,u)
 if len(got)!=6 || got[0].Text!=intro || got[0].Kind!="p" || got[2].Text!=middle || got[4].Text!=end {t.Fatalf("missing real article div copy: %#v",got)}
}
