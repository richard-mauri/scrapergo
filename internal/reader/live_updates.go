package reader

import (
	"clean-article-reader/internal/article"
	"net/url"
	"strings"
)

// removeRepeatedLiveUpdate removes an earlier featured copy of a live-blog
// update only when a long, ordered run of textual blocks appears again later.
// The later copy normally retains its chronological timestamp. We deliberately
// do not deduplicate individual paragraphs, quotations, or ordinary articles.
func removeRepeatedLiveUpdate(a *article.Article) {
	u, err := url.Parse(a.URL)
	if err != nil || !strings.Contains(u.Path, "/live/") {
		return
	}

	// Compare text blocks in reading order while leaving photos and other
	// non-text media intact. Paragraphs, labels and headings count as text;
	// short headings alone cannot satisfy the substantial-run threshold.
	positions := make([]int, 0, len(a.Blocks))
	keys := make([]string, 0, len(a.Blocks))
	for i, b := range a.Blocks {
		if b.Kind == "img" || b.Kind == "table" || b.Text == "" {
			continue
		}
		text := strings.Join(strings.Fields(b.Text), " ")
		if text == "" {
			continue
		}
		positions = append(positions, i)
		keys = append(keys, text)
	}
	// Only an early featured update is eligible for removal.
	const maxLeadBlocks = 25
	const minMatchingBlocks = 6
	const minMatchingWords = 140
	const minLongParagraphs = 4
	bestStart, bestLength := -1, 0
	for start := 0; start < len(keys) && start < maxLeadBlocks; start++ {
		for later := start + minMatchingBlocks + 2; later < len(keys); later++ {
			length, words, longParagraphs := 0, 0, 0
			for start+length < later && later+length < len(keys) && keys[start+length] == keys[later+length] {
				t := keys[start+length]
				words += len(strings.Fields(t))
				if len(strings.Fields(t)) >= 18 {
					longParagraphs++
				}
				length++
			}
			if length >= minMatchingBlocks && words >= minMatchingWords && longParagraphs >= minLongParagraphs && length > bestLength {
				bestStart, bestLength = start, length
			}
		}
	}
	if bestStart < 0 {
		return
	}
	removed := make(map[int]bool, bestLength)
	for i := bestStart; i < bestStart+bestLength; i++ {
		removed[positions[i]] = true
	}
	kept := make([]article.Block, 0, len(a.Blocks)-len(removed))
	for i, b := range a.Blocks {
		if !removed[i] {
			kept = append(kept, b)
		}
	}
	a.Blocks = kept
	// Match the existing extractor's text-count contract after removal.
	words := 0
	for _, b := range kept {
		if b.Kind != "img" && b.Kind != "table" {
			words += len(strings.Fields(b.Text))
		}
	}
	a.WordCount = words
}
