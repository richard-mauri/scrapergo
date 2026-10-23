package structured

import (
 "testing"
 "strings"
)

func TestStripEmbeddedMediaPreservesEditorialNeighbors(t *testing.T) {
 raw:=`<!doctype html><html><body><article><h1>Example</h1><p>Before the audio player.</p><div class="player"><iframe src="https://media.example/player"></iframe></div><p>After the audio player.</p><figure><img src="https://example.org/lead.jpg"/><figcaption>A legitimate caption</figcaption></figure></article></body></html>`
 cleaned:=StripEmbeddedMedia(raw)
 if strings.Contains(cleaned,"<iframe")||strings.Contains(cleaned,"media.example/player"){t.Fatal("media element survived structural cleanup")}
 for _,v:=range []string{"Before the audio player.","After the audio player.","A legitimate caption","lead.jpg"} {if !strings.Contains(cleaned,v){t.Errorf("lost editorial content %q",v)}}
}

func TestStripEmbeddedMediaOtherElements(t *testing.T) {
 raw:=`<html><body><article><p>Report</p><video><source src="movie.mp4"/></video><audio src="audio.mp3"></audio><object data="widget"></object><embed src="widget"/><p>Next paragraph</p></article></body></html>`
 cleaned:=StripEmbeddedMedia(raw)
 for _,v:=range []string{"<video","<audio","<object","<embed","movie.mp4","audio.mp3"}{if strings.Contains(cleaned,v){t.Errorf("media element survived: %s",v)}}
 if !strings.Contains(cleaned,"Next paragraph"){t.Fatal("removed following paragraph")}
}

func TestEscapedPlayerMarkupIsRemovedWithoutDeletingProse(t *testing.T) {
 raw:=`<html><body><article><p>Before &lt;iframe src="https://audio.example/player" width="100%"&gt; after.</p><p>Later article content remains.</p></article></body></html>`
 cleaned:=StripEmbeddedMedia(raw)
 if strings.Contains(cleaned,"audio.example/player") {t.Fatal("escaped media element leaked into output")}
 if !strings.Contains(cleaned,"Before")||!strings.Contains(cleaned,"after.")||!strings.Contains(cleaned,"Later article content") {t.Fatal("legitimate surrounding prose lost")}
}

func TestCodeSamplesAreNotDiscarded(t *testing.T) {
 raw:=`<html><body><article><pre><code>&lt;iframe src="https://docs.example/sample"&gt;</code></pre></article></body></html>`
 cleaned:=StripEmbeddedMedia(raw)
 if !strings.Contains(cleaned,"docs.example/sample"){t.Fatal("legitimate technical code example was removed")}
}

