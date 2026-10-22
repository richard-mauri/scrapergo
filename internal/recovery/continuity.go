// Package recovery owns publisher-neutral decisions for restoring missing
// editorial text. It has no network, HTML-parser, or HTTP dependencies.
package recovery

import (
 "strings"
 "clean-article-reader/internal/article"
 "clean-article-reader/internal/sourcehtml"
)

// OrderedMatches counts consecutive retained paragraphs corroborated by
// source candidate paragraphs, preserving legacy matching and ordering.
func OrderedMatches(retained []string, candidate []article.Block) int {
 index, matches := 0, 0
 for _, b := range candidate {
  if index >= len(retained) { break }
  if b.Kind != "p" { continue }
  a, c := retained[index], Normalize(b.Text)
  if a == c || (len(a) >= 65 && len(c) >= 65 &&
   (strings.HasPrefix(a,c) || strings.HasPrefix(c,a))) {
   index++
   matches++
  }
 }
 return matches
}

// Normalize matches whitespace and entity handling of sourcehtml.Normalize,
// including the two-pass normalization of text assembled by legacy callers.
func Normalize(s string) string {
 return sourcehtml.Normalize(s)
}

// CountWords preserves the established count: image blocks do not contribute.
func CountWords(blocks []article.Block) int {
 n:=0
 for _, b:=range blocks {if b.Kind!="img" {n+=len(strings.Fields(b.Text))}}
 return n
}

// ReplacementAllowed keeps the existing conservative rule: a source article
// must supply new editorial material without reducing selected prose by >10%.
func ReplacementAllowed(selected, candidate []article.Block, selectedText map[string]bool, selectedImages map[string]bool) bool {
 extraParagraphs, extraImages := 0, 0
 for _, b:=range candidate {
  if b.Kind=="p" && len(b.Text)>=25 && !selectedText[Normalize(b.Text)] {extraParagraphs++}
  if b.Kind=="img" && !selectedImages[b.URL] {extraImages++}
 }
 if extraParagraphs<2 && extraImages==0 {return false}
 return CountWords(candidate)*10>=CountWords(selected)*9
}
