package extraction

import (
	"fmt"
	"strings"
	"testing"
)

// This file measures the current *complete* Readability extractor on
// deterministic, publisher-neutral HTML. It is not a measure of network
// retrieval, server handlers, UI rendering, or Safari performance.
//
// Run: go test ./internal/extraction -run '^$' -bench '^BenchmarkReadabilityExtraction' -benchmem -benchtime=500ms -count=5
func benchmarkArticleHTML(withUpdates bool) string {
	var html strings.Builder
	html.WriteString(`<!doctype html><html><head><title>Regional report</title>`)
	html.WriteString(`<meta property="og:title" content="Regional report"></head><body><main><article>`)
	html.WriteString(`<h1>Regional report</h1>`)
	for i := 0; i < 18; i++ {
		if withUpdates && (i == 0 || i == 6 || i == 12) {
			fmt.Fprintf(&html, "<time datetime=\"2026-11-06T09:%02d:00Z\">%d:%02d AM</time>", i*2, 9+i/6, i*2)
			fmt.Fprintf(&html, "<h2>Regional update %d with verified headline</h2>", 1+i/6)
		} else if !withUpdates && (i == 6 || i == 12) {
			fmt.Fprintf(&html, "<h2>Editorial section %d</h2>", i/6)
		}
		fmt.Fprintf(&html, "<p>Field reporters covering district %d verified the published conditions with several independent officials and reviewed the available records for accuracy. The report describes current observations, historical context, and implications for nearby communities.</p>", i)
	}
	html.WriteString("</article></main></body></html>")
	return html.String()
}

func benchmarkExtractWithReadability(b *testing.B, withUpdates bool) {
	b.Helper()
	raw := benchmarkArticleHTML(withUpdates)
	const source = "https://example.org/coverage"
	// Verify setup outside the timed loop so fallback or truncated test inputs
	// cannot silently masquerade as the intended Readability benchmark.
	sample, err := ExtractWithReadability(raw, source, LegacyExtract, InlineTextParts)
	if err != nil {
		b.Fatalf("fixture extraction failed: %v", err)
	}
	if sample.Method != "Go Readability v2" || sample.WordCount < 60 || len(sample.Blocks) < 10 {
		b.Fatalf("fixture did not exercise normal Readability extraction: method=%q words=%d blocks=%d", sample.Method, sample.WordCount, len(sample.Blocks))
	}
	b.ReportAllocs()
	b.SetBytes(int64(len(raw)))
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := ExtractWithReadability(raw, source, LegacyExtract, InlineTextParts)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkReadabilityExtractionOrdinary(b *testing.B) {
	benchmarkExtractWithReadability(b, false)
}

func BenchmarkReadabilityExtractionMultiUpdate(b *testing.B) {
	benchmarkExtractWithReadability(b, true)
}
