package accesspolicy

import (
 "clean-article-reader/internal/article"

	"strings"
	"testing"
)

func TestAccessNotices(t *testing.T) {
	for _, kind := range []string{"registration", "paywall", "restricted"} {
		msg := Notice(kind)
		if !strings.Contains(strings.ToLower(msg), "may") {
			t.Fatalf("notice %q promises too much", msg)
		}
	}
}
func TestArticlePreviewJSONFields(t *testing.T) {
	a := article.Article{AccessType: "restricted", AccessNote: Notice("restricted")}
	if a.AccessType != "restricted" || a.AccessNote == "" {
		t.Fatal("preview metadata missing")
	}
}
