package pageclassify

import (
 "clean-article-reader/internal/sourcehtml"
 "net/url"
 "strings"
)

// IndexEvidence exposes the counters used by IsIndex for diagnostics only.
// It does not change production classification.
func IndexEvidence(raw string) map[string]any {
 root:=sourcehtml.Parse(raw)
 scope:=root
 for _,n:=range sourcehtml.Find(root,func(n *sourcehtml.Node)bool{return n.Tag=="main"}) {
  if !sourcehtml.InNoisy(n,nil) {scope=n;break}
 }
 links,headlineLinks,prose:=0,0,0
 targets:=map[string]bool{}
 for _,n:=range sourcehtml.Find(scope,func(n *sourcehtml.Node)bool{return n.Tag=="a"||n.Tag=="p"}) {
  if sourcehtml.InNoisy(n,scope) {continue}
  switch n.Tag {
  case "p":
   linked:=false
   for p:=n.Parent;p!=nil&&p!=scope;p=p.Parent {if p.Tag=="a" {linked=true;break}}
   if !linked&&len(sourcehtml.Normalize(sourcehtml.Text(n)))>=150 {prose++}
  case "a":
   label:=sourcehtml.Normalize(sourcehtml.Text(n))
   if len(label)<12||len(label)>240 {continue}
   href:=strings.TrimSpace(n.Attr["href"])
   if href==""||strings.HasPrefix(href,"#") {continue}
   parsed,err:=url.Parse(href)
   if err!=nil||(parsed.Scheme!=""&&parsed.Scheme!="https"&&parsed.Scheme!="http") {continue}
   path:=strings.TrimRight(parsed.Path,"/")
   if path=="" {continue}
   links++;targets[path]=true
   isHeadline:=len(label)>=25&&len(label)<=180
   for _,child:=range sourcehtml.Find(n,func(x *sourcehtml.Node)bool{return x!=n&&(x.Tag=="h2"||x.Tag=="h3"||x.Tag=="h4")}) {
    if child!=nil {isHeadline=true;break}
   }
   for p:=n.Parent;p!=nil&&p!=scope;p=p.Parent {
    if p.Tag=="h2"||p.Tag=="h3"||p.Tag=="h4" {isHeadline=true;break}
   }
   if isHeadline {headlineLinks++}
  }
 }
 distinct:=len(targets)
 threshold:=distinct>=12&&links>=18&&headlineLinks>=8
 index:=threshold&&prose<7&&((headlineLinks>=12&&distinct>=18&&links>=24&&prose<=5)||(headlineLinks>=8&&distinct>=12&&links>=18&&prose<=2))
 return map[string]any{
  "scope":scope.Tag,"links":links,"distinctDestinations":distinct,"headlineLinks":headlineLinks,
  "substantialProseParagraphs":prose,"minimumLinkThresholdMet":threshold,"index":index,
 }
}
