package extraction

import (
 "testing"

 "clean-article-reader/internal/article"
 "clean-article-reader/internal/sourcehtml"
)

func TestImageCaptionCreditTripleSourceAndSelected(t *testing.T) {
 const caption="The keyboard and trackpad provide a familiar tactile experience."
 const credit="Photo credit: Sample Photographer"
 source:=`<article><div><div><img src="/keyboard.jpg"><span>`+caption+`</span><span>`+credit+`</span></div><div><span>`+caption+` `+credit+`</span></div></div><p>Independent editorial reporting follows.</p></article>`
 selected:=`<article><div><div><img src="/keyboard.jpg"><p>`+caption+` `+credit+`</p></div><p>`+caption+` `+credit+`</p></div><p>Independent editorial reporting follows.</p></article>`
 blocks:=[]article.Block{
  {Kind:"p",Text:"Independent editorial reporting follows."},
  {Kind:"p",Text:caption},
  {Kind:"p",Text:credit},
  {Kind:"p",Text:caption+" "+credit},
  {Kind:"p",Text:"Following editorial story paragraph."},
 }
 got:=reconcileImageCaptionCreditTriples(blocks,source,selected)
 if len(got)!=4 {t.Fatalf("expected one redundant combined block removed, got %+v",got)}
 if got[1].Text!=caption || got[2].Text!=credit || got[3].Text!=blocks[4].Text {
  t.Fatalf("caption, credit or legitimate prose missing: %+v",got)
 }
}

func TestImageCaptionCreditTripleDoesNotTouchRepeatedProse(t *testing.T) {
 const a="A genuinely repeated editorial passage that appears multiple times."
 const b="Different independent reporting follows"
 blocks:=[]article.Block{{Kind:"p",Text:a},{Kind:"p",Text:b},{Kind:"p",Text:a+" "+b}}
 source:=`<article><p>`+a+`</p><p>`+b+`</p><p>`+a+` `+b+`</p><img src="/unrelated.jpg"></article>`
 got:=reconcileImageCaptionCreditTriples(blocks,source,source)
 if len(got)!=3 {t.Fatalf("real repeated prose removed: %+v",got)}
}

func TestImageCaptionCreditTripleRequiresBothDOMRepresentations(t *testing.T) {
 const a="An illustration clearly identifies the hardware's side connectors."
 const b="Credit: Example Photographer"
 source:=`<article><div><img src="/photo.jpg"><div>`+a+` `+b+`</div></div></article>`
 selected:=`<article><p>`+a+`</p><p>`+b+`</p><p>`+a+` `+b+`</p></article>`
 blocks:=[]article.Block{{Kind:"p",Text:a},{Kind:"p",Text:b},{Kind:"p",Text:a+" "+b}}
 if got:=reconcileImageCaptionCreditTriples(blocks,source,selected);len(got)!=3 {
  t.Fatalf("source-only figure evidence should not delete: %+v",got)
 }
}

func TestImageCaptionCreditTriplePreservesAmbiguousGallery(t *testing.T) {
 const a="A clear photograph shows the two side ports and their arrangement."
 const b="Photographer: Example Contributor"
 shared:=`<article><div><img src="/a.jpg"><img src="/b.jpg"><span>`+a+` `+b+`</span></div></article>`
 blocks:=[]article.Block{{Kind:"p",Text:a},{Kind:"p",Text:b},{Kind:"p",Text:a+" "+b}}
 if got:=reconcileImageCaptionCreditTriples(blocks,shared,shared);len(got)!=3 {
  t.Fatalf("ambiguous two-photo gallery caused deletion: %+v",got)
 }
}

func TestImageCaptionCreditTripleKeepsDistinctPhotographs(t *testing.T) {
 const caption="Two different images legitimately use this same descriptive caption."
 const credit="Photo by Example"
 raw:=`<article><div><img src="/a.jpg"><span>`+caption+` `+credit+`</span></div><div><img src="/b.jpg"><span>`+caption+` `+credit+`</span></div></article>`
 blocks:=[]article.Block{{Kind:"p",Text:caption},{Kind:"p",Text:credit},{Kind:"p",Text:caption+" "+credit},
  {Kind:"p",Text:caption},{Kind:"p",Text:credit},{Kind:"p",Text:caption+" "+credit}}
 // An image-aligned representation does not establish that a combined
 // block is a redundant alternative. The unique image mapping still needs
 // to be proved before applying this transformation across galleries.
 got:=reconcileImageCaptionCreditTriples(blocks,raw,raw)
 if len(got)!=6 {t.Fatalf("distinct photo captions must not be collapsed: %+v",got)}
}

