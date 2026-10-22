// Package documentmedia reconciles editorial image priority and delivered asset identities.
package documentmedia

import (
 "strings"
 "clean-article-reader/internal/article"
)

// NormalizeDocument preserves legacy title deduplication and editorial lead
// selection; callers provide their existing source image identity policy.
func NormalizeDocument(a *article.Article, identity func(string) string, sourceLead ...string) {
 if a==nil{return}
 title:=strings.Join(strings.Fields(strings.ToLower(a.Title))," ")
 for i,b:=range a.Blocks {
  if b.Kind=="img"{continue}
  if (b.Kind=="h1"||b.Kind=="h2"||b.Kind=="h3"||b.Kind=="h4"||b.Kind=="p")&&strings.Join(strings.Fields(strings.ToLower(b.Text))," ")==title {
   a.Blocks=append(a.Blocks[:i],a.Blocks[i+1:]...)
  }
  break
 }
 lead:=""
 if len(sourceLead)>0 {lead=strings.TrimSpace(sourceLead[0])}
 if lead=="" {lead=a.ImageURL}
 // Deduplicate delivered variants by original asset identity, retaining
 // editorial order and captions rather than imposing metadata ordering.
 out:=make([]article.Block,0,len(a.Blocks)+1)
 seen:=make(map[string]int)
 for _,b:=range a.Blocks {
  if b.Kind!="img" {out=append(out,b);continue}
  key:=identity(b.URL)
  if prev,ok:=seen[key];ok{
   existing:=&out[prev]
   if existing.Text=="" {existing.Text=b.Text}
   if existing.Role=="" {existing.Role=b.Role}
   continue
  }
  seen[key]=len(out)
  out=append(out,b)
 }
 // The document's first editorial figure is authoritative. Portraits,
 // icons and byline art are not lead candidates; existing lead roles
 // supplied by recovery identify an editorial figure when present.
 candidate:=-1
 for i,b:=range out {
  if b.Kind=="img" && b.Role=="lead" {candidate=i;break}
 }
 if candidate<0 {
  for i,b:=range out {if b.Kind=="img" && b.Role=="editorial" {candidate=i;break}}
 }
 if candidate<0 {
  for i,b:=range out {
   if b.Kind=="img" && b.Role!="author" && b.Role!="decorative" {candidate=i;break}
  }
 }
 for i:=range out {if out[i].Kind=="img" && out[i].Role=="lead" {out[i].Role=""}}
 if candidate>=0 {
  out[candidate].Role="lead"
  // A metadata rendition of this SAME original can improve image quality
  // without changing its editorial identity, position or caption.
  if lead!="" && identity(lead)==identity(out[candidate].URL) {
   out[candidate].URL=lead
  }
  a.ImageURL=out[candidate].URL
 } else if lead!="" {
  out=append([]article.Block{{Kind:"img",URL:lead,Role:"lead"}},out...)
  a.ImageURL=lead
 }
 a.Blocks=out
}

