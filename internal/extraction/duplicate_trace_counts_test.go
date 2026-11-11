package extraction

import (
 "encoding/json"
 "strings"
 "testing"

 "clean-article-reader/internal/article"
)

func TestDuplicateCandidatePairsDistinguishesTextRepetitionAndOverlap(t *testing.T) {
 paragraph:="Researchers followed the instruments carefully and described the same measurements from independent observations."
 other:="Another paragraph establishes background to the observation and explains its independent significance to readers."
 blocks:=[]article.Block{
  {Kind:"p",Text:paragraph},
  {Kind:"p",Text:other},
  {Kind:"p",Text:paragraph},
  {Kind:"img",Text:paragraph+" Photo credit: Regional Laboratory."},
  {Kind:"p",Text:"An intentional repeated heading."},
  {Kind:"p",Text:"An intentional repeated heading."},
 }
 exact,overlap:=duplicateCandidatePairs(blocks)
 if exact!=2 || overlap!=2 {t.Fatalf("expected two exact and two overlap pairs, got exact=%d overlap=%d",exact,overlap)}
}

func TestRecoveryStageDiagnosticCountsAppearOnlyWhenTraced(t *testing.T) {
 const textA="The observational results are discussed in detail for all readers of the scientific research article."
 blocks:=[]article.Block{{Kind:"p",Text:textA}}
 trace:=NewRecoveryTrace()
 trace.Observe("readability-render-text",blocks)
 blocks=trace.Run("editorial-continuity",blocks,func() []article.Block {
  return append(append([]article.Block(nil),blocks...),article.Block{Kind:"p",Text:textA})
 })
 trace.ObserveTransition("reader-postprocessing",blocks,blocks)
 results:=trace.Results()
 stage:=map[string]RecoveryStageResult{}
 for _,s:=range results {stage[s.Stage]=s}
 if stage["readability-render-text"].AfterExactPairs!=0 {t.Fatalf("baseline falsely counted duplicates: %+v",stage["readability-render-text"])}
 if stage["editorial-continuity"].BeforeExactPairs!=0 || stage["editorial-continuity"].AfterExactPairs!=1 {
  t.Fatalf("stage not attributed correctly: %+v",stage["editorial-continuity"])
 }
 if stage["reader-postprocessing"].AfterExactPairs!=1 {
  t.Fatalf("final duplicate missing: %+v",stage["reader-postprocessing"])
 }
 serial,err:=json.Marshal(results)
 if err!=nil {t.Fatal(err)}
 if strings.Contains(string(serial),textA) || strings.Contains(string(serial),"scientific research") {
  t.Fatalf("diagnostic leaked article text: %s",serial)
 }
 if got:=(*RecoveryTrace)(nil).Run("no-trace",blocks,func() []article.Block {return blocks});len(got)!=2 {t.Fatal("disabled tracing affected blocks")}
}
