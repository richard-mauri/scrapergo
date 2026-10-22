// Package links restores missing article links from already-selected body prose.
// It never imports additional paragraphs or fetches network content.
package links

import (
 "regexp"
 "strings"

 "clean-article-reader/internal/article"
)

var matchWords = regexp.MustCompile(`[\pL\pN]+`)

// TextKey treats insignificant punctuation and spacing consistently with
// the historic article link recovery logic.
func TextKey(s string) string {
 return strings.Join(matchWords.FindAllString(strings.ToLower(s),-1)," ")
}

type Anchor struct {
 Text string
 URL string
}

type ParagraphEvidence struct {
 Text string
 Anchors []Anchor
}

// Restore applies source anchors only to a fully matching, selected article
// block. Existing Parts containing links or existing Links take precedence.
func Restore(a *article.Article, evidence []ParagraphEvidence) {
 if a==nil {return}
 for _, item := range evidence {
  if len(item.Anchors)==0 {continue}
  key:=TextKey(item.Text)
  if len(key)<12 {continue}
  for i:=range a.Blocks {
   block:=&a.Blocks[i]
   if block.Kind=="img" || block.Kind=="topic-link" || TextKey(block.Text)!=key {continue}
   hasURL:=false
   for _,part:=range block.Parts {
    if part.URL!="" {hasURL=true;break}
   }
   if hasURL || len(block.Links)!=0 {break}
   seen:=map[string]bool{}
   for _,anchor:=range item.Anchors {
    label:=anchor.Text
    href:=anchor.URL
    if label=="" || href=="" || seen[label+"|"+href] {continue}
    if !strings.Contains(block.Text,label) {continue}
    seen[label+"|"+href]=true
    block.Links=append(block.Links,article.InlineLink{Text:label,URL:href})
   }
   if len(block.Links)>0 {block.Parts=nil}
   break
  }
 }
}
