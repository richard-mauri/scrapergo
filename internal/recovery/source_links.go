package recovery

import (
 "clean-article-reader/internal/sourcehtml"
 "net/url"

 "clean-article-reader/internal/links"
 "clean-article-reader/internal/article"
)

// recoverMissingArticleLinks translates parser-specific source evidence before
// handing policy decisions to the standalone links package.
func RecoverMissingArticleLinks(a *article.Article, raw, source string) {
 if a==nil {return}
 base,err:=url.Parse(source)
 if err!=nil {return}
 root:=sourcehtml.Parse(raw)
 evidence:=[]links.ParagraphEvidence{}
 for _,node:=range sourcehtml.Find(root,func(n *sourcehtml.Node)bool{return n.Tag=="p"||n.Tag=="li"}) {
  if sourcehtml.InNoisy(node,nil) {continue}
  anchors:=sourcehtml.Find(node,func(n *sourcehtml.Node)bool{return n.Tag=="a"})
  if len(anchors)==0 {continue}
  record:=links.ParagraphEvidence{Text:sourcehtml.Text(node)}
  for _,anchor:=range anchors {
   record.Anchors=append(record.Anchors,links.Anchor{
    Text:sourcehtml.Normalize(sourcehtml.Text(anchor)),
    URL:sourcehtml.ResolveURL(anchor.Attr["href"],base),
   })
  }
  evidence=append(evidence,record)
 }
 links.Restore(a,evidence)
}
