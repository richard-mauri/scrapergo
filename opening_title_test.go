package main

import "testing"

func TestRemoveRepeatedOpeningTitle(t *testing.T) {
 a:=Article{Title:"Meet the people moving north to escape increasingly hot summers",Blocks:[]Block{{Kind:"h2",Text:"Meet the people moving north to escape increasingly hot summers"},{Kind:"p",Text:"A real paragraph remains."}}}
 removeRepeatedOpeningTitle(&a)
 if len(a.Blocks)!=1 || a.Blocks[0].Text!="A real paragraph remains." {t.Fatalf("duplicate opening title not removed: %+v",a.Blocks)}
}
func TestPreserveDistinctSectionHeading(t *testing.T) {
 a:=Article{Title:"An article",Blocks:[]Block{{Kind:"h2",Text:"Background"},{Kind:"p",Text:"A real paragraph."}}}
 removeRepeatedOpeningTitle(&a)
 if len(a.Blocks)!=2 {t.Fatalf("legitimate heading removed: %+v",a.Blocks)}
}
