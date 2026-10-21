package main

import "testing"

func TestNPRAudioEmbedMarkupIsNotArticleText(t *testing.T) {
 raw := `<article>
 <h2>Meet the people moving north to escape increasingly hot summers</h2>
 <p>This article explains migration toward cooler places.</p>
 <h2>Embed</h2>
 <p>&lt;iframe src="https://www.npr.org/player/embed/nx-s1-5977837/nx-s1-9973214" width="100%" height="290" frameborder="0" scrolling="no" title="NPR embedded audio player"&gt;</p>
 <p>Climate scientists are tracking the impacts of hotter summers.</p>
 </article>`
 blocks := extractBlocks(parseHTML(raw))
 for _, b := range blocks {
  if b.Text == "Embed" || isEmbeddedMarkupBlock(b.Text) {
   t.Fatalf("embedded player markup leaked: %q", b.Text)
  }
 }
 found := false
 for _, b := range blocks {if b.Text == "Climate scientists are tracking the impacts of hotter summers." {found=true}}
 if !found {t.Fatal("legitimate article text was lost")}
}
func TestEmbeddedMarkupDetection(t *testing.T) {
 for _, s := range []string{`<iframe src="https://www.npr.org/player/embed/id">`, `iframe src="https://www.npr.org/player/embed/id" width="100%"`, `frameborder="0" scrolling="no"`} {
  if !isEmbeddedMarkupBlock(s) {t.Errorf("did not reject embed markup %q",s)}
 }
 for _, s := range []string{"The article describes an embedded audio interview.", "A normal reporting paragraph."} {
  if isEmbeddedMarkupBlock(s) {t.Errorf("rejected article prose %q",s)}
 }
}
