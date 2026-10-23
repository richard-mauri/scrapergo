package main

import (
 articlemodel "clean-article-reader/internal/article"
 "clean-article-reader/internal/contributors"
 "clean-article-reader/internal/ui"
 "strings"
 "testing"
)

func TestContributorBylineReconciliation(t *testing.T) {
 linked:=func(name string) articlemodel.Block{return articlemodel.Block{Kind:"p",Text:name,Links:[]articlemodel.InlineLink{{Text:name,URL:"https://www.npr.org/people/123"}}}}
 a:=articlemodel.Article{Byline:"Ruth Sherlock",Blocks:[]articlemodel.Block{
  {Kind:"p",Text:"Middle East"}, {Kind:"p",Text:"By"},
  linked("Ruth Sherlock"),{Kind:"p",Text:","},
  {Kind:"img",Role:"author",Text:"Photo of Samantha Balaban",URL:"https://example.org/samantha.jpg"},
  linked("Samantha Balaban"),{Kind:"p",Text:","},
  linked("Nuha Musleh"),{Kind:"p",Text:","},
  linked("Daniel Estrin"),
  {Kind:"img",Role:"lead",URL:"https://example.org/security.jpg"},
  {Kind:"p",Text:"Real reporting remains."},
 }}
 contributors.ReconcileContributorByline(&a)
 expected:="Ruth Sherlock, Samantha Balaban, Nuha Musleh, Daniel Estrin"
 if a.Byline!=expected {t.Fatalf("byline=%q, want %q",a.Byline,expected)}
 for _,b:=range a.Blocks {
  if strings.TrimSpace(b.Text)=="By"||strings.TrimSpace(b.Text)=="," {t.Fatalf("orphan byline fragment %+v",b)}
 }
 if len(a.Blocks)!=4 || a.Blocks[0].Text!="Middle East" || a.Blocks[1].Role!="author" || a.Blocks[2].Role!="lead" || a.Blocks[3].Text!="Real reporting remains." {t.Fatalf("lost editorial blocks: %+v",a.Blocks)}
}
func TestNoBylineFragmentsPreservesEditorialLinks(t *testing.T) {
 a:=articlemodel.Article{Byline:"Writer",Blocks:[]articlemodel.Block{{Kind:"p",Text:"A story about someone."},{Kind:"p",Text:"Further reading",Links:[]articlemodel.InlineLink{{Text:"Further reading",URL:"https://example.org/story"}}}}}
 contributors.ReconcileContributorByline(&a)
 if a.Byline!="Writer"||len(a.Blocks)!=2 {t.Fatalf("ordinary content altered: %+v",a)}
}
func TestOrdinaryArticleParagraphsCollapseWhitespace(t *testing.T){
 page:=ui.Render(appVersion,appBuild)
 if !strings.Contains(page,".content p{margin:1.1em 0;white-space:normal}") {t.Fatal("ordinary paragraphs should collapse source whitespace")}
}
