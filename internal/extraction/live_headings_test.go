package extraction

import (
	"clean-article-reader/internal/article"
	"net/url"
	"testing"
)

func TestRestoreLiveHeadingsPreservesEditorialUpdates(t *testing.T) {
	raw := `<html><body><nav><h2>Unrelated navigation headline that should not be promoted</h2></nav><main>
<article><h2>Adviser says Tehran is one of the main architects of security</h2>
<p>This preceding update has a substantive paragraph that must remain in place.</p></article>
<article><h2>Tanker hit by explosion in Strait of Hormuz, naval agency says</h2>
<p>A tanker was hit by an explosion Thursday in the Strait of Hormuz.</p></article>
</main></body></html>`
	blocks := []article.Block{
		{Kind: "p", Text: "Adviser says Tehran is one of the main architects of security"},
		{Kind: "p", Text: "This preceding update has a substantive paragraph that must remain in place."},
		{Kind: "p", Text: "Tanker hit by explosion in Strait of Hormuz, naval agency says"},
		{Kind: "p", Text: "A tanker was hit by an explosion Thursday in the Strait of Hormuz."},
		{Kind: "p", Text: "Unrelated navigation headline that should not be promoted"},
	}
	u, _ := url.Parse("https://www.cbsnews.com/live/updates/")
	got := RestoreLiveHeadings(raw, blocks, u)
	for i, kind := range []string{"h3", "p", "h3", "p", "p"} {
		if got[i].Kind != kind {
			t.Fatalf("block %d kind = %q, want %q", i, got[i].Kind, kind)
		}
	}
	if got[1].Text != "This preceding update has a substantive paragraph that must remain in place." {
		t.Fatal("preceding editorial update was changed")
	}
}

func TestRestoreLiveHeadingsDoesNotChangeOrdinaryArticles(t *testing.T) {
	raw := `<main><article><h2>A real article heading longer than fifteen characters</h2></article></main>`
	blocks := []article.Block{{Kind: "p", Text: "A real article heading longer than fifteen characters"}}
	u, _ := url.Parse("https://example.com/news/article")
	got := RestoreLiveHeadings(raw, blocks, u)
	if got[0].Kind != "p" {
		t.Fatal("ordinary article paragraph was promoted")
	}
}
