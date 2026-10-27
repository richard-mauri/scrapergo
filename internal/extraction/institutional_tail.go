package extraction

import (
 "clean-article-reader/internal/article"
 "clean-article-reader/internal/sourcehtml"
 "clean-article-reader/internal/structured"
 "clean-article-reader/internal/links"
 "net/url"
 "strings"
)

// RecoverInstitutionalTail retains a bounded, source-backed "About ..." section
// near the end of an article. It does not turn page footers or link panels into
// article prose, and it performs no network access.
func RecoverInstitutionalTail(raw string, selected []article.Block, base *url.URL) []article.Block {
 if base == nil || len(selected)==0 { return selected }
 anchor := ""
 for _, b := range selected {
  if b.Kind == "p" && len(sourcehtml.Normalize(b.Text)) >= 65 { anchor = sourcehtml.Normalize(b.Text) }
 }
 if anchor == "" { return selected }
 root := sourcehtml.Parse(raw)
 nodes := sourcehtml.Find(root,func(n *sourcehtml.Node) bool {
  return n.Tag=="h2" || n.Tag=="h3" || n.Tag=="p"
 })
 anchorIndex:=-1
 for i,n:=range nodes {
  if n.Tag!="p" {continue}
  v:=sourcehtml.Normalize(sourcehtml.Text(n))
  if v==anchor || (len(v)>65 && strings.HasPrefix(v,anchor)) {
   if anchorIndex!=-1 { return selected }
   anchorIndex=i
  }
 }
 if anchorIndex<0 {return selected}
 // Only accept the first nearby institutional heading, before an unrelated
 // heading or another substantial prose section.
 for i:=anchorIndex+1; i<len(nodes) && i<=anchorIndex+8; i++ {
  n:=nodes[i]
  if n.Tag!="h2" && n.Tag!="h3" {continue}
  heading:=sourcehtml.Normalize(sourcehtml.Text(n))
  if !strings.HasPrefix(strings.ToLower(heading),"about ") || len(heading)>110 {return selected}
  if i+1>=len(nodes) || nodes[i+1].Tag!="p" {return selected}
  p:=nodes[i+1]
  prose:=sourcehtml.Normalize(sourcehtml.Text(p))
  if len(strings.Fields(prose))<45 || len(prose)>2500 {return selected}
  // Both selected article prose and tail must be within one editorial main
  // region; do not cross from article content into the site's global footer.
  if !shareEditorialRegion(nodes[anchorIndex],n) {return selected}
  for _, b:=range selected {if sourcehtml.Normalize(b.Text)==heading || sourcehtml.Normalize(b.Text)==prose{return selected}}
  title:=structured.TextBlock(n,base,sourcehtml.ResolveURL,links.IsDocumentURL)
  body:=structured.TextBlock(p,base,sourcehtml.ResolveURL,links.IsDocumentURL)
  return append(append(append([]article.Block{},selected...),title),body)
 }
 return selected
}

func shareEditorialRegion(a,b *sourcehtml.Node) bool {
 for p:=a;p!=nil;p=p.Parent {
  if p.Tag!="main" && p.Tag!="article" {continue}
  for q:=b;q!=nil;q=q.Parent {if q==p{return true}}
 }
 return false
}
