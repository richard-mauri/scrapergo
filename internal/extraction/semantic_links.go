package extraction

import (
 "net/url"
 "clean-article-reader/internal/article"

 "clean-article-reader/internal/links"
 "clean-article-reader/internal/sourcehtml"
)

// preserveSelectedHTMLLinks extracts markup evidence from Readability-selected
// HTML and delegates block matching to the independent links package.
func PreserveSelectedHTMLLinks(blocks []article.Block, cleaned string, base *url.URL, inlineParts func(*sourcehtml.Node,string) []article.TextPart) {
 root:=sourcehtml.Parse(cleaned)
 evidence:=[]links.SelectedParts{}
 for _,node:=range sourcehtml.Find(root,func(n *sourcehtml.Node)bool {
  return n.Tag=="p"||n.Tag=="li"||n.Tag=="h2"||n.Tag=="h3"||n.Tag=="h4"||n.Tag=="blockquote"
 }) {
  if sourcehtml.InNoisy(node,nil)||len(sourcehtml.Find(node,func(n *sourcehtml.Node)bool{return n.Tag=="a"}))==0 {continue}
  if links.SemanticKey(sourcehtml.Text(node))=="" {continue}
  parts:=inlineParts(node,base.String())
  linked:=false
  for _,part:=range parts {if part.URL!="" {linked=true;break}}
  if linked {evidence=append(evidence,links.SelectedParts{Text:sourcehtml.Text(node),Parts:parts})}
 }
 links.AssignSelectedParts(blocks,evidence)
}
