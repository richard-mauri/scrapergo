package media

import (
 "testing"

 "clean-article-reader/internal/article"
)

func TestRoleFromEvidenceUsesNearestAncestor(t *testing.T) {
 cases:=[]struct{name string;parents []Ancestor;want string}{
  {"editorial",[]Ancestor{{Tag:"figure",Class:"imagewrap"}},"editorial"},
  {"byline",[]Ancestor{{Tag:"div",Class:"byline"}},"author"},
  {"author link",[]Ancestor{{Tag:"a",Href:"/people/reporter"}},"author"},
  {"inner editorial wins",[]Ancestor{{Tag:"picture",Class:"has-source-dimensions"},{Tag:"div",Class:"author"}},"editorial"},
  {"no context",[]Ancestor{{Tag:"div",Class:"article"}},""},
 }
 for _,tc:=range cases {t.Run(tc.name,func(t *testing.T){if got:=RoleFromEvidence(tc.parents);got!=tc.want {t.Fatalf("got %q, want %q",got,tc.want)}})}
}

func TestApplySourceRolesPreservesLeadAndEditorialPrecedence(t *testing.T) {
 same:=func(s string)string{return s}
 a:=article.Article{Blocks:[]article.Block{
  {Kind:"img",URL:"https://example.org/lead",Role:"lead"},
  {Kind:"img",URL:"https://example.org/shared"},
  {Kind:"p",Text:"prose"},
 }}
 observations:=[]ImageEvidence{
  {URL:"https://example.org/lead",Ancestors:[]Ancestor{{Class:"author"}}},
  {URL:"https://example.org/shared",Ancestors:[]Ancestor{{Class:"author"}}},
  {URL:"https://example.org/shared",Ancestors:[]Ancestor{{Class:"imagewrap"}}},
 }
 ApplySourceRoles(&a,observations,same)
 if a.Blocks[0].Role!="lead" || a.Blocks[1].Role!="editorial" || a.Blocks[2].Kind!="p" {t.Fatalf("incorrect reconciliation: %+v",a.Blocks)}
}
