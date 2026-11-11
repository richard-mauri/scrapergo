package main

import (
 "testing"

 "clean-article-reader/internal/article"
)

func TestCaptionShapeTripleWithCredit(t *testing.T) {
 const caption="The keyboard and trackpad remain excellent."
 const credit="Sample Photographer"
 blocks:=[]article.Block{{Kind:"p",Text:caption},{Kind:"p",Text:credit},{Kind:"p",Text:caption+" "+credit}}
 s:=classifyCaptionPair(blocks,1,3)
 if s.Gap!=1 || s.MiddleKind!="p" || !s.CaptionPlusMiddleExact || !s.CaptionPlusMiddleCompact {
  t.Fatalf("unrecognized caption/credit triple: %+v",s)
 }
}

func TestCaptionShapeSpacingOnly(t *testing.T) {
 blocks:=[]article.Block{{Kind:"p",Text:"A photograph shows the compact display.Sample Photographer"},{Kind:"p",Text:"A photograph shows the compact display. Sample Photographer"}}
 s:=classifyCaptionPair(blocks,1,2)
 if s.Gap!=0 || s.PairExact || !s.PairCompact {t.Fatalf("spacing-only candidate misclassified: %+v",s)}
}

func TestCaptionShapeIsNotDeletionAuthorization(t *testing.T) {
 const editorial="The finding was reported twice intentionally."
 blocks:=[]article.Block{{Kind:"p",Text:editorial},{Kind:"h2",Text:"Unrelated heading"},{Kind:"p",Text:editorial}}
 s:=classifyCaptionPair(blocks,1,3)
 if s.CaptionPlusMiddleExact || s.CaptionPlusMiddleCompact || s.MiddleKind!="h2" {
  t.Fatalf("editorial repeat mistaken for triple: %+v",s)
 }
}
