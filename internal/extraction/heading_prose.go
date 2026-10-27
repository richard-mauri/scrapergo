package extraction

import (
 "clean-article-reader/internal/article"
 "clean-article-reader/internal/links"
 "clean-article-reader/internal/sourcehtml"
 "clean-article-reader/internal/structured"
 "net/url"
 "strings"
)

// RestoreHeadingAnchoredProse recovers HTML paragraphs between consecutive
// headings that Readability retained. Both headings must match the source,
// in order, so the recovery cannot wander into arbitrary page furniture.
func RestoreHeadingAnchoredProse(raw string, selected []article.Block, base *url.URL) []article.Block {
 if base==nil || len(selected)==0 {return selected}
 root:=sourcehtml.Parse(raw)
 nodes:=sourcehtml.Find(root,func(n *sourcehtml.Node) bool {
  return n.Tag=="h2" || n.Tag=="h3" || n.Tag=="h4" || n.Tag=="p"
 })
 type anchor struct{index int; node int}
 var anchors []anchor
 searchAt:=0
 for i,b:=range selected {
  if b.Kind!="h2" && b.Kind!="h3" && b.Kind!="h4" {continue}
  label:=sourcehtml.Normalize(b.Text)
  match:=-1
  for j:=searchAt;j<len(nodes);j++ {
   if nodes[j].Tag!=b.Kind {continue}
   if sourcehtml.Normalize(sourcehtml.Text(nodes[j]))==label {match=j;break}
  }
  if match>=0 {anchors=append(anchors,anchor{i,match});searchAt=match+1}
 }
 // Multiple independent section headings corroborate the selected article.
 if len(anchors)<3 {return selected}
 existing:=map[string]bool{}
 for _,b:=range selected {existing[sourcehtml.Normalize(b.Text)]=true}
 inserts:=map[int][]article.Block{}
 for k:=0;k+1<len(anchors);k++ {
  from,to:=anchors[k],anchors[k+1]
  if to.node-from.node>75 {continue}
  var recovered []article.Block
  for j:=from.node+1;j<to.node;j++ {
   n:=nodes[j]
   if n.Tag!="p" || sourcehtml.InNoisy(n,nil) || IsAdvertisingText(sourcehtml.Normalize(sourcehtml.Text(n))) {continue}
   v:=sourcehtml.Normalize(sourcehtml.Text(n))
   if len(v)<70 || len(v)>5000 || existing[v] {continue}
   block:=structured.TextBlock(n,base,sourcehtml.ResolveURL,links.IsDocumentURL)
   if block.Kind!="p" {continue}
   existing[v]=true
   recovered=append(recovered,block)
  }
  if len(recovered)<=10 {inserts[from.index]=recovered}
 }
 var result []article.Block
 for i,b:=range selected {
  // Publisher copy-text widgets are not editorial list items. Never discard
  // ordinary lists, even long ones, based on length alone.
  if b.Kind=="li" && len(b.Text)>8000 && strings.Contains(strings.ToLower(b.Text[:min(len(b.Text),150)]),"text of this article") {continue}
  result=append(result,b)
  result=append(result,inserts[i]...)
 }
 return result
}
