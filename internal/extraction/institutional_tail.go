package extraction

import (
 "clean-article-reader/internal/article"
 "clean-article-reader/internal/links"
 "clean-article-reader/internal/sourcehtml"
 "clean-article-reader/internal/structured"
 "net/url"
 "strings"
)

// RecoverInstitutionalTail restores a short source-backed institutional
// explanation immediately following a selected article. It never downloads
// content and does not cross into site footers or navigation.
func RecoverInstitutionalTail(raw string, selected []article.Block, base *url.URL) []article.Block {
 if base == nil || len(selected) == 0 { return selected }
 anchor := ""
 for _, b := range selected {
  if b.Kind == "p" && len(sourcehtml.Normalize(b.Text)) >= 65 { anchor = sourcehtml.Normalize(b.Text) }
 }
 if anchor == "" { return selected }
 root := sourcehtml.Parse(raw)
 nodes := sourcehtml.Find(root,func(n *sourcehtml.Node) bool {
  return n.Tag=="h2" || n.Tag=="h3" || n.Tag=="p"
 })
 anchorIndex := -1
 for i,n := range nodes {
  if n.Tag != "p" {continue}
  v := sourcehtml.Normalize(sourcehtml.Text(n))
  if v == anchor || (len(v) >= 65 && strings.HasPrefix(v,anchor)) {
   if anchorIndex != -1 {return selected}
   anchorIndex = i
  }
 }
 if anchorIndex < 0 {return selected}
 // A nearby heading after the final retained prose is required. Do not
 // traverse another unrelated substantive section to find an About section.
 for i:=anchorIndex+1; i<len(nodes) && i<=anchorIndex+8; i++ {
  n:=nodes[i]
  if n.Tag!="h2" && n.Tag!="h3" {continue}
  heading:=sourcehtml.Normalize(sourcehtml.Text(n))
  if !strings.HasPrefix(strings.ToLower(heading),"about ") || len(heading)>110 {return selected}
  if excludedInstitutionalNode(n) {return selected}
  // The immediately adjacent publisher-controlled information panel may
  // be a sibling of main/article. Proximity and excluded ancestors guard it.
  tail:=[]article.Block{structured.TextBlock(n,base,sourcehtml.ResolveURL,links.IsDocumentURL)}
  substantive:=false
  for j:=i+1; j<len(nodes) && j<=i+4; j++ {
   p:=nodes[j]
   if p.Tag!="p" || excludedInstitutionalNode(p) {break}
   v:=sourcehtml.Normalize(sourcehtml.Text(p))
   if len(v)<18 || len(v)>2500 {break}
   if len(strings.Fields(v))>=35 {substantive=true}
   tail=append(tail,structured.TextBlock(p,base,sourcehtml.ResolveURL,links.IsDocumentURL))
  }
  if !substantive || len(tail)<2 {return selected}
  existing:=map[string]bool{}
  for _,b:=range selected {existing[sourcehtml.Normalize(b.Text)]=true}
  out:=append([]article.Block{}, selected...)
  for _,b:=range tail {if !existing[sourcehtml.Normalize(b.Text)] {out=append(out,b)}}
  return out
 }
 return selected
}

func excludedInstitutionalNode(n *sourcehtml.Node) bool {
 for p:=n;p!=nil;p=p.Parent {
  switch p.Tag {
  case "footer","nav","aside","header","form":return true
  }
 }
 return false
}

func shareEditorialRegion(a,b *sourcehtml.Node) bool {
 for p:=a;p!=nil;p=p.Parent {
  if p.Tag!="main" && p.Tag!="article" {continue}
  for q:=b;q!=nil;q=q.Parent {if q==p{return true}}
 }
 return false
}
