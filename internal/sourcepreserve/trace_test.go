package sourcepreserve

import (
 "encoding/json"
 "strings"
 "testing"
)

func TestTraceStructureGroupsImagesWithoutExposingSourceContent(t *testing.T){
 raw:=`<article><h2>Private section</h2><div><div><img src="/one.jpg"></div><div><img src="/two.jpg"></div><div><span>Publisher caption</span></div></div><div><div><img src="/three.jpg"></div><div><span>Different caption</span></div></div><p>Editorial paragraph</p></article>`
 got,err:=TraceStructure(raw,baseURL)
 if err!=nil {t.Fatal(err)}
 if got.Status!="candidate"||len(got.Images)!=3 {t.Fatalf("unexpected image trace: %+v",got)}
 if got.Images[0].Pattern!="bounded-multi-image-group"||got.Images[1].Pattern!="bounded-multi-image-group"||got.Images[0].GroupIndex!=1||got.Images[1].GroupIndex!=1||got.Images[0].GroupImages!=2||got.Images[0].GroupImageBranches!=2||got.Images[0].GroupTextBranches!=1 {t.Fatalf("group shape lost: %+v",got.Images)}
 if got.Images[2].Pattern!="compact-text-sibling"||got.Images[2].GroupIndex!=0 {t.Fatalf("single image wrongly grouped: %+v",got.Images[2])}
 data,err:=json.Marshal(got)
 if err!=nil {t.Fatal(err)}
 for _,s:=range []string{"Publisher caption","Editorial paragraph","Private section","one.jpg","two.jpg","three.jpg","/one","/two","/three"}{
  if strings.Contains(string(data),s){t.Fatalf("structural output leaked source content %q: %s",s,data)}
 }
}

func TestTraceStructureRejectsAmbiguousArticleAndUnsafeBase(t *testing.T){
 got,err:=TraceStructure(`<article><p>One</p></article><article><p>Two</p></article>`,baseURL)
 if err!=nil||got.Status!="inconclusive"||len(got.Images)!=0 {t.Fatalf("multiple article scopes accepted: %+v %v",got,err)}
 _,err=TraceStructure(`<article><img src="/x.jpg"></article>`,"javascript:bad")
 if err==nil {t.Fatal("unsafe URL accepted")}
}

func TestTraceStructureRespectsExcludedAndFigureBranches(t *testing.T){
 raw:=`<article><div class="ad"><img src="/ad.jpg"></div><div hidden><img src="/hidden.jpg"></div><figure><img src="/fig.jpg"><figcaption>Figure caption</figcaption></figure><div><img src="/unwrapped.jpg"></div></article>`
 got,err:=TraceStructure(raw,baseURL)
 if err!=nil{t.Fatal(err)}
 if len(got.Images)!=1||got.Images[0].ImageIndex!=1||got.Images[0].Pattern!="unwrapped"{
  t.Fatalf("excluded/figure images leaked into trace: %+v",got)
 }
}
