package main

import (
 "clean-article-reader/internal/extraction"
 "net/url"
 "strings"
 "testing"
)

// The source has an opening editorial illustration preceded by short
// introductory text that Readability may omit. That omission must not
// send the first figure to the end of the selected article.
func TestFigureAnchorsSurviveUnmatchedOpeningProse(t *testing.T) {
 base, _ := url.Parse("https://uberblok.storagebits.workers.dev/en/")
 opening := "This museum explains how ZFS protects data through crashes and incomplete writes."
 middle := "A pool is millions of blocks spread over several disks, and the uberblock names its consistent root."
 ending := "The engineers who designed ZFS published details of their work and shared the source."
 html := `<article><h1>Uberblok</h1><p>Storage has always had the same enemies.</p><p>Every animation maps to source code.</p><figure><img src="/media/uberblock-glass.jpg" alt="Magnifying glass over source"/><figcaption>The Hall's object under the glass.</figcaption></figure><h2>The object in the Hall</h2><p>`+middle+`</p><p>`+ending+`</p><figure><img src="/media/hard-disk.jpg" alt="Exploded hard disk"/></figure></article>`
 blocks := []Block{{Kind:"p",Text:opening},{Kind:"h2",Text:"The object in the Hall"},{Kind:"p",Text:middle},{Kind:"p",Text:ending}}
 got := extraction.PositionArticleFigures(blocks, html, base)
 if len(got)!=6 {t.Fatalf("expected four text blocks and two figures, got %d: %#v",len(got),got)}
 first, second := -1,-1
 for i,b:=range got {
  if strings.HasSuffix(b.URL,"/uberblock-glass.jpg") {first=i}
  if strings.HasSuffix(b.URL,"/hard-disk.jpg") {second=i}
 }
 if first!=0 {t.Errorf("opening image at %d, want first block; blocks=%#v",first,got)}
 if second<=first || second<3 {t.Errorf("later figure misplaced at %d (first=%d)",second,first)}
}

func TestFigureAnchorsDoNotExhaustAfterUnmatchedIntermediateParagraph(t *testing.T) {
 base,_:=url.Parse("https://example.org/article")
 html:=`<article><p>Introductory paragraph retained in the selection.</p><p>Interactive explanation omitted from the text selection.</p><figure><img src="/diagrams/figure.png" alt="Figure"/></figure><p>Later paragraph retained in the selection.</p></article>`
 blocks:=[]Block{{Kind:"p",Text:"Introductory paragraph retained in the selection."},{Kind:"p",Text:"Later paragraph retained in the selection."}}
 got:=extraction.PositionArticleFigures(blocks,html,base)
 if len(got)!=3 || got[1].Kind!="img" {t.Fatalf("figure should remain between retained prose, got %#v",got)}
}
