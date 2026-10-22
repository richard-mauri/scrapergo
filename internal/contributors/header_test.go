package contributors

import (
 "testing"
 "clean-article-reader/internal/article"
)

func TestSeparateHeaderConsolidatesLinkedContributors(t *testing.T) {
 a:=article.Article{Blocks:[]article.Block{
  {Kind:"p",Text:"By"},
  {Kind:"p",Text:"First Reporter",Parts:[]article.TextPart{{Text:"First Reporter",URL:"https://example.org/first"}}},
  {Kind:"p",Text:","},
  {Kind:"p",Text:"Second Reporter",Links:[]article.InlineLink{{Text:"Second Reporter",URL:"https://example.org/second"}}},
  {Kind:"img",Role:"author",URL:"https://example.org/portrait.jpg"},
  {Kind:"img",Role:"editorial",URL:"https://example.org/lead.jpg"},
  {Kind:"p",Text:"Article body"},
 }}
 SeparateHeader(&a)
 if a.Byline!="First Reporter, Second Reporter" {t.Fatalf("byline %q",a.Byline)}
 if len(a.Blocks)!=2 || a.Blocks[0].URL!="https://example.org/lead.jpg" || a.Blocks[1].Text!="Article body" {t.Fatalf("blocks %+v",a.Blocks)}
}
func TestSeparateHeaderRequiresEditorialBoundary(t *testing.T) {
 a:=article.Article{Blocks:[]article.Block{{Kind:"p",Text:"By"},{Kind:"p",Text:"First Reporter",Links:[]article.InlineLink{{Text:"First Reporter",URL:"https://example.org/a"}}}}}
 SeparateHeader(&a)
 if len(a.Blocks)!=2 || a.Byline!="" {t.Fatalf("unexpected separation %+v",a)}
}
