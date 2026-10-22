package main

import "testing"

func TestRecoverOpeningSourceFigure(t *testing.T) {
	raw := `<html><body><main><article>
<header><h1>A full news story</h1><div class="byline"><picture><img src="/author.jpg"></picture></div></header>
<figure><picture><img src="/real-lead.jpg" alt="An editorial picture"></picture></figure>
<p>The opening paragraph contains substantial actual editorial prose, providing enough words to mark the beginning of the written news story and distinguish it from the header.</p>
<figure><img src="/later.jpg"></figure></article></main></body></html>`
	a := Article{Blocks: []Block{{Kind:"p",Text:"The opening paragraph contains substantial actual editorial prose."}, {Kind:"img",URL:"https://example.org/other.jpg"}}}
	recoverOpeningSourceFigure(&a, raw, "https://example.org/news/story")
	if len(a.Blocks) != 3 || a.Blocks[0].URL != "https://example.org/real-lead.jpg" || a.Blocks[0].Role != "lead" {
		t.Fatalf("incorrect recovered lead: %+v", a.Blocks)
	}
}

func TestRecoverOpeningSourceFigurePreservesKnownImage(t *testing.T) {
	raw := `<html><main><article><figure><picture><img src="/lead.jpg"></picture></figure><p>This is a long enough paragraph to be treated as genuine body text with an opening statement that contains many separate words for the extraction stage.</p></article></main></html>`
	a := Article{Blocks: []Block{{Kind:"img",URL:"https://example.org/lead.jpg"}, {Kind:"p",Text:"Article text"}}}
	recoverOpeningSourceFigure(&a,raw,"https://example.org/story")
	if len(a.Blocks)!=2 || a.Blocks[0].Role!="lead" {t.Fatalf("should preserve existing image: %+v",a.Blocks)}
}

func TestRecoverOpeningSourceFigureRejectsBylineOnly(t *testing.T) {
	raw := `<html><main><article><header><div class="author"><picture><img src="/author.jpg"></picture></div></header><p>This body paragraph is sufficiently long to mark the start of a real article and includes many words making it recognizable as a prose paragraph.</p></article></main></html>`
	a:=Article{Blocks:[]Block{{Kind:"p",Text:"Text"}}}
	recoverOpeningSourceFigure(&a,raw,"https://example.org/story")
	if len(a.Blocks)!=1 {t.Fatalf("byline portrait incorrectly recovered: %+v",a.Blocks)}
}
