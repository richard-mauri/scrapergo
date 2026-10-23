package extraction

import (
 "clean-article-reader/internal/sourcehtml"
 "testing"

)

func TestInlineArticleImages(t *testing.T) {
	html := `<article><p>This paragraph has enough words to be part of a proper article without being ignored by text </p><figure><picture><source srcset="/image-large.jpg 2x"><img data-src="/photo.jpg" src="data:image/gif;base64,abcd" alt="Photo of match"></picture><figcaption>Photo by Ed McGregor</figcaption></figure><p>A second paragraph with enough words to be included in extracted text without truncation.</p></article>`
	blocks := ExtractBlocks(sourcehtml.Parse(html), "https://example.org/story")
	imageCount := 0
	for _, b := range blocks {
		if b.Kind == "img" {
			imageCount++
			if b.URL != "https://example.org/photo.jpg" {
				t.Fatalf("wrong image URL: %s", b.URL)
			}
		}
	}
	if imageCount != 1 {
		t.Fatalf("expected one image, got %d", imageCount)
	}
}
