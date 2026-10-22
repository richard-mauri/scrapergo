package links

import (
 "strings"
 "unicode"
)

// semanticKey compares complete paragraph text without depending on whitespace
// differences introduced by HTML layout or Readability's text renderer.
func SemanticKey(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			b.WriteRune(r)
		} else {
			b.WriteRune(' ')
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

