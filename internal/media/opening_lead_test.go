package media

import (
	"strings"
	"testing"

	"clean-article-reader/internal/article"
)

func TestReconcileOpeningLead(t *testing.T) {
	identity := func(url string) string { return strings.Split(url, "?")[0] }
	t.Run("prepend missing lead and preserve order", func(t *testing.T) {
		a := article.Article{Blocks: []article.Block{
			{Kind: "p", Text: "Beginning"},
			{Kind: "img", URL: "https://example.org/later.jpg"},
		}}
		if !ReconcileOpeningLead(&a, "https://example.org/opening.jpg", identity) {
			t.Fatal("expected recovered opening image")
		}
		if len(a.Blocks) != 3 || a.Blocks[0].URL != "https://example.org/opening.jpg" ||
			a.Blocks[0].Role != "lead" || a.Blocks[1].Text != "Beginning" ||
			a.Blocks[2].URL != "https://example.org/later.jpg" {
			t.Fatalf("incorrect recovery: %+v", a.Blocks)
		}
	})
	t.Run("promote existing rendition without duplicating", func(t *testing.T) {
		a := article.Article{Blocks: []article.Block{
			{Kind: "img", URL: "https://example.org/opening.jpg?w=800"},
			{Kind: "img", URL: "https://example.org/later.jpg"},
		}}
		if !ReconcileOpeningLead(&a, "https://example.org/opening.jpg?w=400", identity) {
			t.Fatal("expected promotion")
		}
		if len(a.Blocks) != 2 || a.Blocks[0].Role != "lead" ||
			a.Blocks[0].URL != "https://example.org/opening.jpg?w=800" ||
			a.Blocks[1].Role != "" {
			t.Fatalf("rendition duplicated or gallery changed: %+v", a.Blocks)
		}
	})
	t.Run("no candidate means no mutation", func(t *testing.T) {
		a := article.Article{Blocks: []article.Block{{Kind: "p", Text: "Keep"}}}
		if ReconcileOpeningLead(&a, "", identity) || len(a.Blocks) != 1 {
			t.Fatalf("unexpected empty-candidate mutation: %+v", a.Blocks)
		}
	})
}
