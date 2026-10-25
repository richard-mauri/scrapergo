package reader

import (
	"strings"

	"clean-article-reader/internal/article"
)

// removeAccessChrome excludes publisher UI notices from editorial paragraphs.
// These strings are intentionally narrow: never filter a paragraph merely
// because it mentions signing in, subscriptions, or preview access.
func removeAccessChrome(a *article.Article) {
	if a == nil || len(a.Blocks) == 0 {
		return
	}
	out := a.Blocks[:0]
	for _, b := range a.Blocks {
		if b.Kind == "p" && isAccessChrome(b.Text) {
			continue
		}
		out = append(out, b)
	}
	a.Blocks = out
}

func isAccessChrome(raw string) bool {
	s := strings.ToLower(strings.Join(strings.Fields(raw), " "))
	s = strings.Trim(s, ".! ")
	switch s {
	case "thank you for your patience while we verify access",
		"already a subscriber? log in",
		"want all of the times? subscribe",
		"please enable javascript",
		"we are having trouble retrieving the article content":
		return true
	}
	// Rendered page controls that explicitly describe the current article's
	// authentication flow; don't match ordinary editorial writing or quotes.
	return strings.HasPrefix(s, "you have a preview view of this article while we are checking your access.") ||
		(strings.HasPrefix(s, "thank you for your patience while we verify access.") &&
			strings.Contains(s, "if you are in reader mode please exit"))
}
