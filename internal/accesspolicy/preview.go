package accesspolicy

import "strings"

// ExplicitPreview identifies a publisher's affirmative description of the
// delivered page as a preview. Subscription metadata alone is insufficient.
// Source excludes scripts, styles and HTML comments to avoid matching phrases
// embedded only in application code.
func ExplicitPreview(raw string) bool {
 source := Source(raw)
 return strings.Contains(source, "you have a preview view of this article while we are checking your access")
}
