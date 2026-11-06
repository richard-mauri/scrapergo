package extraction

import (
 "strings"
 "testing"

 "clean-article-reader/internal/article"
)

func TestPreserveReadabilityStructureTankerBoundary(t *testing.T) {
 cleaned := `<html><body><section>
 <p>Iran's Supreme Leader has not appeared publicly after the February strikes.</p>
 <time datetime="2026-11-06T07:30:00Z">7:30 AM</time>
 <h2>Tanker hit by explosion in Strait of Hormuz, U.K. naval agency says</h2>
 <p>A tanker was hit by an explosion in an important shipping lane on Thursday.</p>
 </section></body></html>`
 blocks:=[]article.Block{
  {Kind:"p",Text:"Iran's Supreme Leader has not appeared publicly after the February strikes. 7:30 AM"},
  {Kind:"p",Text:"Tanker hit by explosion in Strait of Hormuz, U.K. naval agency says"},
  {Kind:"p",Text:"A tanker was hit by an explosion in an important shipping lane on Thursday."},
 }
 got:=PreserveReadabilityStructure(blocks,cleaned)
 if len(got)!=4 {t.Fatalf("blocks: %+v",got)}
 if got[0].Kind!="p" || strings.Contains(got[0].Text,"7:30 AM") {t.Fatalf("preceding prose damaged: %+v",got[0])}
 if got[1].Kind!="time" || got[1].Text!="7:30 AM" {t.Fatalf("time not separated: %+v",got[1])}
 if got[2].Kind!="h2" || !strings.HasPrefix(got[2].Text,"Tanker hit") {t.Fatalf("heading not restored: %+v",got[2])}
 if got[3].Kind!="p" {t.Fatalf("body lost: %+v",got[3])}
}

func TestPreserveReadabilityStructureStandaloneTimestamp(t *testing.T) {
 html:=`<section><time datetime="2026-11-06T07:30:00Z">7:30 AM</time><h3>Another update</h3><p>Substantive article prose.</p></section>`
 got:=PreserveReadabilityStructure([]article.Block{{Kind:"p",Text:"7:30 AM"},{Kind:"p",Text:"Another update"},{Kind:"p",Text:"Substantive article prose."}},html)
 if got[0].Kind!="time" || got[1].Kind!="h3" || got[2].Kind!="p" {t.Fatalf("unexpected: %+v",got)}
}

func TestPreserveReadabilityStructureLinks(t *testing.T) {
 html:=`<section><p>Original source link.</p><time>7:30 AM</time><h2>Latest shipping report</h2></section>`
 input:=[]article.Block{
  {Kind:"p",Text:"Original source link. 7:30 AM",Parts:[]article.TextPart{{Text:"Original "},{Text:"source link",URL:"https://example.org/notice"},{Text:". 7:30 AM"}}},
  {Kind:"p",Text:"Latest shipping report"},
 }
 got:=PreserveReadabilityStructure(input,html)
 if len(got)!=3 || got[0].Text!="Original source link." || got[1].Kind!="time" || got[2].Kind!="h2" {t.Fatalf("unexpected: %+v",got)}
 if len(got[0].Parts)!=3 || got[0].Parts[1].URL!="https://example.org/notice" || strings.Contains(got[0].Parts[2].Text,"7:30") {t.Fatalf("inline link or timestamp not preserved: %+v",got[0].Parts)}
}

func TestPreserveReadabilityStructureNoUncorroboratedChange(t *testing.T) {
 input:=[]article.Block{{Kind:"p",Text:"Prose ends at 7:30 AM"},{Kind:"p",Text:"Not actually a heading"}}
 cases:=[]string{
  `<section><p>Not actually a heading</p></section>`,
  `<section><time>7:45 AM</time><h2>Not actually a heading</h2></section>`,
  `<section><time>7:30 AM</time><h2>Not actually a heading</h2><h2>Not actually a heading</h2></section>`,
 }
 for _,cleaned:=range cases {
  got:=PreserveReadabilityStructure(input,cleaned)
  if len(got)!=2 || got[0].Text!=input[0].Text || got[0].Kind!="p" || got[1].Kind!="p" {t.Fatalf("incorrectly inferred structure: %+v",got)}
 }
}

func TestPreserveReadabilityStructurePlainArticleHeading(t *testing.T) {
 got:=PreserveReadabilityStructure([]article.Block{{Kind:"p",Text:"Regular introduction"},{"Kind":"p",Text:"Section discussion"},{"Kind":"p",Text:"More discussion."}},`<article><p>Regular introduction</p><h2>Section discussion</h2><p>More discussion.</p></article>`)
 if len(got)!=3 || got[0].Kind!="p" || got[1].Kind!="h2" || got[2].Kind!="p" {t.Fatalf("ordinary heading: %+v",got)}
}

func TestReadabilityPipelineRestoresSemanticHeadings(t *testing.T) {
 intro:=strings.Repeat("Regional officials reviewed the situation with nearby governments in detail. ",5)
 body:=strings.Repeat("Maritime authorities received a report and are gathering additional information from crew members. ",5)
 raw:=`<!doctype html><html><head><title>Rolling coverage</title></head><body><main><h1>Rolling coverage</h1>
 <article><p>`+intro+`</p><time datetime="2026-11-06T07:30:00Z">7:30 AM</time>
 <h2>Tanker hit by explosion in Strait of Hormuz, U.K. naval agency says</h2>
 <p>`+body+`</p></article></main></body></html>`
 a,err:=ExtractWithReadability(raw,"https://example.org/coverage",LegacyExtract,InlineTextParts)
 if err!=nil{t.Fatal(err)}
 heading:=false
 for _,b:=range a.Blocks {
  if b.Kind=="h2" && strings.HasPrefix(b.Text,"Tanker hit by explosion") {heading=true}
 }
 if !heading {t.Fatalf("extraction flattened source-backed heading: %+v",a.Blocks)}
}
