package sourcepreserve

import (
 "strings"
 "testing"
)

const baseURL="https://example.org/story"

func TestInspectPreservesPairedFiguresAndSourceOrder(t *testing.T){
 raw:=`<html><body><nav>Navigation</nav><article><h1>Review</h1><p>Introduction <a href="/facts">sources</a>.</p><div class="gallery"><figure><img src="/lid.jpg"><figcaption>Apple lid <small>Photo credit: A</small></figcaption></figure><figure><img src="/keys.jpg"><figcaption>Keyboard <small>Photo credit: B</small></figcaption></figure></div><p>Independent conclusion.</p></article><div class="ad">Advertisement</div></body></html>`
 got,err:=Inspect(raw,baseURL)
 if err!=nil{t.Fatal(err)}
 if got.Status!="candidate"||len(got.Items)!=5||got.AmbiguousFigures!=0{t.Fatalf("bad report: %+v",got)}
 if got.Items[0].Kind!="h1"||got.Items[1].Kind!="p"||len(got.Items[1].Links)!=1||got.Items[1].Links[0]!="https://example.org/facts" {t.Fatalf("heading/link lost: %+v",got.Items)}
 if got.Items[2].ImageURL!="https://example.org/lid.jpg"||got.Items[3].ImageURL!="https://example.org/keys.jpg" {t.Fatalf("figure order lost: %+v",got.Items)}
 if !strings.Contains(got.Items[2].Caption,"Apple lid")||got.Items[2].Credit!="Photo credit: A"||!strings.Contains(got.Items[3].Caption,"Keyboard")||got.Items[3].Credit!="Photo credit: B" {t.Fatalf("figure attribution lost: %+v",got.Items)}
 if got.Items[4].Text!="Independent conclusion." {t.Fatalf("prose lost: %+v",got.Items)}
}

func TestInspectFiltersAdPopupAndUnsafeMarkupWithoutDeletingProse(t *testing.T){
 raw:=`<article><p>A legitimate advertisement study remains editorial.</p><div class="advertisement"><p>Buy now</p></div><div role="dialog"><p>Popup</p></div><div hidden><p>Hidden alternative</p></div><div aria-hidden="true"><p>Aria invisible</p></div><p><a href="javascript:alert(1)">unsafe</a> and <a href="https://example.org/ok">safe</a>.</p><script>alert(1)</script></article>`
 got,err:=Inspect(raw,baseURL)
 if err!=nil{t.Fatal(err)}
 if len(got.Items)!=2||got.Items[0].Text!="A legitimate advertisement study remains editorial." {t.Fatalf("filter failed: %+v",got)}
 if strings.Contains(got.Items[1].Text,"alert")||len(got.Items[1].Links)!=1||got.Items[1].Links[0]!="https://example.org/ok" {t.Fatalf("unsafe link/script survived: %+v",got.Items[1])}
}

func TestInspectDoesNotCollapseDistinctImagesSharingCaption(t *testing.T){
 raw:=`<article><figure><img src="/a.jpg"><figcaption>Same caption</figcaption></figure><figure><img src="/b.jpg"><figcaption>Same caption</figcaption></figure><p>Intentional repeated quote.</p><p>Intentional repeated quote.</p></article>`
 got,err:=Inspect(raw,baseURL)
 if err!=nil{t.Fatal(err)}
 if len(got.Items)!=4||got.Items[0].ImageURL==got.Items[1].ImageURL||got.Items[0].Caption!=got.Items[1].Caption||got.Items[2].Text!=got.Items[3].Text {t.Fatalf("legitimate repetition lost: %+v",got)}
}

func TestInspectReportsAmbiguousImageOwnership(t *testing.T){
 raw:=`<article><figure><img src="/one.jpg"><img src="/two.jpg"><figcaption>Which image?</figcaption></figure><p>Retain article text.</p></article>`
 got,err:=Inspect(raw,baseURL)
 if err!=nil{t.Fatal(err)}
 if got.AmbiguousFigures!=1||len(got.Items)!=1||got.Items[0].Text!="Retain article text."{t.Fatalf("ambiguous association guessed: %+v",got)}
}

func TestInspectRejectsMultipleArticleScopesAndInvalidURL(t *testing.T){
 got,err:=Inspect(`<article><p>First</p></article><article><p>Second</p></article>`,baseURL)
 if err!=nil||got.Status!="inconclusive"||len(got.Items)!=0 {t.Fatalf("incorrect multi-article selection: %+v %v",got,err)}
 if _,err:=Inspect(`<article><p>OK</p></article>`,"javascript:alert(1)");err==nil{t.Fatal("unsafe base url accepted")}
}

func TestInspectDoesNotTrustUserinfoOrUnsafeImageScheme(t *testing.T){
 raw:=`<article><figure><img src="javascript:alert(1)"><figcaption>Unsafe</figcaption></figure><figure><img src="https://user:pw@example.org/photo.jpg"><figcaption>Credentials</figcaption></figure><p>Safe prose.</p></article>`
 got,err:=Inspect(raw,baseURL)
 if err!=nil{t.Fatal(err)}
 if got.AmbiguousFigures!=2||len(got.Items)!=1||got.Items[0].Text!="Safe prose."{t.Fatalf("unsafe image included: %+v",got)}
}

