package main

import (
 "strings"
 "testing"
)

const scopeCaption="The photograph shows the redesigned keyboard and its large trackpad in a bright room."

func TestPhotoScopeSameImageAndSeparateAlternative(t *testing.T) {
 raw:=`<article><div><div><img src="/a.jpg"><span>`+scopeCaption+`</span></div><div><span>`+scopeCaption+`</span></div></div></article>`
 got:=collectPhotoScopes(raw,scopeCaption)
 if len(got)!=2 {t.Fatalf("expected both source presentations, got %+v",got)}
 if got[0].PhotoOrdinal!=1 || got[0].GroupImages!=1 || got[0].ScopeTag!="div" {
  t.Fatalf("image-adjacent caption not linked to local scope: %+v",got[0])
 }
 if got[1].PhotoOrdinal!=1 || got[1].GroupImages!=1 || got[1].ScopeTag!="div" {
  t.Fatalf("alternate caption should share one-image parent group: %+v",got[1])
 }
 if got[0].ScopePath==got[1].ScopePath {
  t.Fatalf("separate branches must not claim identical scope: %+v",got)
 }
}

func TestPhotoScopeTwoIndependentPhotographsSameCaption(t *testing.T) {
 raw:=`<article><section><div><img src="/a.jpg"><span>`+scopeCaption+`</span></div><div><img src="/b.jpg"><span>`+scopeCaption+`</span></div></section></article>`
 got:=collectPhotoScopes(raw,scopeCaption)
 if len(got)!=2 || got[0].PhotoOrdinal!=1 || got[1].PhotoOrdinal!=2 {
  t.Fatalf("identical captions on two independent images must stay distinct: %+v",got)
 }
}

func TestPhotoScopeOrdinaryRepeatedProseHasNoImageOwner(t *testing.T) {
 raw:=`<article><p>`+scopeCaption+`</p><p>`+scopeCaption+`</p><img src="/unrelated.jpg"></article>`
 got:=collectPhotoScopes(raw,scopeCaption)
 if len(got)!=2 || got[0].PhotoOrdinal!=0 || got[1].PhotoOrdinal!=0 || got[0].GroupImages!=0 {
  t.Fatalf("article-wide image must not be treated as prose owner: %+v",got)
 }
}

func TestPhotoScopeGalleryIsNotSingleImageOwner(t *testing.T) {
 raw:=`<article><section><img src="/a.jpg"><img src="/b.jpg"><span>`+scopeCaption+`</span></section></article>`
 got:=collectPhotoScopes(raw,scopeCaption)
 if len(got)!=1 || got[0].PhotoOrdinal!=0 || got[0].GroupImages!=2 {
  t.Fatalf("ambiguous two-image group should not have a single owner: %+v",got)
 }
}

func TestPhotoScopeDoesNotReportSensitiveAttributeValues(t *testing.T) {
 raw:=`<article><div id="private-999" class="sensitive-layout"><img src="https://private.example/secret.jpg"><span>`+scopeCaption+`</span></div></article>`
 got:=collectPhotoScopes(raw,scopeCaption)
 if len(got)!=1 {t.Fatalf("expected match: %+v",got)}
 for _,v:=range []string{got[0].ScopePath,got[0].ScopeTag,got[0].MatchPath} {
  if strings.Contains(v,"private") || strings.Contains(v,"secret") || strings.Contains(v,"sensitive") {
   t.Fatalf("sensitive HTML metadata leaked: %s",v)
  }
 }
}
