package main

import (
 "strings"
 "testing"
)

func TestStripEmbeddedMediaPreservesEditorialNeighbors(t *testing.T) {
 raw:=`<!doctype html><html><body><article><h1>Example</h1><p>Before the audio player.</p><div class="player"><iframe src="https://media.example/player"></iframe></div><p>After the audio player.</p><figure><img src="https://example.org/lead.jpg"/><figcaption>A legitimate caption</figcaption></figure></article></body></html>`
 cleaned:=stripEmbeddedMedia(raw)
 if strings.Contains(cleaned,"<iframe")||strings.Contains(cleaned,"media.example/player"){t.Fatal("media element survived structural cleanup")}
 for _,v:=range []string{"Before the audio player.","After the audio player.","A legitimate caption","lead.jpg"} {if !strings.Contains(cleaned,v){t.Errorf("lost editorial content %q",v)}}
}
func TestStripEmbeddedMediaOtherElements(t *testing.T) {
 raw:=`<html><body><article><p>Report</p><video><source src="movie.mp4"/></video><audio src="audio.mp3"></audio><object data="widget"></object><embed src="widget"/><p>Next paragraph</p></article></body></html>`
 cleaned:=stripEmbeddedMedia(raw)
 for _,v:=range []string{"<video","<audio","<object","<embed","movie.mp4","audio.mp3"}{if strings.Contains(cleaned,v){t.Errorf("media element survived: %s",v)}}
 if !strings.Contains(cleaned,"Next paragraph"){t.Fatal("removed following paragraph")}
}
func TestNormalizeArticleDocumentTitleAndImage(t *testing.T) {
 a:=Article{Title:"News headline",ImageURL:"https://example.org/hero.jpg",Blocks:[]Block{{Kind:"h2",Text:"News headline"},{Kind:"p",Text:"Paragraph"},{Kind:"img",URL:"https://example.org/hero.jpg"},{Kind:"img",URL:"https://example.org/portrait.jpg"}}}
 normalizeArticleDocument(&a)
 if len(a.Blocks)!=3 || a.Blocks[0].Text!="Paragraph"{t.Fatalf("opening title handling: %+v",a.Blocks)}
 if a.ImageURL!="https://example.org/hero.jpg" || a.Blocks[1].Role!="lead" {t.Fatalf("lead identity or metadata lost: %+v",a)}
 if a.Blocks[2].URL!="https://example.org/portrait.jpg"{t.Fatal("portrait lost")}
}
func TestNormalizeArticleDocumentDistinctHero(t *testing.T) {
 a:=Article{Title:"News headline",ImageURL:"https://example.org/lead.jpg",Blocks:[]Block{{Kind:"img",Role:"author",URL:"https://example.org/author.jpg"},{Kind:"p",Text:"Editorial paragraph"}}}
 normalizeArticleDocument(&a)
 if len(a.Blocks)!=3 || a.Blocks[0].Role!="lead" || a.Blocks[0].URL!="https://example.org/lead.jpg" || a.Blocks[1].URL!="https://example.org/author.jpg" {t.Fatalf("distinct lead or author image missing: %+v",a.Blocks)}
}

func TestEscapedPlayerMarkupIsRemovedWithoutDeletingProse(t *testing.T) {
 raw:=`<html><body><article><p>Before &lt;iframe src="https://audio.example/player" width="100%"&gt; after.</p><p>Later article content remains.</p></article></body></html>`
 cleaned:=stripEmbeddedMedia(raw)
 if strings.Contains(cleaned,"audio.example/player") {t.Fatal("escaped media element leaked into output")}
 if !strings.Contains(cleaned,"Before")||!strings.Contains(cleaned,"after.")||!strings.Contains(cleaned,"Later article content") {t.Fatal("legitimate surrounding prose lost")}
}
func TestCodeSamplesAreNotDiscarded(t *testing.T) {
 raw:=`<html><body><article><pre><code>&lt;iframe src="https://docs.example/sample"&gt;</code></pre></article></body></html>`
 cleaned:=stripEmbeddedMedia(raw)
 if !strings.Contains(cleaned,"docs.example/sample"){t.Fatal("legitimate technical code example was removed")}
}

func TestSourceMetadataMarksMatchingInlineLead(t *testing.T) {
 a:=Article{Title:"Story",Blocks:[]Block{{Kind:"img",Role:"author",URL:"https://example.org/author.jpg"},{Kind:"img",URL:"https://example.org/cover.jpg"}}}
 normalizeArticleDocument(&a,"https://example.org/cover.jpg")
 if len(a.Blocks)!=2 || a.Blocks[0].Role=="lead" || a.Blocks[1].Role!="lead" {t.Fatalf("lead duplicated or misidentified: %+v",a.Blocks)}
}
