package main

import "clean-article-reader/internal/media"

// classifyImagesFromSource extracts generic DOM observations; source role policy
// lives in internal/media and does not depend on the application HTML parser.
func classifyImagesFromSource(a *Article, sourceHTML, articleURL string) {
 if a==nil{return}
 root:=parseHTML(sourceHTML)
 observations:=[]media.ImageEvidence{}
 for _,n:=range find(root,func(n *Node)bool{return n.Tag=="img"}) {
  src:=articleImageURL(n,articleURL)
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
 media.ApplySourceRoles(a,observations,imageSourceIdentity)
}