func TestInspectClassifiesUnwrappedAndMultiImageStructures(t *testing.T){
 raw:=`<article><p>Ordinary prose remains.</p><div><img src="/unwrapped.jpg"></div><figure><img src="/first.jpg"><img src="/second.jpg"><figcaption>Two images, uncertain caption.</figcaption></figure><figure><img src="/one.jpg"><figcaption>First caption</figcaption><figcaption>Second caption</figcaption></figure><figure><img src="javascript:bad"><figcaption>Unsafe image</figcaption></figure></article>`
 got,err:=Inspect(raw,baseURL)
 if err!=nil{t.Fatal(err)}
 if got.AmbiguousFigures!=4 || got.UnwrappedImages!=1 || got.MultiImageFigures!=1 || got.MultipleCaptions!=1 || got.InvalidFigureImages!=1 {
  t.Fatalf("incorrect reason counts: %+v",got)
 }
 if len(got.Items)!=1 || got.Items[0].Text!="Ordinary prose remains." {t.Fatalf("ineligible figure content promoted to prose: %+v",got.Items)}
}

func TestInspectUnwrappedImageDoesNotGuessCaptionsByProximity(t *testing.T){
 raw:=`<article><div><img src="/a.jpg"><span>Photographer attribution</span></div><p>Independent editorial text.</p></article>`
 got,err:=Inspect(raw,baseURL)
 if err!=nil{t.Fatal(err)}
 if got.UnwrappedImages!=1 || got.AmbiguousFigures!=1 || len(got.Items)!=1 || got.Items[0].Text!="Independent editorial text." {
  t.Fatalf("inferred image caption or dropped prose: %+v",got)
 }
}

func TestInspectCompactSingleImageSiblingTextDiagnostic(t *testing.T){
 raw:=`<article><div><div><img src="/a.jpg"></div><div><span>Caption A</span></div></div><div><div><img src="/b.jpg"></div><div><span>Caption B</span></div></div><p>Editorial paragraph.</p></article>`
 got,err:=Inspect(raw,baseURL)
 if err!=nil {t.Fatal(err)}
 if got.UnwrappedImages!=2 || got.CompactImageTextWrappers!=2 || got.OtherUnwrappedImages!=0 || got.AmbiguousFigures!=2 {
  t.Fatalf("single-image text sibling evidence was not identified: %+v",got)
 }
 if len(got.Items)!=1||got.Items[0].Text!="Editorial paragraph." {t.Fatalf("diagnostic changed extracted prose: %+v",got.Items)}
}

func TestInspectDoesNotInferCaptionFromBroadMultiImageGallery(t *testing.T){
 raw:=`<article><div><div><img src="/left.jpg"></div><div><img src="/right.jpg"></div><div><span>Shared non-caption editorial text.</span></div></div><p>Independent analysis.</p></article>`
 got,err:=Inspect(raw,baseURL)
 if err!=nil {t.Fatal(err)}
 if got.UnwrappedImages!=2 || got.CompactImageTextWrappers!=0 || got.OtherUnwrappedImages!=2 {
  t.Fatalf("group text incorrectly counted as unique photo ownership evidence: %+v",got)
 }
 if len(got.Items)!=1||got.Items[0].Text!="Independent analysis."{t.Fatalf("editorial prose changed: %+v",got.Items)}
}

func TestInspectWrapperTextOutsideBoundedScopeNotCounted(t *testing.T){
 raw:=`<article><section><div><div><div><div><img src="/deep.jpg"></div></div></div></div><p>Separate article paragraph.</p></section></article>`
 got,err:=Inspect(raw,baseURL)
 if err!=nil {t.Fatal(err)}
 if got.UnwrappedImages!=1 || got.CompactImageTextWrappers!=0 || got.OtherUnwrappedImages!=1 {
  t.Fatalf("distant paragraph incorrectly associated with image: %+v",got)
 }
}

func TestInspectOtherUnwrappedImagesPartitionIntoBoundedGroups(t *testing.T) {
 raw:=`<article><div><div><img src="/left.jpg"></div><div><img src="/right.jpg"></div><div><span>Shared text, no per-photo caption proof.</span></div></div><div><img src="/standalone.jpg"></div><p>Independent article prose.</p></article>`
 got,err:=Inspect(raw,baseURL)
 if err!=nil {t.Fatal(err)}
 if got.UnwrappedImages!=3 || got.CompactImageTextWrappers!=0 || got.OtherUnwrappedImages!=3 || got.OtherInMultiImageGroups!=2 || got.OtherWithoutBoundedGroup!=1 {
  t.Fatalf("incorrect bounded group partition: %+v",got)
 }
 if len(got.Items)!=1 || got.Items[0].Text!="Independent article prose." {t.Fatalf("diagnostic changed prose: %+v",got.Items)}
}

// An unrelated section paragraph may satisfy the *weak* compact-wrapper test;
// do not treat compact evidence as caption ownership. This test checks only
// that unrelated images are not grouped at the entire-article scope.
func TestInspectMultiImageGroupingDoesNotUseArticleScope(t *testing.T) {
 raw:=`<article><p>A standalone lead.</p><div><img src="/one.jpg"></div><section><p>A separate section.</p><div><img src="/two.jpg"></div></section></article>`
 got,err:=Inspect(raw,baseURL)
 if err!=nil {t.Fatal(err)}
 if got.CompactImageTextWrappers!=1 || got.OtherUnwrappedImages!=1 || got.OtherInMultiImageGroups!=0 || got.OtherWithoutBoundedGroup!=1 {
  t.Fatalf("article-wide multi-image group was incorrectly inferred: %+v",got)
 }
 if len(got.Items)!=2 {t.Fatalf("ordinary prose lost: %+v",got.Items)}
}
