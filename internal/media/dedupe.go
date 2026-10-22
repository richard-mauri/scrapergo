// Package media owns publisher-neutral image identity and deduplication.
package media

import (
 "net/url"
 "regexp"
 "strings"

 "clean-article-reader/internal/article"
)

var imageSizeSuffix = regexp.MustCompile(`(?i)-[0-9]{2,5}x[0-9]{2,5}(\.[a-z0-9]+)$`)

// Identity ignores delivery parameters and conventional size suffixes.
// Distinct source filenames and paths retain distinct identities.
func Identity(raw string) string {
 u, err := url.Parse(strings.TrimSpace(raw))
 if err != nil || u.Hostname() == "" { return "" }
 path := strings.ToLower(u.EscapedPath())
 path = imageSizeSuffix.ReplaceAllString(path, "$1")
 return strings.ToLower(u.Hostname()) + path
}

// Deduplicate preserves the first inline occurrence of each photograph.
// Metadata ImageURL is not an article block and must not consume an image
// identity: normalization reconciles the source's metadata rendition later.
// Other blocks, captions and attribution are unchanged.
func Deduplicate(a *article.Article) {
 if a == nil { return }
 seen := map[string]bool{}
 blocks := make([]article.Block,0,len(a.Blocks))
 for _, b := range a.Blocks {
  if b.Kind == "img" {
   key := Identity(b.URL)
   if key != "" && seen[key] { continue }
   if key != "" { seen[key] = true }
  }
  blocks = append(blocks,b)
 }
 a.Blocks = blocks
}
