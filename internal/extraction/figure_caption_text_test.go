package extraction

import (
    "testing"

    "clean-article-reader/internal/article"
)

func TestRemoveFlattenedFigureCaptions(t *testing.T) {
    const html = `<article>
      <p>Opening editorial paragraph unrelated to the following media.</p>
      <figure><img src="/one.jpg"><figcaption><span>Right-side ports and connectors.</span><span>Photo credit: A. Reporter</span></figcaption></figure>
      <figure><img src="/two.jpg"><figcaption><span>Left-side ports and connectors.</span><span>Photo credit: A. Reporter</span></figcaption></figure>
      <p>Closing editorial paragraph unrelated to the photographs.</p>
    </article>`
    blocks := []article.Block{
        {Kind:"p",Text:"Opening editorial paragraph unrelated to the following media."},
        {Kind:"p",Text:"Right-side ports and connectors."},
        {Kind:"p",Text:"Photo credit: A. Reporter"},
        {Kind:"p",Text:"Right-side ports and connectors. Photo credit: A. Reporter"},
        {Kind:"p",Text:"Left-side ports and connectors."},
        {Kind:"p",Text:"Photo credit: A. Reporter"},
        {Kind:"p",Text:"Left-side ports and connectors. Photo credit: A. Reporter"},
        {Kind:"p",Text:"Closing editorial paragraph unrelated to the photographs."},
    }
    got := removeFlattenedFigureCaptions(blocks,html)
    if len(got)!=2 || got[0].Text!=blocks[0].Text || got[1].Text!=blocks[len(blocks)-1].Text {
        t.Fatalf("flattened caption copies remain or article prose lost: %+v",got)
    }
}

func TestRemoveFlattenedFigureCaptionsPreservesIndependentProse(t *testing.T) {
    const html = `<article>
      <p>View from the riverbank.</p>
      <figure><img src="/river.jpg"><figcaption>View from the riverbank.</figcaption></figure>
      <p>More evidence from the scene.</p>
    </article>`
    blocks := []article.Block{
        {Kind:"p",Text:"View from the riverbank."},
        {Kind:"p",Text:"More evidence from the scene."},
    }
    got:=removeFlattenedFigureCaptions(blocks,html)
    if len(got)!=len(blocks) {t.Fatalf("independent prose was removed: %+v",got)}
}

func TestRemoveFlattenedFigureCaptionsNeedsImageAndCaption(t *testing.T) {
    const html = `<article>
      <figcaption>Not associated with an image.</figcaption>
      <figure><img src="/photo.jpg"></figure>
      <p>Not associated with an image.</p>
    </article>`
    blocks:=[]article.Block{{Kind:"p",Text:"Not associated with an image."}}
    got:=removeFlattenedFigureCaptions(blocks,html)
    if len(got)!=1 {t.Fatalf("unrelated text was removed: %+v",got)}
}

func TestRemoveFlattenedFigureCaptionsKeepsRealProseRepetition(t *testing.T) {
    const html = `<article><p>The same answer was given twice.</p><p>The same answer was given twice.</p></article>`
    blocks:=[]article.Block{{Kind:"p",Text:"The same answer was given twice."},{Kind:"p",Text:"The same answer was given twice."}}
    got:=removeFlattenedFigureCaptions(blocks,html)
    if len(got)!=2 {t.Fatalf("legitimate repeated prose removed: %+v",got)}
}
