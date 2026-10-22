package main

import (
 "clean-article-reader/internal/sourcehtml"
 "net/url"

 "clean-article-reader/internal/links"
)

// recoverMissingArticleLinks translates parser-specific source evidence before
// handing policy decisions to the standalone links package.
func recoverMissingArticleLinks(a *Article, raw, source string) {
 if a==nil {return}
 base,err:=url.Parse(source)
 if err!=nil {return}
 root:=parseHTML(raw)
 evidence:=[]links.ParagraphEvidence{}
 for _,node:=range find(root,func(n *Node)bool{return n.Tag=="p"||n.Tag=="li"}) {
  if sourcehtml.InNoisy(node,nil) {continue}
  anchors:=find(node,func(n *Node)bool{return n.Tag=="a"})
  if len(anchors)==0 {continue}
  record:=links.ParagraphEvidence{Text:textOf(node)}
  for _,anchor:=range anchors {
   record.Anchors=append(record.Anchors,links.Anchor{
    Text:normalize(textOf(anchor)),
    URL:sourcehtml.ResolveURL(anchor.Attr["href"],base),
   })
  }
  evidence=append(evidence,record)
 }
 links.Restore(a,evidence)
}
