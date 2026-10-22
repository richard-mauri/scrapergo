package links

import (
 "testing"

 "clean-article-reader/internal/article"
)

func TestRestoreOnlyMatchingArticleText(t *testing.T) {
 a:=article.Article{Blocks:[]article.Block{
  {Kind:"p",Text:"The report refers to a public filing and describes its context.",Parts:[]article.TextPart{{Text:"The report refers to a public filing and describes its context."}}},
  {Kind:"p",Text:"Unrelated body text remains intact."},
 }}
 Restore(&a,[]ParagraphEvidence{
  {Text:"A recommendation sidebar",Anchors:[]Anchor{{Text:"sidebar",URL:"https://example.org/sidebar"}}},
  {Text:"The report refers to a public filing and describes its context.",Anchors:[]Anchor{{Text:"public filing",URL:"https://example.org/filing"}}},
 })
 if len(a.Blocks)!=2 || len(a.Blocks[0].Links)!=1 || a.Blocks[0].Links[0].URL!="https://example.org/filing" {t.Fatalf("lost or added links: %+v",a.Blocks)}
 if len(a.Blocks[0].Parts)!=0 {t.Fatal("plain parts should yield to restored links")}
 if len(a.Blocks[1].Links)!=0 {t.Fatal("unrelated text modified")}
}
func TestRestoreDoesNotReplaceExistingLinks(t *testing.T) {
 a:=article.Article{Blocks:[]article.Block{{Kind:"p",Text:"The report cites the official public filing.",Parts:[]article.TextPart{{Text:"official public filing",URL:"https://example.org/original"}}}}}
 Restore(&a,[]ParagraphEvidence{{Text:a.Blocks[0].Text,Anchors:[]Anchor{{Text:"public filing",URL:"https://example.org/new"}}}})
 if len(a.Blocks[0].Links)!=0 || a.Blocks[0].Parts[0].URL!="https://example.org/original" {t.Fatal("overwrote existing link")}
}
func TestRestoreRejectsUnmatchedAndRepeatedAnchors(t *testing.T) {
 a:=article.Article{Blocks:[]article.Block{{Kind:"p",Text:"Officials reviewed the report and its recommendations."}}}
 Restore(&a,[]ParagraphEvidence{{Text:a.Blocks[0].Text,Anchors:[]Anchor{{Text:"the report",URL:"https://example.org/report"},{Text:"the report",URL:"https://example.org/report"},{Text:"an advertisement",URL:"https://example.org/ad"}}}})
 if len(a.Blocks[0].Links)!=1 {t.Fatalf("unexpected restored links: %+v",a.Blocks[0].Links)}
}
