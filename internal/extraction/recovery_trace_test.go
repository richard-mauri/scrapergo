package extraction

import (
 "encoding/json"
 "reflect"
 "strings"
 "testing"

 "clean-article-reader/internal/article"
)

func TestRecoveryTraceDetectsInPlaceModification(t *testing.T) {
 trace:=NewRecoveryTrace()
 blocks:=[]article.Block{{Kind:"p",Text:"A paragraph",Parts:[]article.TextPart{{Text:"A paragraph"}}}}
 got:=trace.Run("selected-html-links",blocks,func() []article.Block {
  blocks[0].Parts[0].URL="https://example.org/link"
  return blocks
 })
 if got[0].Parts[0].URL==""{t.Fatal("stage did not run")}
 results:=trace.Results()
 for _,r:=range results {
  if r.Stage=="selected-html-links" && r.Outcome!="modified" {t.Fatalf("in-place mutation not detected: %+v",r)}
  if r.Stage=="legacy-fallback" && r.Outcome!="not_reached" {t.Fatalf("incorrect default outcome: %+v",r)}
 }
 data,_:=json.Marshal(results)
 if strings.Contains(string(data),"A paragraph") || strings.Contains(string(data),"example.org/link") {
  t.Fatal("trace must not contain content or link URLs")
 }
}

func TestRecoveryTraceLeavesArticleOutputIdentical(t *testing.T) {
 raw:=`<html><head><title>General Example</title></head><body><article><h1>General Example</h1><p>`+
 strings.Repeat("A complete article discusses independent engineering evidence and practical context for readers. ",14)+
 `</p><h2>Technical Details</h2><p>`+
 strings.Repeat("The second section explains experimental results, measurements, and their practical interpretation. ",14)+
 `</p><time>Updated recently</time><h2>Further Context</h2><p>`+
 strings.Repeat("Additional reporting supplies background, research details and independent perspectives for readers. ",14)+
 `</p></article></body></html>`
 baseline,err:=ExtractWithReadability(raw,"https://example.org/ordinary",LegacyExtract,InlineTextParts)
 if err!=nil{t.Fatal(err)}
 trace:=NewRecoveryTrace()
 traced,err:=ExtractWithReadabilityTraced(raw,"https://example.org/ordinary",LegacyExtract,InlineTextParts,trace)
 if err!=nil{t.Fatal(err)}
 if !reflect.DeepEqual(baseline,traced){t.Fatalf("traced article differs from normal output\nnormal: %#v\ntraced: %#v",baseline,traced)}
 if len(trace.Results())!=len(RecoveryStages){t.Fatal("missing trace stage")}
}
