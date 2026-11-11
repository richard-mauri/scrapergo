package extraction

import (
 "strings"

 "clean-article-reader/internal/article"
 "clean-article-reader/internal/sourcehtml"
)

// duplicateCandidatePairs applies the same conservative screening criteria as
// NewsGo's standalone duplicate audit, without retaining text or identifiers.
// Counts are diagnostic only: an editorial repetition is not necessarily a bug.
// Invoked only when SCRAPERGO_RECOVERY_TRACE=1.
func duplicateCandidatePairs(blocks []article.Block) (exact, overlaps int) {
 texts:=make([]string,0,len(blocks))
 for _,b:=range blocks {
  if b.Kind=="table" || b.Kind=="topic-link" {continue}
  key:=strings.ToLower(sourcehtml.Normalize(b.Text))
  if len([]rune(key))<24 {continue}
  for _,earlier:=range texts {
   if key==earlier {exact++;continue}
   if len([]rune(key))<32 || len([]rune(earlier))<32 {continue}
   shorter,longer:=earlier,key
   if len([]rune(shorter))>len([]rune(longer)) {shorter,longer=longer,shorter}
   if len([]rune(shorter))>=30 && (strings.HasPrefix(longer,shorter+" ") || strings.HasSuffix(longer," "+shorter)) {overlaps++}
  }
  texts=append(texts,key)
 }
 return exact,overlaps
}
