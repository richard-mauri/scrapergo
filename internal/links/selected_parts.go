package links

import "clean-article-reader/internal/article"

// SelectedParts represents inline markup recovered from one source paragraph
// already selected for the extracted article.
type SelectedParts struct {
 Text string
 Parts []article.TextPart
}

// AssignSelectedParts matches source inline markup to canonical selected blocks.
// Each target is assigned at most once, preserving existing source order.
func AssignSelectedParts(blocks []article.Block, evidence []SelectedParts) {
 used:=make(map[int]bool)
 for _,item:=range evidence {
  key:=SemanticKey(item.Text)
  if key=="" {continue}
  for i:=range blocks {
   if used[i]||blocks[i].Kind=="img"||blocks[i].Kind=="topic-link" {continue}
   if SemanticKey(blocks[i].Text)!=key {continue}
   used[i]=true
   blocks[i].Parts=item.Parts
   blocks[i].Links=nil
   break
  }
 }
}
