package extraction

import (
 "clean-article-reader/internal/article"
 "net/url"
 "strings"
 "testing"
)

func TestRecoverInstitutionalTail(t *testing.T) {
 anchor := "The collection brings together community participation and public art through a nationwide program for families and young athletes."
 about := "The agency was established many years ago and operates a broad range of important public programs. It produces circulating currency, provides public information, and supports specialized collections for residents and visitors. Its programs are managed under a federal mandate, with operational accountability and a commitment to long-term public service."
 raw := "<main><article><p>"+anchor+"</p><p>Learn more about the program.</p></article><section><h2>About the National Agency</h2><p>"+about+"</p></section><section><h2>Additional Resources</h2><p>Sign up for alerts and follow us online.</p></section></main>"
 u,_:=url.Parse("https://example.org/release")
 got:=RecoverInstitutionalTail(raw,[]article.Block{{Kind:"p",Text:anchor},{Kind:"p",Text:"Learn more about the program."}},u)
 if len(got)!=4 {t.Fatalf("want heading and paragraph, got %#v",got)}
 if got[2].Text!="About the National Agency" || got[3].Text!=about {t.Fatalf("wrong tail: %#v",got)}
 for _,b:=range got {if strings.Contains(b.Text,"Sign up") {t.Fatal("included resource panel")}}
}

func TestRecoverInstitutionalTailExcludesGenericFooter(t *testing.T) {
 anchor:="A long selected editorial paragraph about a public event containing enough content to be a reliable source anchor for recovery."
 about:="This corporation publishes information and other resources about its commercial services and products to customers across many markets. Its institutional descriptions describe general corporate responsibilities and customer relationships while pointing visitors toward other resources, business programs, and special offers for the public."
 raw:="<main><article><p>"+anchor+"</p></article></main><footer><h2>About the Corporation</h2><p>"+about+"</p></footer>"
 u,_:=url.Parse("https://example.org/article")
 got:=RecoverInstitutionalTail(raw,[]article.Block{{Kind:"p",Text:anchor}},u)
 if len(got)!=1 {t.Fatalf("footer was incorrectly appended: %#v",got)}
}
