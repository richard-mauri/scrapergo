package main

import (
 "bytes"
 "net/url"
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

// imageSourceIdentity identifies renditions only when the CDN supplies
// an explicit original asset URL. It never guesses by filename or dimensions.
func imageSourceIdentity(raw string) string {
 u,err:=url.Parse(raw)
 if err!=nil || (u.Scheme!="https"&&u.Scheme!="http") || u.Hostname()=="" {return raw}
 if strings.EqualFold(u.Hostname(),"npr.brightspotcdn.com") {
  embedded:=u.Query().Get("url")
  if asset,err:=url.Parse(embedded);err==nil && (asset.Scheme=="http"||asset.Scheme=="https") && asset.Hostname()!="" {
   asset.Fragment=""
   return asset.String()
  }
 }
 u.Fragment=""
 return u.String()
}

// normalizeArticleDocument reconciles the article title and image identities.
// A publisher lead rendition wins, while captions and credits survive from
// the matching inline figure. API imageUrl metadata is kept for compatibility.
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
 // Collect a single block per source identity, retaining order. Distinct
 // author photos must never be suppressed by an unrelated lead image.
 out:=make([]Block,0,len(a.Blocks)+1)
 seen:=make(map[string]int)
 for _,b:=range a.Blocks {
  if b.Kind!="img" {out=append(out,b);continue}
  key:=imageSourceIdentity(b.URL)
  if prev,ok:=seen[key];ok{
   existing:=&out[prev]
   if existing.Text=="" {existing.Text=b.Text}
   if existing.Role=="" {existing.Role=b.Role}
   continue
  }
  seen[key]=len(out)
  out=append(out,b)
 }
 if lead!="" {
  key:=imageSourceIdentity(lead)
  if prev,ok:=seen[key];ok {
   out[prev].URL=lead
   out[prev].Role="lead"
  } else {
   out=append([]Block{{Kind:"img",URL:lead,Role:"lead"}},out...)
  }
 }
 a.Blocks=out
}

// recoverAuthorPortrait retains a photo identified by the source document's
// byline/author structure. It never treats a generic inline photograph as an
// author portrait based on its dimensions or URL naming.
func recoverAuthorPortrait(a *Article, sourceHTML, baseURL string) {
 if a==nil{return}
 root:=parseHTML(sourceHTML)
 byline:=regexp.MustCompile(`(?i)(?:^|[\s_-])(?:author|byline|contributor)(?:$|[\s_-])`)
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
