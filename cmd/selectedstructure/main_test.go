package main

import (
 "testing"
 "clean-article-reader/internal/extraction"
)

func TestImageStagesRetainsOnlyReachedExistingTraceAndImageCounts(t *testing.T){
 in:=[]extraction.RecoveryStageResult{
  {Stage:"readability-render-text",Outcome:"observed",BeforeImages:0,AfterImages:0},
  {Stage:"position-figures",Outcome:"modified",BeforeImages:0,AfterImages:4},
  {Stage:"reader-postprocessing",Outcome:"observed",BeforeImages:4,AfterImages:3},
  {Stage:"partial-html-fallback",Outcome:"not_reached"},
 }
 got:=imageStages(in)
 if len(got)!=3 {t.Fatalf("expected three reached stages, got %+v",got)}
 if got[1].Stage!="position-figures"||got[1].BeforeImages!=0||got[1].AfterImages!=4 {t.Fatalf("stage image counts changed: %+v",got[1])}
 if got[2].Stage!="reader-postprocessing"||got[2].BeforeImages!=4||got[2].AfterImages!=3 {t.Fatalf("reader count changed: %+v",got[2])}
}
