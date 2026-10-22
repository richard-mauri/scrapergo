package media

import "clean-article-reader/internal/article"

// RemovePostEditorialAuthorPortraits excludes contributor profile photos from
// the article body after the editorial story has begun. Contributor portraits
// preceding the editorial lead (as in multi-author headers) are preserved.
// Byline metadata and every editorial image remain unchanged.
func RemovePostEditorialAuthorPortraits(a *article.Article) {
 if a == nil { return }
 started := false
 out := make([]article.Block,0,len(a.Blocks))
 for _,b := range a.Blocks {
  if b.Kind == "img" && b.Role == "author" && started { continue }
  out = append(out,b)
  if (b.Kind == "img" && (b.Role == "lead" || b.Role == "editorial")) ||
     (b.Kind == "p" && len(b.Text) >= 90) { started = true }
 }
 a.Blocks=out
}
