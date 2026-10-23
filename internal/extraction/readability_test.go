package extraction

import (

	"strings"
	"testing"
)

func TestReadabilityArticle(t *testing.T) {
	raw := `<html><head><title>Sample title</title></head><body><article><h1>Sample title</h1><p>` + strings.Repeat("This is a published paragraph about scientific research and its implications for readers. ", 15) + `</p><p>` + strings.Repeat("A second substantial paragraph explains the work and provides context to the public. ", 15) + `</p></article></body></html>`
	a, err := ExtractWithReadability(raw, "https://example.org/story",LegacyExtract,InlineTextParts)
	if err != nil {
		t.Fatal(err)
	}
	if a.WordCount < 60 {
		t.Fatalf("insufficient content: %+v", a)
	}
}
