package main

import (
 "bytes"
 "strings"
 "testing"

 "clean-article-reader/internal/article"
)

func TestSourceOccurrencesDistinguishEditorialParagraphAndCaption(t *testing.T) {
 const passage="Field observers recorded a repeating sequence of measurements and independently verified the result."
 html:=`<article><p>`+passage+`</p><figure><img src="/a.jpg"><figcaption>`+passage+`</figcaption></figure><p>Completely unrelated reporting.</p></article>`
 p,c:=sourceOccurrences(html,passage)
 if p!=1 || c!=1 {t.Fatalf("expected one editorial paragraph and one caption, got %d/%d",p,c)}
}

func TestSourceOccurrencesOnlyCountMatchingSourceElements(t *testing.T) {
 const passage="Technical teams described independent experiments with comparable equipment and careful observational results."
 html:=`<article><p>`+passage+`</p><p>Unrelated second paragraph</p></article>`
 p,c:=sourceOccurrences(html,passage)
 if p!=1 || c!=0 {t.Fatalf("unexpected source count %d/%d",p,c)}
}

func TestSourceEvidenceDoesNotPrintProse(t *testing.T) {
 const passage="Researchers catalogued an extensive record of equipment and observations that were later discussed in detail."
 html:=`<html><head><title>Field report</title></head><body><article><h1>Field report</h1><p>`+passage+`</p><p>`+passage+`</p></article></body></html>`
 // The production flag runs this against existing captures, while this
 // regression ensures matching text is never emitted as part of a report.
 blocks:=[]article.Block{{Kind:"p",Text:passage},{Kind:"p",Text:passage}}
 pairs:=matches(blocks)
 if len(pairs)!=1 {t.Fatalf("expected exact match, got %+v",pairs)}
 var _ bytes.Buffer
 p,c:=sourceOccurrences(html,passage)
 if p!=2 || c!=0 {t.Fatalf("unexpected repeated source count %d/%d",p,c)}
 if strings.Contains(pairs[0].Digest,passage) {t.Fatal("digest exposed source text")}
}
