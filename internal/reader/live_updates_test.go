package reader

import (
	"clean-article-reader/internal/article"
	"fmt"
	"strings"
	"testing"
)

func exampleLiveBlocks() ([]article.Block, string) {
	paragraphs := make([]article.Block, 7)
	for i := range paragraphs {
		paragraphs[i] = article.Block{
			Kind: "p",
			Text: fmt.Sprintf("Update paragraph %d reports specific developing information and contextual details for readers following the election, with enough substantive words to distinguish a repeated editorial update from navigation or a quotation.", i),
		}
	}
	blocks := []article.Block{
		{Kind: "p", Text: "From 24 minutes ago"},
		{Kind: "h2", Text: "Latest election developments"},
		{Kind: "img", URL: "https://images.example.org/lead.jpg", Text: "Lead photograph"},
	}
	// Marker from the owner's exported text: the critical opening paragraph
	// must survive before Key events, not merely appear somewhere downstream.
	paragraphs[0].Text = "In case you fell asleep early, " + paragraphs[0].Text
	blocks = append(blocks, paragraphs...)
	blocks = append(blocks,
		article.Block{Kind: "h2", Text: "Key events"},
		article.Block{Kind: "p", Text: "08:05 - An independent live update about another contested race."},
		article.Block{Kind: "p", Text: "07:52 - A different update about a ballot measure."},
		article.Block{Kind: "p", Text: "07:42 EST"},
		article.Block{Kind: "h2", Text: "Latest election developments"},
	)
	blocks = append(blocks, paragraphs...)
	return blocks, paragraphs[0].Text
}

func TestLiveBlogRepeatedFeaturedUpdate(t *testing.T) {
	blocks, repeated := exampleLiveBlocks()
	a := article.Article{URL: "https://example.org/us-news/live/2026/11/04/results", Blocks: blocks, WordCount: 9999}
	removeRepeatedLiveUpdate(&a)
	if len(a.Blocks) != len(blocks)-8 {
		t.Fatalf("expected duplicated heading and seven paragraphs removed, got %d of %d", len(a.Blocks), len(blocks))
	}
	if a.Blocks[2].Kind != "img" || a.Blocks[2].URL != "https://images.example.org/lead.jpg" {
		t.Fatal("lead editorial image lost")
	}
	if a.Blocks[1].Text != "Latest election developments" || !strings.HasPrefix(a.Blocks[3].Text, "In case you fell asleep") {
		t.Fatal("opening editorial heading or paragraph was deleted")
	}
	texts := make([]string, 0, len(a.Blocks))
	for _, b := range a.Blocks {
		texts = append(texts, b.Text)
	}
	all := strings.Join(texts, "\n")
	if strings.Count(all, repeated) != 1 {
		t.Fatal("repeated paragraphs were not reduced to one occurrence")
	}
	if indexLead, indexKey := strings.Index(all, "In case you fell asleep"), strings.Index(all, "Key events"); indexLead < 0 || indexKey < 0 || indexLead >= indexKey {
		t.Fatal("opening update must precede Key events")
	}
	if indexEnd, indexTimestamp := strings.Index(all, "Update paragraph 6"), strings.Index(all, "07:42 EST"); indexEnd < 0 || indexTimestamp <= indexEnd {
		t.Fatal("chronological timestamp must follow retained opening update")
	}
	for _, required := range []string{"08:05", "07:52", "07:42 EST", "Key events", "Latest election developments"} {
		if !strings.Contains(all, required) {
			t.Fatalf("independent update or timestamp missing: %s", required)
		}
	}
	if a.WordCount < 140 || a.WordCount > 350 {
		t.Fatalf("word count not recalculated after pruning: %d", a.WordCount)
	}
}

func TestLiveBlogDedupRespectsUnrelatedArticlesAndShortRepeats(t *testing.T) {
	original, _ := exampleLiveBlocks()
	for _, url := range []string{
		"https://example.org/news/single-article",
		"https://example.org/us-news/live/2026/11/04/results",
	} {
		blocks := append([]article.Block(nil), original...)
		if strings.Contains(url, "/live/") {
			// A single repeated quotation is editorially legitimate.
			blocks = []article.Block{
				{Kind: "p", Text: "A repeated quote that readers might see twice."},
				{Kind: "p", Text: "An unrelated update with a separate context."},
				{Kind: "p", Text: "A repeated quote that readers might see twice."},
			}
		}
		a := article.Article{URL: url, Blocks: blocks, WordCount: 111}
		removeRepeatedLiveUpdate(&a)
		if len(a.Blocks) != len(blocks) || a.WordCount != 111 {
			t.Fatalf("ordinary article or short repeated quote mutated for %s", url)
		}
	}
}
