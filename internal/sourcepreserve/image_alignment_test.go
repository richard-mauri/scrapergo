package sourcepreserve

import (
 "encoding/json"
 "strings"
 "testing"
)
func alignment(t *testing.T,src,selected string) ImageAlignment {
 t.Helper()
 doc,err:=ConvertSelectedHTML(selected,baseURL)
 if err!=nil{t.Fatal(err)}
 r,err:=CompareImageAlignment(src,baseURL,doc)
 if err!=nil{t.Fatal(err)}
 return r
}
func TestImageAlignmentExactIdentityAndSourceOrder(t *testing.T){
 source:=`<html><nav><img src="/not-article.jpg"></nav><article><figure><img src="/a.jpg"><figcaption>A <cite>Credit</cite></figcaption></figure><div><img src="/b.jpg"></div><img src="/c.jpg"></article><aside><img src="/sidebar.jpg"></aside></html>`
 selected:=`<article><figure><img src="https://example.org/a.jpg"><figcaption>A</figcaption></figure><img src="/b.jpg"><img src="/c.jpg"></article>`
 got:=alignment(t,source,selected)
 if got.SourceStatus!="unique-article"||got.SourceImages!=3||got.SelectedImages!=3||got.ExactURLMatches!=3||got.SourceOnly!=0||got.SelectedOnly!=0||!got.OrderPreserved||got.SourceExplicitCaptionImages!=1||got.SelectedExplicitCaptionImages!=1 {t.Fatalf("incorrect match: %+v",got)}
 for i,m:=range got.Matches{if m.SourceIndex!=i+1||m.SelectedIndex!=i+1{t.Fatalf("wrong ordinal: %+v",got.Matches)}}
 b,_:=json.Marshal(got)
 if strings.Contains(string(b),"a.jpg")||strings.Contains(string(b),"Credit")||strings.Contains(string(b),"sidebar.jpg"){t.Fatalf("source leaked: %s",b)}
}
func TestImageAlignmentDetectsReorderAndMissingImages(t *testing.T){
 src:=`<article><img src="/a.jpg"><img src="/b.jpg"><img src="/c.jpg"></article>`
 selected:=`<article><img src="/c.jpg"><img src="/a.jpg"><img src="/new.jpg"></article>`
 got:=alignment(t,src,selected)
 if got.ExactURLMatches!=2||got.SourceOnly!=1||got.SelectedOnly!=1||got.OrderPreserved||len(got.Matches)!=2{t.Fatalf("reordering not detected: %+v",got)}
 if got.Matches[0].SourceIndex!=3||got.Matches[0].SelectedIndex!=1||got.Matches[1].SourceIndex!=1||got.Matches[1].SelectedIndex!=2 {t.Fatalf("wrong mapping: %+v",got.Matches)}
}
func TestImageAlignmentDuplicateURLsMatchOneToOne(t *testing.T){
 src:=`<article><img src="/a.jpg"><img src="/a.jpg"><img src="/b.jpg"></article>`
 selected:=`<article><img src="/a.jpg"><img src="/b.jpg"><img src="/a.jpg"><img src="/a.jpg"></article>`
 got:=alignment(t,src,selected)
 if got.ExactURLMatches!=3||got.SourceOnly!=0||got.SelectedOnly!=1||got.OrderPreserved {t.Fatalf("repeat URL matched improperly: %+v",got)}
 if got.Matches[0].SourceIndex!=1||got.Matches[1].SourceIndex!=3||got.Matches[2].SourceIndex!=2{t.Fatalf("occurrence matching wrong: %+v",got.Matches)}
}
func TestImageAlignmentRespectsExplicitSourceArticleScope(t *testing.T){
 doc,err:=ConvertSelectedHTML(`<article><img src="/a.jpg"></article>`,baseURL);if err!=nil{t.Fatal(err)}
 for _,src:=range []string{`<main><img src="/a.jpg"></main>`,`<article><img src="/a.jpg"></article><article><img src="/b.jpg"></article>`} {
  got,err:=CompareImageAlignment(src,baseURL,doc);if err!=nil{t.Fatal(err)}
  if got.SourceStatus!="inconclusive"||got.ExactURLMatches!=0||got.SelectedImages!=1||got.SourceImages!=0{t.Fatalf("guessed source scope: %+v",got)}
 }
}
func TestImageAlignmentDoesNotMatchUnsafeOrVariantURLs(t *testing.T){
 src:=`<article><img src="/a.jpg?width=400"><img src="javascript:bad"><img src="/b.jpg"></article>`
 selected:=`<article><img src="/a.jpg?width=800"><img src="/b.jpg"></article>`
 got:=alignment(t,src,selected)
 if got.SourceImages!=2||got.SelectedImages!=2||got.ExactURLMatches!=1||got.SourceOnly!=1||got.SelectedOnly!=1{t.Fatalf("variant/unsafe mismatch: %+v",got)}
}
