package main

import (
 "net/url"
 "strings"
)

// classifyImagesFromSource maps recovered pictures back to the original HTML
// via the same source-image identity used by deduplication. Editorial
// imagewrap containers take precedence over unclassified contributor art.
func classifyImagesFromSource(a *Article, sourceHTML, articleURL string) {
 if a==nil{return}
 root:=parseHTML(sourceHTML)
 roles:=map[string]string{}
 for _,n:=range find(root,func(n *Node)bool{return n.Tag=="img"}) {
  src:=articleImageURL(n,articleURL)
  if src==""{continue}
  key:=imageSourceIdentity(src)
  role:=""
  for ancestor:=n.Parent;ancestor!=nil;ancestor=ancestor.Parent{
   classes:=" "+strings.ToLower(ancestor.Attr["class"])+" "
   if strings.Contains(classes," imagewrap ") || strings.Contains(classes," has-source-dimensions ") {
    role="editorial"
    break
   }
   if strings.Contains(classes," byline ") || strings.Contains(classes," author ") || strings.Contains(classes," contributor ") {
    role="author"
    break
   }
   if ancestor.Tag=="a" {
    href,err:=url.Parse(ancestor.Attr["href"])
    if err==nil && (strings.Contains(href.Path,"/people/")||strings.Contains(href.Path,"/authors/")) {
     role="author"
     break
    }
   }
  }
  if role=="editorial" || (role=="author" && roles[key]!="editorial") {
   roles[key]=role
  }
 }
 for i:=range a.Blocks {
  b:=&a.Blocks[i]
  if b.Kind!="img" {continue}
  switch roles[imageSourceIdentity(b.URL)] {
  case "editorial":
   if b.Role!="lead" {b.Role="editorial"}
  case "author":
   if b.Role!="lead" {b.Role="author"}
  }
 }
}
