package main

import "testing"

func TestMinimalSourceMatchesFindsInlineAndNonParagraphSource(t *testing.T) {
 const target="The display provides higher contrast and still supports professional workflows."
 raw:=`<article><div class="copy">The display provides <strong>higher contrast</strong> and still supports professional workflows.</div><section><div>`+target+`</div></section></article>`
 found:=minimalSourceMatches(raw,target)
 if found["div"]!=2 || len(found)!=1 {t.Fatalf("expected two independent source divs, got %+v",found)}
 if got:=formatSourceTags(found);got!="div:2" {t.Fatalf("unexpected stable summary %q",got)}
}
func TestMinimalSourceMatchesPreservesOneCaptionAndOneParagraph(t *testing.T) {
 const target="Independent photography documented the equipment in its original configuration."
 raw:=`<article><p>`+target+`</p><figure><img src="/photo.jpg"><figcaption>`+target+`</figcaption></figure></article>`
 found:=minimalSourceMatches(raw,target)
 if found["p"]!=1 || found["figcaption"]!=1 || len(found)!=2 {t.Fatalf("unexpected source provenance: %+v",found)}
}
func TestMinimalSourceMatchesIgnoresScriptCopies(t *testing.T) {
 const target="An editorial description appears only inside an inaccessible script data structure."
 raw:=`<article><p>Actual reporting text.</p><script type="application/json">"`+target+`"</script></article>`
 got:=minimalSourceMatches(raw,target)
 if len(got)!=0 {t.Fatalf("script text incorrectly counted: %+v",got)}
}
