package main

import (
 "bytes"
 "encoding/json"
 "strings"
 "testing"

 "clean-article-reader/internal/diagnostics"
 "clean-article-reader/internal/extraction"
)

func TestDuplicateProvenanceRecentWindowAndPrivacy(t *testing.T) {
 makeRow:=func(id string,after int) diagnostics.RecoveryRecord {
  return diagnostics.RecoveryRecord{
   ArticleID:id, Publisher:"example.org",
   Stages:[]extraction.RecoveryStageResult{
    {Stage:"readability-render-text",Outcome:"observed",AfterExactPairs:0},
    {Stage:"editorial-continuity",Outcome:"modified",BeforeExactPairs:0,AfterExactPairs:after},
    {Stage:"reader-postprocessing",Outcome:"observed",BeforeExactPairs:after,AfterExactPairs:after},
   },
  }
 }
 old:=diagnostics.RecoveryRecord{Stages:[]extraction.RecoveryStageResult{{Stage:"editorial-continuity",Outcome:"modified",AfterExactPairs:100}}}
 rows:=[]diagnostics.RecoveryRecord{old,makeRow("secret-first",1),makeRow("secret-second",2)}
 var input bytes.Buffer
 for _,row:=range rows {b,err:=json.Marshal(row);if err!=nil {t.Fatal(err)};input.Write(b);input.WriteByte('\n')}
 selected,err:=lastTraces(&input,1)
 if err!=nil || len(selected)!=1 || selected[0].ArticleID!="secret-second" {t.Fatalf("selected=%+v err=%v",selected,err)}
 stats:=compare(selected)
 if stats["editorial-continuity"].addedExact!=2 || stats["reader-postprocessing"].afterExact!=2 {t.Fatalf("unexpected aggregate=%+v",stats)}
 var report bytes.Buffer
 writeReport(&report,selected)
 if strings.Contains(report.String(),"secret-second") || strings.Contains(report.String(),"example.org") {t.Fatalf("private identifiers included in report: %s",report.String())}
 if !strings.Contains(report.String(),"editorial-continuity") {t.Fatal("missing stage report")}
}

func TestDuplicateProvenanceRejectsInvalidJSON(t *testing.T) {
 _,err:=lastTraces(strings.NewReader("{oops}\n"),10)
 if err==nil {t.Fatal("invalid trace record must be rejected")}
}
