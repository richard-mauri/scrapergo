package recovery

import (
 "clean-article-reader/internal/media"
 "clean-article-reader/internal/article"
 "clean-article-reader/internal/sourcehtml"
)

// classifyImagesFromSource extracts generic DOM observations; source role policy
// lives in internal/media and does not depend on the application HTML parser.
func ClassifyImagesFromSource(a *article.Article, sourceHTML, articleURL string, resolveImage func(*sourcehtml.Node,string) string, identity func(string)string) {
 if a==nil{return}
 root:=sourcehtml.Parse(sourceHTML)
 observations:=[]media.ImageEvidence{}
 for _,n:=range sourcehtml.Find(root,func(n *sourcehtml.Node)bool{return n.Tag=="img"}) {
  src:=resolveImage(n,articleURL)
  if src=="" {continue}
  observation:=media.ImageEvidence{URL:src}
  for ancestor:=n.Parent;ancestor!=nil;ancestor=ancestor.Parent {
   observation.Ancestors=append(observation.Ancestors,media.Ancestor{
    Tag:ancestor.Tag,
    Class:ancestor.Attr["class"],
    Href:ancestor.Attr["href"],
   })
  }
  observations=append(observations,observation)
 }
 media.ApplySourceRoles(a,observations,identity)
}
