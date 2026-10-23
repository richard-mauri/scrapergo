package ui

import (
 "strings"
 "testing"
)

func TestFailedImageFallbackInReader(t *testing.T) {
 js := readerJS
 for _,want := range []string{
  "function removeFailedImage(img)",
  "img.addEventListener('error'",
  "figure.remove()",
  "removeFailedImage(im);fig.appendChild(im)",
  "removeFailedImage(pic);group.appendChild(pic)",
  "editorialCaption(b.text,data.title)",
  "editorialCaption(item.text,data.title)",
 } {
  if !strings.Contains(js,want) { t.Errorf("missing safe image fallback: %s",want) }
 }
 if !strings.Contains(js,"im.src=b.url") || !strings.Contains(js,"im.src=item.url") {
  t.Fatal("source image URLs must remain unchanged for successfully loaded images")
 }
}
