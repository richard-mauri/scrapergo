package main

import (
 "bytes"
 "regexp"
 "strings"

 "golang.org/x/net/html"
)

var escapedMediaElementRE = regexp.MustCompile(`(?is)<\s*/?\s*(?:iframe|audio|video|embed|object)\b[^>]*>`)

// stripEmbeddedMedia removes actual non-editorial media DOM nodes before
// Readability serializes HTML to text. Sibling article prose is preserved.
func stripEmbeddedMedia(raw string) string {
 root,err:=html.Parse(strings.NewReader(raw))
 if err!=nil{return raw}
 var visit func(*html.Node)
 visit=func(n *html.Node){
  if n.Type==html.TextNode && escapedMediaElementRE.MatchString(n.Data) {
    insideCode:=false
    for p:=n.Parent;p!=nil;p=p.Parent { if p.Type==html.ElementNode&&(p.Data=="pre"||p.Data=="code"){insideCode=true;break} }
    if !insideCode {n.Data=escapedMediaElementRE.ReplaceAllString(n.Data,"")}
  }
  for child:=n.FirstChild;child!=nil;{
   next:=child.NextSibling
   if child.Type==html.ElementNode && (child.Data=="iframe"||child.Data=="embed"||child.Data=="audio"||child.Data=="video"||child.Data=="object"){
    n.RemoveChild(child)
   }else{visit(child)}
   child=next
  }
 }
 visit(root)
 var b bytes.Buffer
 if err:=html.Render(&b,root);err!=nil{return raw}
 return b.String()
}

// normalizeArticleDocument creates a single renderable stream. ImageURL remains
// available as metadata for existing API clients, but any selected lead image
// is represented by precisely one Block with Role="lead".
func normalizeArticleDocument(a *Article, sourceLead ...string) {
 if a==nil{return}
 title:=strings.Join(strings.Fields(strings.ToLower(a.Title))," ")
 for i,b:=range a.Blocks {
  if b.Kind=="img"{continue}
  if (b.Kind=="h1"||b.Kind=="h2"||b.Kind=="h3"||b.Kind=="h4"||b.Kind=="p")&&strings.Join(strings.Fields(strings.ToLower(b.Text))," ")==title {
   a.Blocks=append(a.Blocks[:i],a.Blocks[i+1:]...)
  }
  break
 }
 lead:=""
 if len(sourceLead)>0 {lead=strings.TrimSpace(sourceLead[0])}
 if lead=="" {lead=a.ImageURL}
 if lead=="" {return}
 for i:=range a.Blocks {
  if a.Blocks[i].Kind=="img"&&a.Blocks[i].URL==lead {
   a.Blocks[i].Role="lead"
   return
  }
 }
 a.Blocks=append([]Block{{Kind:"img",URL:lead,Role:"lead"}},a.Blocks...)
}

// recoverAuthorPortrait retains a photo identified by the source document's
// byline/author structure. It never treats a generic inline photograph as an
// author portrait based on its dimensions or URL naming.
func recoverAuthorPortrait(a *Article, sourceHTML, baseURL string) {
 if a==nil{return}
 root:=parseHTML(sourceHTML)
 byline:=regexp.MustCompile(`(?i)(?:^|[\\s_-])(?:author|byline|contributor)(?:$|[\\s_-])`)
 candidates:=find(root,func(n *Node)bool{
  return n.Tag=="div"||n.Tag=="section"||n.Tag=="span"||n.Tag=="figure"
 })
 for _,group:=range candidates{
  if !byline.MatchString(group.Attr["class"]+" "+group.Attr["id"]){continue}
  photos:=find(group,func(n *Node)bool{return n.Tag=="img"})
  for _,p:=range photos{
   src:=articleImageURL(p,baseURL)
   if src==""{continue}
   for i:=range a.Blocks{
    if a.Blocks[i].Kind=="img"&&a.Blocks[i].URL==src {
     a.Blocks[i].Role="author"
     return
    }
   }
   a.Blocks=append([]Block{{Kind:"img",Role:"author",URL:src,Text:normalize(p.Attr["alt"])}},a.Blocks...)
   return
  }
 }
}
