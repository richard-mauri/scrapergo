package media

import (
 "net/url"
 "strings"

 "clean-article-reader/internal/article"
)

// Ancestor describes source-DOM context without exposing parser internals.
type Ancestor struct {
 Tag string
 Class string
 Href string
}

// ImageEvidence is an image URL and ordered closest-to-farthest ancestors.
type ImageEvidence struct {
 URL string
 Ancestors []Ancestor
}

// RoleFromEvidence applies publisher-neutral editorial/contributor evidence.
func RoleFromEvidence(ancestors []Ancestor) string {
 for _, ancestor := range ancestors {
  classes := " "+strings.ToLower(ancestor.Class)+" "
  if strings.Contains(classes," imagewrap ") || strings.Contains(classes," has-source-dimensions ") {return "editorial"}
  if strings.Contains(classes," byline ") || strings.Contains(classes," author ") || strings.Contains(classes," contributor ") {return "author"}
  if ancestor.Tag=="a" {
   href,err:=url.Parse(ancestor.Href)
   if err==nil && (strings.Contains(href.Path,"/people/") || strings.Contains(href.Path,"/authors/")) {return "author"}
  }
 }
 return ""
}

// ApplySourceRoles reconciles source observations with existing document blocks.
// Existing lead decisions remain authoritative. Editorial evidence wins over
// author evidence for repeated renditions of an original asset.
func ApplySourceRoles(a *article.Article, observations []ImageEvidence, identity func(string)string) {
 if a==nil {return}
 roles:=map[string]string{}
 for _,o:=range observations {
  key:=identity(o.URL)
  role:=RoleFromEvidence(o.Ancestors)
  if role=="editorial" || (role=="author"&&roles[key]!="editorial") {roles[key]=role}
 }
 for i:=range a.Blocks {
  b:=&a.Blocks[i]
  if b.Kind!="img" || b.Role=="lead" {continue}
  switch roles[identity(b.URL)] {
  case "editorial": b.Role="editorial"
  case "author": b.Role="author"
  }
 }
}
