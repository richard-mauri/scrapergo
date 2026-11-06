package extraction

import (
	"fmt"
	"strings"
	"testing"

	"clean-article-reader/internal/article"
)

// These are publisher-neutral contracts for the existing RenderHTML-backed
// overlay. They deliberately do not add special cases to production behavior.
func TestReadabilityStructureMultipleUpdatesInOrder(t *testing.T) {
	cleaned := `<main><article>
		<p>First dispatch remains editorial prose.</p>
		<time datetime="2026-11-06T08:00:00Z">8:00 AM</time>
		<h2>First verified update</h2><p>First report details.</p>
		<time datetime="2026-11-06T08:30:00Z">8:30 AM</time>
		<h3>Second verified update</h3><p>Second report details.</p>
	</article></main>`
	input := []article.Block{
		{Kind: "p", Text: "First dispatch remains editorial prose. 8:00 AM"},
		{Kind: "p", Text: "First verified update"},
		{Kind: "p", Text: "First report details. 8:30 AM"},
		{Kind: "p", Text: "Second verified update"},
		{Kind: "p", Text: "Second report details."},
	}
	got := PreserveReadabilityStructure(input, cleaned)
	wantKinds := []string{"p", "time", "h2", "p", "time", "h3", "p"}
	wantTexts := []string{
		"First dispatch remains editorial prose.", "8:00 AM",
		"First verified update", "First report details.", "8:30 AM",
		"Second verified update", "Second report details.",
	}
	if len(got) != len(wantKinds) {
		t.Fatalf("got %d blocks, want %d: %+v", len(got), len(wantKinds), got)
	}
	for i := range wantKinds {
		if got[i].Kind != wantKinds[i] || got[i].Text != wantTexts[i] {
			t.Fatalf("block %d: got %s %q; want %s %q", i, got[i].Kind, got[i].Text, wantKinds[i], wantTexts[i])
		}
	}
}

func TestReadabilityStructurePreservesUncorroboratedText(t *testing.T) {
	cases := []struct {
		name string
		html string
	}{
		{
			name: "ordinary text with no heading evidence",
			html: `<article><p>8:15 AM</p><p>Market conditions</p></article>`,
		},
		{
			name: "duplicate source headings are ambiguous",
			html: `<article><h2>Market conditions</h2><p>Other information.</p><h2>Market conditions</h2></article>`,
		},
		{
			name: "missing time node cannot justify splitting time-like prose",
			html: `<article><p>Forecast ends at 8:15 AM.</p><h2>Market conditions</h2></article>`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := []article.Block{
				{Kind: "p", Text: "Forecast ends at 8:15 AM."},
				{Kind: "p", Text: "Market conditions"},
			}
			got := PreserveReadabilityStructure(input, tc.html)
			if len(got) != 2 || got[0].Kind != "p" || got[0].Text != input[0].Text {
				t.Fatalf("unsupported timestamp boundary changed: %+v", got)
			}
			if tc.name != "missing time node cannot justify splitting time-like prose" &&
				(got[1].Kind != "p" || got[1].Text != "Market conditions") {
				t.Fatalf("unsupported heading promoted: %+v", got)
			}
			if tc.name == "missing time node cannot justify splitting time-like prose" &&
				(got[1].Kind != "h2" || got[1].Text != "Market conditions") {
				t.Fatalf("corroborated heading not retained without timestamp: %+v", got)
			}
		})
	}
}

func TestReadabilityStructureHeadingLevelAndLinkRetention(t *testing.T) {
	html := `<article><p>Context with a linked report.</p><time>9:10 AM</time><h4>Detailed update</h4></article>`
	input := []article.Block{
		{
			Kind: "p",
			Text: "Context with a linked report. 9:10 AM",
			Parts: []article.TextPart{
				{Text: "Context with a "},
				{Text: "linked report", URL: "https://example.org/report"},
				{Text: ". 9:10 AM"},
			},
		},
		{Kind: "p", Text: "Detailed update"},
	}
	got := PreserveReadabilityStructure(input, html)
	if len(got) != 3 || got[0].Kind != "p" || got[0].Text != "Context with a linked report." ||
		got[1].Kind != "time" || got[1].Text != "9:10 AM" ||
		got[2].Kind != "h4" || got[2].Text != "Detailed update" {
		t.Fatalf("unexpected structural output: %+v", got)
	}
	if got[0].Parts[1].URL != "https://example.org/report" ||
		got[0].Parts[2].Text != "." {
		t.Fatalf("linked text damaged: %+v", got[0].Parts)
	}
	if input[0].Parts[2].Text != ". 9:10 AM" || input[1].Kind != "p" {
		t.Fatalf("caller input modified: %+v", input)
	}
}

// Opt-in baseline: go test ./internal/extraction -run '^$' -bench BenchmarkPreserveReadabilityStructure -benchmem
// The benchmark is synthetic and measures only the structure overlay, not
// fetching, Readability selection, or the complete article processing pipeline.
func BenchmarkPreserveReadabilityStructure(b *testing.B) {
	var html strings.Builder
	html.WriteString("<main><article>")
	blocks := make([]article.Block, 0, 100)
	for i := 0; i < 25; i++ {
		prose := fmt.Sprintf("Editorial update number %d with a link-free paragraph.", i)
		stamp := fmt.Sprintf("%d:00 AM", i+1)
		head := fmt.Sprintf("Confirmed structural headline %d", i)
		fmt.Fprintf(&html, "<p>%s</p><time>%s</time><h2>%s</h2>", prose, stamp, head)
		blocks = append(blocks,
			article.Block{Kind: "p", Text: prose + " " + stamp},
			article.Block{Kind: "p", Text: head},
		)
	}
	html.WriteString("</article></main>")
	source := html.String()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out := PreserveReadabilityStructure(blocks, source)
		if len(out) != 75 {
			b.Fatalf("expected 75 blocks, got %d", len(out))
		}
	}
}
