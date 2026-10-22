package main

import (
 "net/url"

 "clean-article-reader/internal/links"
 "clean-article-reader/internal/sourcehtml"
)

// preserveSelectedHTMLLinks extracts markup evidence from Readability-selected
// HTML and delegates block matching to the independent links package.
func preserveSelectedHTMLLinks(blocks []Block, cleaned string, base *url.URL) {
 root:=parseHTML(cleaned)
 evidence:=[]links.SelectedParts{}
 for _,node:=range find(root,func(n *Node)bool {
  return n.Tag=="p"||n.Tag=="li"||n.Tag=="h2"||n.Tag=="h3"||n.Tag=="h4"||n.Tag=="blockquote"
 }) {
  if sourcehtml.InNoisy(node,nil)||len(find(node,func(n *Node)bool{return n.Tag=="a"}))==0 {continue}
  if links.SemanticKey(textOf(node))=="" {continue}
  parts:=inlineTextParts(node,base.String())
  linked:=false
  for _,part:=range parts {if part.URL!="" {linked=true;break}}
  if linked {evidence=append(evidence,links.SelectedParts{Text:textOf(node),Parts:parts})}
 }
 links.AssignSelectedParts(blocks,evidence)
}
