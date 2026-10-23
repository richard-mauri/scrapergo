package main

import (
 "testing"
 "clean-article-reader/internal/recovery"
)

func TestRecoverOpeningSourceFigure(t *testing.T) {
	raw := `<html><body><main><article>
<header><h1>A full news story</h1><div class="byline"><picture><img src="/author.jpg"></picture></div></header>
<figure><picture><img src="/real-lead.jpg" alt="An editorial picture"></picture></figure>
<p>The opening paragraph contains substantial actual editorial prose, providing enough words to mark the beginning of the written news story and distinguish it from the header.</p>
<figure><img src="/later.jpg"></figure></article></main></body></html>`
	a := Article{Blocks: []Block{{Kind:"p",Text:"The opening paragraph contains substantial actual editorial prose."}, {Kind:"img",URL:"https://example.org/other.jpg"}}}
	recovery.RecoverOpeningSourceFigure(&a, raw, "https://example.org/news/story")
	if len(a.Blocks) != 3 || a.Blocks[0].URL != "https://example.org/real-lead.jpg" || a.Blocks[0].Role != "lead" {
		t.Fatalf("incorrect recovered lead: %+v", a.Blocks)
	}
}

func TestRecoverOpeningSourceFigurePreservesKnownImage(t *testing.T) {
	raw := `<html><main><article><figure><picture><img src="/lead.jpg"></picture></figure><p>This is a long enough paragraph to be treated as genuine body text with an opening statement that contains many separate words for the extraction stage.</p></article></main></html>`
	a := Article{Blocks: []Block{{Kind:"img",URL:"https://example.org/lead.jpg"}, {Kind:"p",Text:"Article text"}}}
	recovery.RecoverOpeningSourceFigure(&a,raw,"https://example.org/story")
	if len(a.Blocks)!=2 || a.Blocks[0].Role!="lead" {t.Fatalf("should preserve existing image: %+v",a.Blocks)}
}

func TestRecoverOpeningSourceFigureRejectsBylineOnly(t *testing.T) {
	raw := `<html><main><article><header><div class="author"><picture><img src="/author.jpg"></picture></div></header><p>This body paragraph is sufficiently long to mark the start of a real article and includes many words making it recognizable as a prose paragraph.</p></article></main></html>`
	a:=Article{Blocks:[]Block{{Kind:"p",Text:"Text"}}}
	recovery.RecoverOpeningSourceFigure(&a,raw,"https://example.org/story")
	if len(a.Blocks)!=1 {t.Fatalf("byline portrait incorrectly recovered: %+v",a.Blocks)}
}

func TestRecoverOpeningArticleHeaderFeaturedImage(t *testing.T) {
 raw := `<html><main><article><header><h1>Windshield frit</h1><div class="entry-featured-image"><img src="/featured.png?w=800"></div><div class="byline"><img src="/author.jpg"></div></header><div class="entry-content"><p>This article discusses the small dots around windshields and why glass makers use the dark frit band in modern vehicles for several useful purposes.</p><figure><img src="/frit.png?w=400"></figure></div></article></main></html>`
 a := Article{Blocks: []Block{{Kind:"p",Text:"This article discusses the small dots around windshields."},{Kind:"img",URL:"https://example.org/frit.png?w=400"}}}
 recovery.RecoverOpeningSourceFigure(&a,raw,"https://example.org/story")
 if len(a.Blocks)!=3 || a.Blocks[0].URL!="https://example.org/featured.png?w=800" || a.Blocks[0].Role!="lead" || a.Blocks[2].URL!="https://example.org/frit.png?w=400" {t.Fatalf("header lead not recovered: %+v",a.Blocks)}
 recovery.RecoverOpeningSourceFigure(&a,raw,"https://example.org/story")
 if len(a.Blocks)!=3 {t.Fatalf("duplicate lead recovered: %+v",a.Blocks)}
}

func TestRecoverOpeningArticleHeaderDoesNotPromoteAuthorPortrait(t *testing.T) {
 raw := `<main><article><header><div class="byline"><picture><img src="/portrait.jpg"></picture></div><img src="/decoration.png"></header><p>This is the very substantial beginning of an editorial article containing enough distinct words to establish prose and restrict image recovery to the correct story boundary.</p><figure><img src="/later.jpg"></figure></article></main>`
 a := Article{Blocks: []Block{{Kind:"p",Text:"Editorial text."}}}
 recovery.RecoverOpeningSourceFigure(&a,raw,"https://example.org/story")
 if len(a.Blocks)!=1 {t.Fatalf("author or decoration promoted: %+v",a.Blocks)}
}
