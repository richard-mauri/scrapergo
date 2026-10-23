package extraction

import (
 articlemodel "clean-article-reader/internal/article"
"fmt"
	"net/url"
	"strings"
	"testing"
)

func TestGenericStructuredDocument(t *testing.T) {
	var raw strings.Builder
	raw.WriteString(`<html><nav><h2>Navigation heading</h2></nav><main><article><p>This comprehensive guide contains technical descriptions of the equipment and their various applications.</p><ul><li>First safety instruction provides essential operating guidance.</li><li>Second safety instruction provides important equipment details.</li></ul>`)
	for i := 1; i <= 8; i++ {
		fmt.Fprintf(&raw, `<section><h2>Model %d overview</h2><img src="/img/%d.jpg" alt="Model %d illustration"/><p>Model %d includes multiple technical specifications and a detailed discussion of the design.</p></section>`, i, i, i, i)
	}
	raw.WriteString(`</article></main><footer><h2>Footer</h2></footer></html>`)
	base, _ := url.Parse("https://example.org/guide")
	got, ok := RecoverStructuredDocument(raw.String(), []articlemodel.Block{{Kind: "p", Text: "Model 1 overview"}}, base)
	if !ok {
		t.Fatal("missed structured content")
	}
	heads, imgs, lis := 0, 0, 0
	for _, b := range got {
		switch b.Kind {
		case "h2":
			heads++
		case "img":
			imgs++
		case "li":
			lis++
		}
	}
	if heads != 8 || imgs != 8 || lis != 2 {
		t.Fatalf("headings=%d imgs=%d bullets=%d", heads, imgs, lis)
	}
	if !strings.Contains(got[len(got)-1].Text, "design") {
		t.Fatalf("out of order: %#v", got[len(got)-1])
	}
}
func TestGenericStructuredDocumentDeclinesMenu(t *testing.T) {
	var raw strings.Builder
	raw.WriteString(`<nav>`)
	for i := 0; i < 20; i++ {
		fmt.Fprintf(&raw, `<h2>Menu item number %d</h2>`, i)
	}
	raw.WriteString(`</nav><main><p>Ordinary body paragraph remains readable in its original order.</p></main>`)
	base, _ := url.Parse("https://example.org/")
	if _, ok := RecoverStructuredDocument(raw.String(), nil, base); ok {
		t.Fatal("menu accepted as article")
	}
}
