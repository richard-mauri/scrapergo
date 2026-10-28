package extraction

import "testing"

func TestDiscoverPDFLinkOJSViewer(t *testing.T) {
 const source = "https://journaloflostspecies.org/index.php/jls/article/view/37/33"
 raw := `<html><body><a href="../view/37">Return to Article Details</a><a href="/index.php/jls/article/download/37/33/527">Download PDF</a></body></html>`
 got := DiscoverPDFLink(raw, source)
 want := "https://journaloflostspecies.org/index.php/jls/article/download/37/33/527"
 if got != want { t.Fatalf("PDF link=%q, want %q",got,want) }
}

func TestDiscoverPDFLinkRejectsUnrelatedAndUnsafeReferences(t *testing.T) {
 const src = "https://journal.example.org/article/view/2/3"
 cases := []struct{name,raw string}{
  {"ordinary HTML",`<a href="/news">Read full story</a>`},
  {"cross-origin",`<a href="https://evil.example/article/download/2/3">Download PDF</a>`},
  {"downgrade",`<a href="http://journal.example.org/article/download/2/3">Download PDF</a>`},
  {"credentials",`<a href="https://user:pass@journal.example.org/file.pdf">PDF</a>`},
  {"non-pdf download label",`<a href="/article/download/2/3">Download HTML</a>`},
  {"script scheme",`<a href="javascript:alert(1)">PDF</a>`},
 }
 for _,tc := range cases { t.Run(tc.name,func(t *testing.T){ if got:=DiscoverPDFLink(tc.raw,src);got!="" {t.Fatalf("unexpected candidate %q",got)} }) }
}

func TestDiscoverPDFLinkEmbeddedPDF(t *testing.T) {
 got:=DiscoverPDFLink(`<embed src="/papers/fireflies.pdf#page=1" type="application/pdf">`, "https://journal.example.org/viewer")
 if got != "https://journal.example.org/papers/fireflies.pdf" {t.Fatalf("got %q",got)}
}
