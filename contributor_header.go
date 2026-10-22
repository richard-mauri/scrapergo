package main

import "strings"

// separateContributorHeader consumes the source-confirmed introductory byline
// segment, stopping at the first editorial figure or actual story paragraph.
// Never remove images or content after that boundary.
func separateContributorHeader(a *Article) {
 if a==nil{return}
 boundary:=-1
 for i,b:=range a.Blocks {
  if b.Kind=="img" && (b.Role=="editorial"||b.Role=="lead") {boundary=i;break}
 }
 if boundary<0{return}
 names:=[]string{}
 seen:=map[string]bool{}
 remove:=map[int]bool{}
 inByline:=false
 for i:=0;i<boundary;i++ {
  b:=a.Blocks[i]
  t:=strings.TrimSpace(b.Text)
  if b.Kind=="p" && strings.EqualFold(t,"By") {inByline=true;remove[i]=true;continue}
  if !inByline {continue}
  if b.Kind=="img" && b.Role=="author" {remove[i]=true;continue}
  if b.Kind=="p" && (t==","||t=="") {remove[i]=true;continue}
  if b.Kind=="p" && t!="" && len(b.Links)>0 {
   isName:=false
   for _,l:=range b.Links {if strings.EqualFold(strings.TrimSpace(l.Text),t) {isName=true;break}}
   if isName {
    if !seen[strings.ToLower(t)] {names=append(names,t);seen[strings.ToLower(t)]=true}
    remove[i]=true;continue
   }
  }
  // Audio controls and section headings mark the end of byline parsing.
  if t!="" {break}
 }
 if len(names)<2 {return}
 out:=make([]Block,0,len(a.Blocks))
 for i,b:=range a.Blocks {if !remove[i] {out=append(out,b)}}
 a.Blocks=out
 a.Byline=strings.Join(names,", ")
}
