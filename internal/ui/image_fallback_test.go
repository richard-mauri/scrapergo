package ui

import (
 "strings"
 "testing"
)

func TestFailedImageFallbackInReader(t *testing.T) {
 js := readerJS
 for _,want := range []string{
  "function removeFailedImage(img,source,article)",
  "img.addEventListener('error'",
  "figure.remove()",
  "removeFailedImage(im,item.url,data);fig.appendChild(im)",
  "removeFailedImage(pic,first.url,data);group.appendChild(pic)",
  "editorialCaption(b.text,data.title)",
  "editorialCaption(item.text,data.title)",
  "img.naturalWidth<=1",
  "/api/article-image?url=",
  "encodeURIComponent(source)",
  "removeFailedImage(im,b.url,data)",
 } {
  if !strings.Contains(js,want) { t.Errorf("missing safe image fallback: %s",want) }
 }
 if !strings.Contains(js,"im.src=b.url") || !strings.Contains(js,"im.src=item.url") {
  t.Fatal("source image URLs must remain unchanged for successfully loaded images")
 }
}
