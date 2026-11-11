package main

import (
 "strings"
 "testing"

 "clean-article-reader/internal/article"
)
func TestMatchFindsExactAndOverlapWithoutPrintingContents(t *testing.T) {
 sentence:="Technical observations can be repeated when markup contains two representations of the same editorial paragraph."
 blocks:=[]article.Block{
  {Kind:"p",Text:sentence},
  {Kind:"h2",Text:"Unrelated"},
  {Kind:"p",Text:sentence},
  {Kind:"img",URL:"https://example.org/a.jpg",Text:sentence+" Credit: sample"},
  {Kind:"p",Text:"Short"},
 }
 found:=matches(blocks)
 if len(found)!=2 || found[0].First!=1 || found[0].Second!=3 || found[0].Type!="exact" ||
  found[1].First!=1 || found[1].Second!=4 || found[1].Type!="overlap" {
  t.Fatalf("unexpected pair attribution: %+v",found)
 }
 if strings.Contains(found[0].Digest,sentence) {t.Fatal("digest leaked source content")}
}
func TestMatchIgnoresShortAndNonTextBlocks(t *testing.T) {
 blocks:=[]article.Block{
  {Kind:"p",Text:"A short label"},
  {Kind:"p",Text:"A short label"},
  {Kind:"table",Text:strings.Repeat("numbers and tables ",10)},
  {Kind:"p",Text:strings.Repeat("numbers and tables ",10)},
 }
 if got:=matches(blocks);len(got)!=0 {t.Fatalf("false matches: %+v",got)}
}
