package main

import (
 "strings"
 "testing"
)

func TestReaderDoesNotRenderHeroTwice(t *testing.T) {
 html:=renderedIndexHTML()
 if !strings.Contains(html,"if(data.imageUrl&&sameArticleImage(b.url,data.imageUrl))continue;") {t.Fatal("duplicate body hero image not skipped")}
 if !strings.Contains(html,"function sameArticleImage(a,b)") {t.Fatal("missing image URL comparison")}
 if !strings.Contains(html,"aName.length>=16&&aName===bName") {t.Fatal("missing filename comparison for CDN-resized duplicates")}
 if !strings.Contains(html,"if(data.imageUrl){const pic=") {t.Fatal("hero image rendering unexpectedly removed")}
 if !strings.Contains(html,"fig.appendChild(im)") {t.Fatal("other inline images must be retained")}
}
