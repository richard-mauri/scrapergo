package reader

import (
 "testing"

 "clean-article-reader/internal/article"
)

func TestRemoveAccessChromePreservesEditorial(t *testing.T) {
 a := &article.Article{Blocks: []article.Block{
  {Kind:"p", Text:"You have a preview view of this article while we are checking your access. When we have confirmed access, the full article content will load."},
  {Kind:"p", Text:"An editor interviewed researchers about online security."},
  {Kind:"p", Text:"Thank you for your patience while we verify access. If you are in Reader mode please exit and log into your Times account, or subscribe for all of The Times."},
  {Kind:"p", Text:"Thank you for your patience while we verify access."},
  {Kind:"p", Text:"Already a subscriber? Log in."},
  {Kind:"p", Text:"Want all of The Times? Subscribe."},
  {Kind:"p", Text:"A researcher said, \"Please enable JavaScript\" was the warning on her testing device."},
  {Kind:"img", URL:"https://example.org/photo.jpg"},
 }}
 removeAccessChrome(a)
 if len(a.Blocks) != 3 { t.Fatalf("remaining blocks %d: %+v", len(a.Blocks), a.Blocks) }
 if a.Blocks[0].Text != "An editor interviewed researchers about online security." || a.Blocks[1].Kind != "p" || a.Blocks[2].Kind != "img" { t.Fatalf("editorial blocks changed: %+v", a.Blocks) }
}

func TestRemoveAccessChromePreservesOtherContext(t *testing.T) {
 for _, s := range []string{
  "The publisher asked readers to subscribe after an unexpected service outage.",
  "An author wrote: Thank you for your patience while we verify access.",
  "The subscriber was told to log in before continuing.",
 } {
  if isAccessChrome(s) { t.Fatalf("removed editorial prose: %q",s) }
 }
}
