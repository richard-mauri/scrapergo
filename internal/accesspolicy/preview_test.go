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


func TestExplicitPreviewEvidence(t *testing.T) {
 cases := []struct{name, html string; want bool}{
  {"rendered preview", `<p>You have a preview view of this article while we are checking your access.</p>`, true},
  {"subscription metadata", `<p>Subscribe to read the full article.</p>`, false},
  {"script only", `<script>const msg="You have a preview view of this article while we are checking your access"</script><p>News report.</p>`, false},
  {"short ordinary article", `<article><p>Report.</p></article>`, false},
 }
 for _,tc:=range cases {t.Run(tc.name,func(t *testing.T){if got:=ExplicitPreview(tc.html);got!=tc.want {t.Fatalf("got %v, want %v",got,tc.want)}})}
}
