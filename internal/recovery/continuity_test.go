package recovery

import (
 "testing"
 "clean-article-reader/internal/article"
)

func TestOrderedMatchesRequiresSequence(t *testing.T) {
 retained:=[]string{"The first editorial paragraph includes sufficient context to establish an exact source anchor and is long enough.","The second editorial paragraph provides additional context to establish another exact source anchor."}
 correct:=[]article.Block{{Kind:"p",Text:retained[0]},{Kind:"img",URL:"https://example.org/photo.jpg"},{Kind:"p",Text:retained[1]}}
 if got:=OrderedMatches(retained,correct);got!=2 {t.Fatalf("ordered matches %d",got)}
 reversed:=[]article.Block{{Kind:"p",Text:retained[1]},{Kind:"p",Text:retained[0]}}
 if got:=OrderedMatches(retained,reversed);got!=1 {t.Fatalf("reversed matches %d",got)}
}

func TestReplacementAllowedGuardsCompleteness(t *testing.T) {
 old:=[]article.Block{{Kind:"p",Text:"Original editorial paragraph has many words establishing the complete selected article and several independent contextual statements."}}
 seenText:=map[string]bool{Normalize(old[0].Text):true}
 candidates:=[]article.Block{{Kind:"p",Text:old[0].Text},{Kind:"p",Text:"A newly sourced paragraph provides additional details."},{Kind:"p",Text:"Another newly sourced paragraph completes the report."}}
 if !ReplacementAllowed(old,candidates,seenText,map[string]bool{}) {t.Fatal("two new paragraphs should allow replacement")}
 if ReplacementAllowed(old,old,seenText,map[string]bool{}) {t.Fatal("identical source should not replace")}
 tooShort:=[]article.Block{{Kind:"img",URL:"https://example.org/extra.jpg"}}
 if ReplacementAllowed(old,tooShort,seenText,map[string]bool{}) {t.Fatal("substantial prose loss should reject replacement")}
}