func TestImageCaptionCreditTriplePreservesLinkedCombinedBlock(t *testing.T) {
 const a="A detailed photo caption containing some linked source material."
 const b="Photographer Credit"
 raw:=`<article><div><img src="/photo.jpg"><span>`+a+` `+b+`</span></div></article>`
 blocks:=[]article.Block{{Kind:"p",Text:a},{Kind:"p",Text:b},{Kind:"p",Text:a+" "+b,Links:[]article.InlineLink{{Text:"source",URL:"https://example.org"}}}}
 if got:=reconcileImageCaptionCreditTriples(blocks,raw,raw);len(got)!=3 {
  t.Fatalf("rich combined caption cannot be discarded: %+v",got)
 }
}

func TestImageCaptionCreditTripleDoesNotRemoveIndependentSourceParagraph(t *testing.T) {
 const caption="A landscape photograph illustrates the extensive restoration work."
 const credit="Credit: Conservation Team"
 // There is one image/caption, but the separate editorial paragraph has
 // intentionally repeated the full wording. This is not an alternative
 // non-paragraph photo caption representation in the original source.
 raw:=`<article><div><div><img src="/landscape.jpg"><span>`+caption+` `+credit+`</span></div><p>`+caption+` `+credit+`</p></div></article>`
 blocks:=[]article.Block{{Kind:"p",Text:caption},{Kind:"p",Text:credit},{Kind:"p",Text:caption+" "+credit}}
 if got:=reconcileImageCaptionCreditTriples(blocks,raw,raw);len(got)!=3 {
  t.Fatalf("independent source editorial paragraph deleted: %+v",got)
 }
}

func TestImageCaptionCreditTripleNoSourceParsingNeededWithoutCandidate(t *testing.T) {
 blocks:=[]article.Block{{Kind:"p",Text:"The report opens with independent research."},{Kind:"p",Text:"Readers can evaluate the evidence."}}
 // Invalid HTML has no bearing on unrelated ordinary prose.
 if got:=reconcileImageCaptionCreditTriples(blocks,"<<bad>>","<<bad>>");len(got)!=len(blocks) {
  t.Fatalf("ordinary content was changed: %+v",got)
 }
}

func TestImageCaptionCombinedRepresentationsDropOnlyWeakerCopy(t *testing.T) {
 const caption="An image shows the two connectors and the compact enclosure."
 const credit="Photo credit: Field Reporter"
 source:=`<article><div><div><img src="/ports.jpg"><span>`+caption+`</span><span>`+credit+`</span></div><div><span>`+caption+` `+credit+`</span></div></div></article>`
 selected:=`<article><div><div><img src="/ports.jpg"><p>`+caption+credit+`</p></div><p>`+caption+` `+credit+`</p></div></article>`
 blocks:=[]article.Block{{Kind:"p",Text:caption+credit},{Kind:"p",Text:caption+" "+credit}}
 got:=reconcileImageCaptionCreditTriples(blocks,source,selected)
 if len(got)!=1 || got[0].Text!=blocks[1].Text {
  t.Fatalf("failed to retain correctly spaced combined caption: %+v",got)
 }
}

func TestImageCaptionCombinedRepresentationsKeepDistinctPhotoCopies(t *testing.T) {
 const caption="The same detailed description legitimately belongs to both images."
 const credit="Photo credit: Field Reporter"
 source:=`<article><div><img src="/first.jpg"><span>`+caption+` `+credit+`</span></div><div><img src="/second.jpg"><span>`+caption+` `+credit+`</span></div></article>`
 blocks:=[]article.Block{{Kind:"p",Text:caption+" "+credit},{Kind:"p",Text:caption+" "+credit}}
 got:=reconcileImageCaptionCreditTriples(blocks,source,source)
 if len(got)!=2 {t.Fatalf("two distinct photographs lost independent captions: %+v",got)}
}

func TestImageCaptionCombinedRepresentationsKeepPlainProse(t *testing.T) {
 const text="This report includes a deliberately repeated quotation about the research."
 source:=`<article><p>`+text+`</p><p>`+text+`</p><img src="/unrelated.jpg"></article>`
 blocks:=[]article.Block{{Kind:"p",Text:text},{Kind:"p",Text:text}}
 got:=reconcileImageCaptionCreditTriples(blocks,source,source)
 if len(got)!=2 {t.Fatalf("plain repeated prose was deleted: %+v",got)}
}

