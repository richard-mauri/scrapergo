package extraction

import (
 "clean-article-reader/internal/sourcehtml"


	"net/url"
	"strings"
	"testing"
)

func TestSerializedCaptionImagesGeneric(t *testing.T) {
	base, _ := url.Parse("https://example.org/articles/guide")
	raw := `<script type="application/json">["https://cdn.example.org/media/photo.webp","A detailed caption about the main illustration in this technical guide."]</script>`
	got := SerializedCaptionImages(raw, base)
	caption := "A detailed caption about the main illustration in this technical guide."
	if got[caption] != "https://cdn.example.org/media/photo.webp" {
		t.Fatalf("image/caption association lost: %v", got)
	}
}

func TestSemanticTextBlockKeepsMarkup(t *testing.T) {
	base, _ := url.Parse("https://example.org/guide")
	root := sourcehtml.Parse(`<p><strong>First:</strong> read the guide<br><a href="/details"><em>Second:</em> follow the link</a></p>`)
	ps := find(root, func(n *sourcehtml.Node) bool { return n.Tag == "p" })
	if len(ps) != 1 {
		t.Fatal("paragraph missing")
	}
	b := SemanticTextBlock(ps[0], base)
	hasBold, hasNewline, hasLink := false, false, false
	for _, p := range b.Parts {
		if p.Bold && strings.Contains(p.Text, "First:") {
			hasBold = true
		}
		if strings.Contains(p.Text, "\n") {
			hasNewline = true
		}
		if strings.Contains(p.URL, "/details") {
			hasLink = true
		}
	}
	if !hasBold || !hasNewline || !hasLink {
		t.Fatalf("semantic formatting lost: %#v", b.Parts)
	}
}
