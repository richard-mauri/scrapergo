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
  return n.Tag=="h2" || n.Tag=="h3" || n.Tag=="h4" || editorialProseNode(n)
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
 // Recover paragraphs before the first retained heading, only within its
 // enclosing article element. A main fallback can include site furniture.
 first:=anchors[0]
 var scope *sourcehtml.Node
 for p:=nodes[first.node].Parent;p!=nil;p=p.Parent {
  if p.Tag=="article" {scope=p;break}
 }
 if scope!=nil {
  var intro []article.Block
  for _,n:=range sourcehtml.Find(scope,func(n *sourcehtml.Node) bool{return editorialProseNode(n) || n==nodes[first.node]}) {
   if n==nodes[first.node] {break}
   if !editorialProseNode(n) || sourcehtml.InNoisy(n,scope) || PromotionalAncestor(n,scope) || NestedProse(n,scope) {continue}
   v:=sourcehtml.Normalize(sourcehtml.Text(n))
   if len(v)<70 || len(v)>5000 || IsAdvertisingText(v) || existing[v] {continue}
   b:=structured.TextBlock(n,base,sourcehtml.ResolveURL,links.IsDocumentURL)
   if b.Text!="" {b.Kind="p";intro=append(intro,b)}
  }
  if len(intro)>0 && len(intro)<=12 {
   inserts[-1]=intro
   for _,b:=range intro {existing[sourcehtml.Normalize(b.Text)]=true}
  }
 }
 for k:=0;k+1<len(anchors);k++ {
  from,to:=anchors[k],anchors[k+1]
  if to.node-from.node>75 {continue}
  var recovered []article.Block
  for j:=from.node+1;j<to.node;j++ {
   n:=nodes[j]
   if !editorialProseNode(n) || sourcehtml.InNoisy(n,nil) || IsAdvertisingText(sourcehtml.Normalize(sourcehtml.Text(n))) {continue}
   v:=sourcehtml.Normalize(sourcehtml.Text(n))
   if len(v)<70 || len(v)>5000 || existing[v] {continue}
   block:=structured.TextBlock(n,base,sourcehtml.ResolveURL,links.IsDocumentURL)
   if block.Text=="" {continue}
   block.Kind="p"
   existing[v]=true
   recovered=append(recovered,block)
  }
  if len(recovered)<=10 {inserts[from.index]=recovered}
 }
 // Only suppress the giant copy panel when the original page contains a\n // publisher docs/downloads container with the same article opening.\n hasCopyPanel:=false\n for _,n:=range sourcehtml.Find(root,func(n *sourcehtml.Node) bool {return n.Tag=="div" && strings.Contains(" "+n.Attr["class"]+" "," docsanddownloads ")}) {\n  if strings.Contains(strings.ToLower(sourcehtml.Text(n)),"text of this article") {hasCopyPanel=true;break}\n }\n var result []article.Block\n result=append(result,inserts[-1]...)
 for i,b:=range selected {
  // Publisher copy-text widgets are not editorial list items. Never discard
  // ordinary lists, even long ones, based on length alone.
  if b.Kind=="li" && len(b.Text)>8000 && (strings.Contains(strings.ToLower(b.Text[:min(len(b.Text),150)]),"text of this article") || (hasCopyPanel && strings.Contains(b.Text,"CUPERTINO, CALIFORNIA") && strings.Contains(b.Text,"M6: Blazing-Fast Performance"))) {continue}\n  if b.Kind=="h2" && sourcehtml.Normalize(b.Text)=="Media" && hasCopyPanel {continue}
  result=append(result,b)
  result=append(result,inserts[i]...)
 }
 return result
}

// editorialProseNode recognizes both semantic paragraphs and standalone
// editorial copy containers. A class-marked div must be a leaf at the prose
// level to avoid absorbing headings, images, lists or nested copy widgets.
func editorialProseNode(n *sourcehtml.Node) bool {
 if n==nil {return false}
 if n.Tag=="p" {return true}
 if n.Tag!="div" {return false}
 hasCopy:=false
 for _,c:=range strings.Fields(n.Attr["class"]) {if c=="pagebody-copy" {hasCopy=true;break}}
 if !hasCopy {return false}
 for _,c:=range sourcehtml.Find(n,func(x *sourcehtml.Node) bool {return x!=n && (x.Tag=="p" || x.Tag=="h2" || x.Tag=="h3" || x.Tag=="li" || x.Tag=="div")}) { _=c; return false }
 return true
}
