package main

import "testing"

func TestReadabilityNPRPlayerFragmentsSanitized(t *testing.T) {
 a:=Article{Blocks:[]Block{
  {Kind:"h2",Text:"Meet the people moving north to escape increasingly hot summers"},
  {Kind:"p",Text:"A normal introductory paragraph with article details."},
  {Kind:"h2",Text:"Embed"},
  {Kind:"p",Text:"•"},
  {Kind:"p",Text:"\">"},
  {Kind:"blockquote",Text:`<iframe src="https://www.npr.org/player/embed/nx-s1-5977837/nx-s1-9973214" width="100%" height="290" frameborder="0" scrolling="no" title="NPR embedded audio player">`},
  {Kind:"p",Text:"The next legitimate paragraph remains visible."},
  {Kind:"img",URL:"https://example.org/photo.jpg",Text:"An article photograph"},
 }}
 sanitizeEmbeddedPlayerBlocks(&a)
 if len(a.Blocks)!=4 { t.Fatalf("unexpected remaining blocks: %+v",a.Blocks)}
 if a.Blocks[0].Kind!="h2" || a.Blocks[1].Kind!="p" || a.Blocks[2].Text!="The next legitimate paragraph remains visible." || a.Blocks[3].Kind!="img" {t.Fatalf("legitimate content changed: %+v",a.Blocks)}
}
func TestReadabilityNonEmbedTextPreserved(t *testing.T) {
 a:=Article{Blocks:[]Block{{Kind:"h2",Text:"Embed"},{Kind:"p",Text:"The article discusses audio reporting but contains no player markup."}}}
 sanitizeEmbeddedPlayerBlocks(&a)
 if len(a.Blocks)!=2 {t.Fatalf("removed legitimate content: %+v",a.Blocks)}
}
