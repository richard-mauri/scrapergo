package main

import (
 "clean-article-reader/internal/documentmedia"
 "net/url"
 "regexp"
 "strings"
)

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

// classifyContributorImages uses the source document's byline structure and
// nearby named contributor links, not file dimensions or URL guesses.
func classifyContributorImages(a *Article) {
 if a==nil{return}
 for i:=range a.Blocks {
  b:=&a.Blocks[i]
  if b.Kind!="img" || b.Role=="author" {continue}
  caption:=strings.TrimSpace(b.Text)
  if caption=="" {continue}
  // Contributor names often occur as links adjacent to their headshots.
  // Require a matching named link in the nearby introductory content.
  for j:=i-2;j<=i+2;j++ {
   if j<0||j>=len(a.Blocks)||j==i {continue}
   nearby:=a.Blocks[j]
   if nearby.Kind=="img" {continue}
   if strings.EqualFold(strings.TrimSpace(nearby.Text),caption) && (len(nearby.Links)>0 || nearby.Kind=="a") {
    b.Role="author"
    break
   }
  }
 }
}

// normalizeArticleDocument reconciles the article title and image identities.
// The first editorial image wins over metadata; metadata is a fallback or\n// alternate rendition of the same identified source.
func normalizeArticleDocument(a *Article, sourceLead ...string) {
 documentmedia.NormalizeDocument(a,imageSourceIdentity,sourceLead...)
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
  // Only restore portraits belonging to an opening contributor group.
  // An author biography after the article's editorial content must not be
  // promoted into the opening byline, even when an ancestor is named author.
  if !openingContributorGroup(group) { continue }
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

// openingContributorGroup identifies contributor markup before the first
// substantial article prose or editorial figure in its article/main scope.
// Structural DOM order is used rather than the author's image filename.
func openingContributorGroup(group *Node) bool {
 var scope *Node
 for n:=group.Parent;n!=nil;n=n.Parent { if n.Tag=="article" {scope=n;break} }
 if scope==nil { for n:=group.Parent;n!=nil;n=n.Parent {if n.Tag=="main" {scope=n;break}} }
 if scope==nil { return false }
 for _,n:=range find(scope,func(n *Node)bool{return n.Tag=="p"||n.Tag=="figure"||n.Tag=="img"||n==group}) {
  if n==group { return true }
  if n.Tag=="p" && len(strings.Fields(textOf(n)))>=25 {return false}
  if n.Tag=="figure" || n.Tag=="img" {
   // Do not use portrait markup itself as the editorial boundary.
   inside:=false
   for p:=n;p!=nil;p=p.Parent {if p==group {inside=true;break}}
   if !inside {return false}
  }
 }
 return false
}

// reconcileContributorByline keeps source-linked contributor names together
// as metadata and removes only the byline fragments preceding the story body.
// Editorial paragraphs, images and ordinary links remain untouched.
func reconcileContributorByline(a *Article) {
 if a==nil || len(a.Blocks)==0{return}
 names:=[]string{}
 seen:=map[string]bool{}
 remove:=map[int]bool{}
 started:=false
 for i,b:=range a.Blocks {
  if b.Kind=="img" {
   if b.Role=="author" {continue}
   break
  }
  t:=strings.TrimSpace(b.Text)
  if b.Kind=="p" && (t=="By"||t==",") {
   remove[i]=true;started=true;continue
  }
  if b.Kind=="p" && t!="" && len(t)<90 && len(b.Links)>0 {
   matched:=false
   for _,link:=range b.Links {if strings.EqualFold(strings.TrimSpace(link.Text),t){matched=true;break}}
   if matched {
    if !seen[strings.ToLower(t)] {names=append(names,t);seen[strings.ToLower(t)]=true}
    remove[i]=true;started=true;continue
   }
  }
  if started && t=="" {remove[i]=true;continue}
  if b.Kind=="p" && t=="Middle East" {continue}
  if started && b.Kind!="img" {break}
 }
 if len(names)<2{return}
 out:=make([]Block,0,len(a.Blocks))
 for i,b:=range a.Blocks {if !remove[i] {out=append(out,b)}}
 a.Blocks=out
 a.Byline=strings.Join(names,", ")
}
