package main

import (
 "strings"
 "testing"
)

func TestSourceProvenancePreservesOrderAndAncestorPaths(t *testing.T) {
 const text="Repeated editorial wording with sufficient substance to compare reliably in a reader."
 raw:="<html><body><article><section><p>"+text+"</p><p>"+text+"</p></section></article></body></html>"
 got:=scanProvenance(raw,text)
 if len(got)!=2 {t.Fatalf("expected two independent source paragraphs, got %+v",got)}
 if got[0].Path!="root/html[1]/body[1]/article[1]/section[1]/p[1]" ||
  got[1].Path!="root/html[1]/body[1]/article[1]/section[1]/p[2]" ||
  got[0].Order>=got[1].Order {t.Fatalf("incorrect source DOM provenance %+v",got)}
 for _,g:=range got {
  if g.Parent!="section" || g.Container!="section" || g.Origin!="text" || g.Mode!="text" {
   t.Fatalf("incorrect ancestry %+v",g)
  }
 }
}

func TestSourceProvenanceDistinguishesOneContainerFromTwoSelectedParagraphs(t *testing.T) {
 const passage="The photography shows a compact computer with three connectors and a redesigned enclosure."
 source:="<article><figure><div><span>The photography shows a compact computer</span><span> with three connectors and a redesigned enclosure.</span></div></figure></article>"
 selected:="<article><p>"+passage+"</p><p>"+passage+"</p></article>"
 first:=scanProvenance(source,passage)
 second:=scanProvenance(selected,passage)
 if len(first)!=1 || first[0].Tag!="div" || first[0].Container!="figure" {
  t.Fatalf("source should have one matching figure container: %+v",first)
 }
 if len(second)!=2 || second[0].Order>=second[1].Order || second[0].Tag!="p" {
  t.Fatalf("selected HTML should have two distinct matching p elements: %+v",second)
 }
}

func TestSourceProvenanceCaptionCreditAndAltAreDistinct(t *testing.T) {
 const caption="A close-up photograph shows the ports along the edge of the device."
 raw:="<article><figure><img src='/image.jpg' alt='"+caption+"'><figcaption><span>"+caption+"</span><span>Credit: Reporter</span></figcaption></figure></article>"
 got:=scanProvenance(raw,caption)
 if len(got)!=2 {t.Fatalf("caption text and alt should each have its own provenance: %+v",got)}
 if got[0].Tag!="img" || got[0].Origin!="attr:alt" || got[1].Tag!="span" ||
  got[1].Container!="figcaption" || got[1].Origin!="text" {
  t.Fatalf("lost caption vs alt distinction: %+v",got)
 }
}

func TestSourceProvenanceCompactInlineMarkupAndPrivacy(t *testing.T) {
 const passage="Readability may split a sentence across emphasized fragments while retaining the entire source text."
 raw:="<article><div>Readability may split a sentence across <em>emphasized</em> fragments while retaining the entire source text.</div></article>"
 got:=scanProvenance(raw,passage)
 if len(got)!=1 || got[0].Tag!="div" || got[0].Mode!="text" {t.Fatalf("inline text not recognized: %+v",got)}
 for _,g:=range got {
  if strings.Contains(g.Path,"Readability") || strings.Contains(g.Parent,"Readability") || strings.Contains(g.Origin,passage) {
   t.Fatalf("text escaped into public provenance: %+v",g)
  }
 }
}

func TestSourceProvenanceIgnoresBlockedTagsAndLimitsIncidentalMetadata(t *testing.T) {
 const sentence="These private source tokens must not be reported as actual editorial article content."
 raw:="<main><script>"+sentence+"</script><nav><p>"+sentence+"</p></nav><article><p>Ordinary valid reporting.</p></article></main>"
 if got:=scanProvenance(raw,sentence);len(got)!=0 {t.Fatalf("noise was incorrectly counted: %+v",got)}
}


func TestSourceProvenanceReportsSiblingMediaAndExplicitHiddenOnly(t *testing.T) {
 const passage="A carefully described image caption that should remain attached to its corresponding photograph."
 raw:="<article><div><img src='/pic.jpg'><span>"+passage+"</span><span aria-hidden='true'>"+passage+"</span></div></article>"
 got:=scanProvenance(raw,passage)
 if len(got)!=2 {t.Fatalf("expected two independent caption representations: %+v",got)}
 for i,n:=range got {
  if n.Media!="parent-img:1,figure:0,figcaption:0,siblings:3" {
   t.Fatalf("wrong media context for %d: %+v",i,n)
  }
 }
 if got[0].Visibility!="no-explicit-hint" || got[1].Visibility!="aria-hidden" {
  t.Fatalf("incorrect visibility provenance: %+v",got)
 }
}

func TestSourceProvenanceInlineHiddenAncestor(t *testing.T) {
 const passage="A detailed description of an image that appears in a hidden alternative block."
 raw:="<article><div hidden><span>"+passage+"</span></div></article>"
 got:=scanProvenance(raw,passage)
 if len(got)!=1 || got[0].Visibility!="hidden-attr" {t.Fatalf("hidden ancestor lost: %+v",got)}
}

func TestSourceProvenanceWithoutExplicitHintDoesNotClaimVisible(t *testing.T) {
 const passage="A second caption is hidden through a responsive stylesheet not present inline."
 raw:="<article><div class='hidden-by-external-css'><span>"+passage+"</span></div></article>"
 got:=scanProvenance(raw,passage)
 if len(got)!=1 || got[0].Visibility!="no-explicit-hint" {t.Fatalf("unknown CSS incorrectly inferred: %+v",got)}
}
