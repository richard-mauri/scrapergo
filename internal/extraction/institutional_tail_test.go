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

func TestRecoverInstitutionalTailRetainsShortFinalParagraph(t *testing.T) {
 anchor:="The new public program builds on a longstanding national tradition of encouraging participation while offering opportunities for families and community organizations."
 first:="The national agency was established over a century ago and now produces a wide range of public services, including materials used across the country. Its public responsibilities include administering national projects, distributing specialized information, and providing support to communities through many carefully managed programs."
 second:="Its programs are self-sustaining and operate without taxpayer support."
 raw:="<main><article><p>"+anchor+"</p><p>Learn more about the program.</p></article><section><h2>About the National Agency</h2><p>"+first+"</p><p>"+second+"</p></section><section><h2>Additional Resources</h2><p>Sign up for product notifications.</p></section></main>"
 u,_:=url.Parse("https://example.org/release")
 got:=RecoverInstitutionalTail(raw,[]article.Block{{Kind:"p",Text:anchor},{Kind:"p",Text:"Learn more about the program."}},u)
 if len(got)!=5 {t.Fatalf("expected heading and two paragraphs, got %#v",got)}
 if got[4].Text!=second {t.Fatalf("short final paragraph not preserved: %#v",got)}
}

func TestRecoverInstitutionalTailInAdjacentPublisherPanel(t *testing.T) {
 anchor:="A public initiative will bring new programs to communities and offer participating families opportunities to learn more about the national collection."
 body:="Established over two centuries ago, the public agency performs essential responsibilities for the nation through a broad range of operational programs. Its ongoing public mission includes production, administration and stewardship of specialized materials while carrying out legislative directives and serving communities across the country."
 raw:="<div><main><article><p>"+anchor+"</p><p>Learn more about the collection.</p></article></main><section><h2>About the Agency</h2><p>"+body+"</p><p>These programs operate at no cost to taxpayers.</p></section><section><h2>Additional Resources</h2><p>Sign up and follow us.</p></section></div>"
 u,_:=url.Parse("https://example.org/press-release")
 got:=RecoverInstitutionalTail(raw,[]article.Block{{Kind:"p",Text:anchor},{Kind:"p",Text:"Learn more about the collection."}},u)
 if len(got)!=5 {t.Fatalf("publisher sibling panel not recovered: %#v",got)}
 if got[4].Text!="These programs operate at no cost to taxpayers." {t.Fatalf("wrong ending: %#v",got)}
}
