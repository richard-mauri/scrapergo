package extraction

import (
    "strings"
    "testing"
)

func TestReadabilityArticleDoesNotRepeatFigureCaptionInProse(t *testing.T) {
    const caption = "A coastal monitoring station during an evening survey."
    const credit = "Photo credit: Field Reporter"
    var raw strings.Builder
    raw.WriteString(`<!doctype html><html><head><title>Coastal monitoring report</title></head><body><main><article><h1>Coastal monitoring report</h1>`)
    for i:=0;i<5;i++ {
        raw.WriteString(`<p>Survey teams inspected the regional monitoring equipment and compared the readings with their historical records. They documented the current conditions and verified their observations with independent sources before publishing this report.</p>`)
    }
    raw.WriteString(`<figure><img src="https://images.example.org/station.jpg" alt="Monitoring station"><figcaption><span>`+caption+`</span><span>`+credit+`</span></figcaption></figure>`)
    for i:=0;i<5;i++ {
        raw.WriteString(`<p>The results help readers understand the changing conditions along the coastline. Researchers will continue evaluating the instruments and collecting additional information over the next several months.</p>`)
    }
    raw.WriteString(`</article></main></body></html>`)
    out,err:=ExtractWithReadability(raw.String(),"https://example.org/story",LegacyExtract,InlineTextParts)
    if err!=nil {t.Fatal(err)}
    if out.Method!="Go Readability v2" {t.Fatalf("unexpected fallback: %q",out.Method)}
    foundImage:=false
    for _,block:=range out.Blocks {
        if block.Kind=="img" && strings.Contains(block.URL,"station.jpg") {foundImage=true}
        if block.Kind=="p" && (block.Text==caption || block.Text==credit || block.Text==caption+" "+credit) {
            t.Fatalf("figure caption leaked as article prose: %+v",block)
        }
    }
    if !foundImage {t.Fatalf("editorial image lost: %+v",out.Blocks)}
}
