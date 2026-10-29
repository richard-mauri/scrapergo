package pageclassify

import (
 "fmt"
 "strings"
 "testing"
)
func TestIndexEvidenceMatchesClassifier(t *testing.T) {
 var b strings.Builder
 b.WriteString("<html><main>")
 for i:=0;i<30;i++ {b.WriteString(fmt.Sprintf("<a href='/story/%d'><h3>Detailed market headline about company number %d</h3></a>",i,i))}
 b.WriteString("</main></html>")
 raw:=b.String()
 evidence:=IndexEvidence(raw)
 if evidence["index"]!=IsIndex(raw) {t.Fatalf("evidence %v disagrees with IsIndex",evidence)}
 if evidence["links"].(int)!=30 {t.Fatalf("unexpected counts: %v",evidence)}
}
