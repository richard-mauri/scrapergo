package main

import (
 "bytes"
 "fmt"
 "net/url"
 "strings"

 "clean-article-reader/internal/article"
 "clean-article-reader/internal/sourcehtml"

 readability "codeberg.org/readeck/go-readability/v2"
)

// Source evidence is an opt-in offline diagnostic. It compares candidate
// passages against source and selected Readability HTML paragraphs/figcaptions.
// It never prints or persists raw article text or HTML. A source occurrence
// count is a clue, not proof that a duplicate is editorially unintended.
func sourceOccurrences(raw string, fragment string) (paragraphs, captions int) {
 root:=sourcehtml.Parse(raw)
 target:=normalize(fragment)
 if target=="" {return}
 for _,n:=range sourcehtml.Find(root,func(n *sourcehtml.Node)bool{
  return n.Tag=="p" || n.Tag=="figcaption"
 }) {
  // Child paragraphs in a caption should not count twice.
  if n.Tag=="p" {
   nested:=false
   for par:=n.Parent;par!=nil;par=par.Parent {
    if par.Tag=="figcaption" {nested=true;break}
   }
   if nested {continue}
  }
  value:=normalize(sourcehtml.Text(n))
  if value=="" {continue}
  // HTML serializers may add whitespace across inline elements; this is a
  // conservative text-presence test rather than precise DOM provenance.
  present:=value==target || strings.Contains(value,target)
  if !present {continue}
  if n.Tag=="figcaption" {captions++} else {paragraphs++}
 }
 return
}

func reportSourceEvidence(raw string, sourceURL string, blocks []article.Block, pairs []match) error {
 base,err:=url.Parse(sourceURL)
 if err!=nil || !base.IsAbs() {return fmt.Errorf("invalid source URL")}
 parsed,err:=readability.FromReader(strings.NewReader(raw),base)
 if err!=nil {return fmt.Errorf("Readability parse: %w",err)}
 var selected bytes.Buffer
 if err:=parsed.RenderHTML(&selected);err!=nil {return fmt.Errorf("Readability RenderHTML: %w",err)}
 fmt.Println("Source evidence (counts of source/selected HTML elements containing the matched shorter text; zero does not prove absence)")
 fmt.Println("Pair         Source p/figcaption   Readability p/figcaption")
 for _,pair:=range pairs {
  if pair.First<1 || pair.Second<1 || pair.First>len(blocks) || pair.Second>len(blocks) {continue}
  a,b:=normalize(blocks[pair.First-1].Text),normalize(blocks[pair.Second-1].Text)
  target:=a
  if len([]rune(b))<len([]rune(a)) {target=b}
  sourceP,sourceCap:=sourceOccurrences(raw,target)
  selectedP,selectedCap:=sourceOccurrences(selected.String(),target)
  fmt.Printf("%3d/%-3d      %3d/%-3d                 %3d/%-3d\n",pair.First,pair.Second,sourceP,sourceCap,selectedP,selectedCap)
 }
 fmt.Println("Presence counts alone cannot distinguish legitimate repetition, hidden markup or Readability flattening. No article text or URL is printed.")
 return nil
}
