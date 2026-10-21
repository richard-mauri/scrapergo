package main

import (
 "bytes"
 "regexp"
 "strings"

 "golang.org/x/net/html"
)

var escapedMediaElementRE = regexp.MustCompile(`(?is)<\\s*/?\\s*(?:iframe|audio|video|embed|object)\\b[^>]*>`)

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

// normalizeArticleDocument resolves duplicate metadata and content structure
// before any UI renders it. The image URL is kept as metadata if it isn't
// the exact identity of an editorial figure, allowing a separate lead.
func normalizeArticleDocument(a *Article) {
 if a==nil{return}
 title:=strings.Join(strings.Fields(strings.ToLower(a.Title))," ")
 for i,b:=range a.Blocks {
  if b.Kind=="img"{continue}
  if (b.Kind=="h1"||b.Kind=="h2"||b.Kind=="h3"||b.Kind=="h4")&&strings.Join(strings.Fields(strings.ToLower(b.Text))," ")==title {
   a.Blocks=append(a.Blocks[:i],a.Blocks[i+1:]...)
  }
  break
 }
 if a.ImageURL!="" {
  for _,b:=range a.Blocks {if b.Kind=="img"&&b.URL==a.ImageURL {a.ImageURL="";break}}
 }
}
