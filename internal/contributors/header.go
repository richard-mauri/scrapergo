// Package contributors owns contributor-name recognition and byline header separation.
package contributors

import (
 "strings"
 "clean-article-reader/internal/article"
)

// LinkedName accepts name anchors from either structured inline Parts or Links.
func LinkedName(b article.Block) bool {
 if b.Kind != "p" { return false }
 name := strings.TrimSpace(b.Text)
 if name == "" || len(name) > 90 { return false }
 for _,p := range b.Parts {
  if p.URL != "" && strings.EqualFold(strings.TrimSpace(p.Text),name) { return true }
 }
 for _,l := range b.Links {
  if l.URL != "" && strings.EqualFold(strings.TrimSpace(l.Text),name) { return true }
 }
 return false
}

// SeparateHeader removes only a contiguous multi-contributor preamble before
// the first confirmed editorial figure. Other story blocks are retained.
func SeparateHeader(a *article.Article) {
 if a == nil { return }
 boundary := -1
 for i,b := range a.Blocks {
  if b.Kind == "img" && (b.Role == "editorial" || b.Role == "lead") { boundary=i;break }
 }
 if boundary < 0 { return }
 start := -1
 for i:=0;i<boundary;i++ {
  if a.Blocks[i].Kind == "p" && strings.EqualFold(strings.TrimSpace(a.Blocks[i].Text),"By") { start=i;break }
 }
 if start < 0 { return }
 names:=[]string{}
 removed:=map[int]bool{}
 seen:=map[string]bool{}
 for i:=start;i<boundary;i++ {
  b:=a.Blocks[i]
  t:=strings.TrimSpace(b.Text)
  if i==start || (b.Kind=="p" && (t==","||t=="")) || (b.Kind=="img"&&b.Role=="author") {
   removed[i]=true;continue
  }
  if LinkedName(b) {
   if !seen[strings.ToLower(t)] { names=append(names,t);seen[strings.ToLower(t)]=true }
   removed[i]=true;continue
  }
  break
 }
 if len(names)<2 { return }
 out:=make([]article.Block,0,len(a.Blocks))
 for i,b:=range a.Blocks {if !removed[i] {out=append(out,b)}}
 a.Blocks=out
 a.Byline=strings.Join(names,", ")
}