func TestCaptionCreditGateEvidenceMatchesProductionPredicate(t *testing.T) {
 const caption="An image shows the redesigned keyboard and trackpad in clear daylight."
 const credit="Photo credit: Example Reporter"
 source:=`<article><div><div><img src="/keyboard.jpg"><span>`+caption+`</span><span>`+credit+`</span></div><div><span>`+caption+` `+credit+`</span></div></div></article>`
 selected:=`<article><div><div><img src="/keyboard.jpg"><p>`+caption+` `+credit+`</p></div><p>`+caption+` `+credit+`</p></div></article>`
 blocks:=[]article.Block{{Kind:"p",Text:caption},{Kind:"p",Text:credit},{Kind:"p",Text:caption+" "+credit}}
 good:=InspectCaptionCreditTriple(blocks,2,source,selected)
 if !good.TripleShape || !good.SourceImageScope || !good.SelectedImageScope || !good.Eligible {
  t.Fatalf("eligible source-backed triple not reported: %+v",good)
 }
 bad:=InspectCaptionCreditTriple(blocks,2,source,`<article><p>`+caption+` `+credit+`</p></article>`)
 if !bad.TripleShape || !bad.SourceImageScope || bad.SelectedImageScope || bad.Eligible {
  t.Fatalf("selected HTML guard should prevent removal: %+v",bad)
 }
 if got:=InspectCaptionCreditTriple(blocks,0,source,selected);got.Eligible || got.TripleShape {
  t.Fatalf("invalid block index reported eligible: %+v",got)
 }
}

func TestImageCaptionCreditFlattenedReadabilityPairedPhotoGroup(t *testing.T) {
 const caption="The keyboard and trackpad remain excellent for the intended workload."
 const credit="Example Photographer"
 // Source has both image-associated caption and an alternative NON-p
 // representation in the same unique image scope.
 source:=`<article><div><div><img src="/keyboard.jpg"><span>`+caption+`</span><span>`+credit+`</span></div><div><span>`+caption+` `+credit+`</span></div></div></article>`
 // Readability preserves the image/caption, but puts the duplicate p in
 // a neighboring branch of a wider two-photo group, not a one-image wrapper.
 selected:=`<article><div><div><img src="/other.jpg"><span>Different caption</span></div><div><div><img src="/keyboard.jpg"><p>`+caption+` `+credit+`</p></div></div><div><p><span>`+caption+` `+credit+`</span></p></div></div></article>`
 blocks:=[]article.Block{{Kind:"p",Text:caption},{Kind:"p",Text:credit},{Kind:"p",Text:caption+" "+credit}}
 if hasImageCaptionCreditScope(sourcehtml.Parse(source),caption,credit,false)!=true {t.Fatal("source fixture missing unique image scope")}
 if hasImageCaptionCreditScope(sourcehtml.Parse(selected),caption,credit,true) {t.Fatal("selected fixture unexpectedly satisfies strict guard")}
 got:=reconcileImageCaptionCreditTriples(blocks,source,selected)
 if len(got)!=2 || got[0].Text!=caption || got[1].Text!=credit {t.Fatalf("failed to remove only redundant combined text: %+v",got)}
 e:=InspectCaptionCreditTriple(blocks,2,source,selected)
 if !e.SelectedImageScope || !e.SourceImageScope || !e.Eligible {t.Fatalf("production/diagnostic disagree: %+v",e)}
}

func TestFlattenedReadabilityTwoDistinctImagesSameCaptionNotReconciled(t *testing.T) {
 const caption="The same caption is intentionally attached to two different photographs."
 const credit="Example Photographer"
 // Two distinct images render the same caption; there is no unique owner.
 source:=`<article><div><div><img src="/one.jpg"><span>`+caption+`</span><span>`+credit+`</span></div><div><span>`+caption+` `+credit+`</span></div></div><div><div><img src="/two.jpg"><span>`+caption+`</span><span>`+credit+`</span></div><div><span>`+caption+` `+credit+`</span></div></div></article>`
 selected:=`<article><div><div><img src="/one.jpg"><p>`+caption+` `+credit+`</p></div><div><img src="/two.jpg"><p>`+caption+` `+credit+`</p></div><div><p>`+caption+` `+credit+`</p></div></div></article>`
 blocks:=[]article.Block{{Kind:"p",Text:caption},{Kind:"p",Text:credit},{Kind:"p",Text:caption+" "+credit}}
 if got:=reconcileImageCaptionCreditTriples(blocks,source,selected);len(got)!=3 {t.Fatalf("collapsed two distinct image captions: %+v",got)}
}

func TestFlattenedReadabilityNoImageRetainsEditorialRepetition(t *testing.T) {
 const caption="An article paragraph is deliberately repeated for editorial emphasis."
 const credit="Example Photographer"
 source:=`<article><div><div><img src="/one.jpg"><span>`+caption+`</span><span>`+credit+`</span></div><div><span>`+caption+` `+credit+`</span></div></div></article>`
 selected:=`<article><p>`+caption+` `+credit+`</p><p>`+caption+` `+credit+`</p></article>`
 blocks:=[]article.Block{{Kind:"p",Text:caption},{Kind:"p",Text:credit},{Kind:"p",Text:caption+" "+credit}}
 if got:=reconcileImageCaptionCreditTriples(blocks,source,selected);len(got)!=3 {t.Fatalf("text-only Readability output must not authorize removal: %+v",got)}
}
