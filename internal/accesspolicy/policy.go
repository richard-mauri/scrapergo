package accesspolicy

import (
 "encoding/json"
 "net/url"
 "regexp"
 "strings"
)

// accessNotice deliberately avoids claiming that a displayed preview is complete.
func Notice(kind string) string {
	switch kind {
	case "registration":
		return "Free registration may be required to continue reading. The text below may be only a preview."
	case "paywall":
		return "A paid subscription may be required for the full article. The text below may be only a preview."
	default:
		return "The publisher marks this article as access-restricted. The text below may be incomplete."
	}
}

// restrictionSource removes executable JavaScript and CSS before inspecting
// human-facing access requirements. UI code can contain paywall phrases as
// messages or examples without the requested page actually being restricted.
// JSON-LD is retained because isAccessibleForFree is legitimate access metadata.
var scriptBlockRE = regexp.MustCompile(`(?is)<script\b([^>]*)>(.*?)</script\s*>`)
var styleBlockRE = regexp.MustCompile(`(?is)<style\b[^>]*>.*?</style\s*>`)
var htmlCommentRE = regexp.MustCompile(`(?s)<!--.*?-->`)

func Source(raw string) string {
	kept := make([]string, 0, 2)
	cleaned := scriptBlockRE.ReplaceAllStringFunc(raw, func(block string) string {
		prefixEnd := strings.IndexByte(block, '>')
		if prefixEnd >= 0 {
			startTag := strings.ToLower(block[:prefixEnd+1])
			body := strings.TrimSpace(strings.TrimSuffix(block[prefixEnd+1:], "</script>"))
			// Some publishers emit a bare JSON object in a script element
			// without a type attribute. Only accept syntactically valid JSON
			// containing access-control properties; never scan JS source text.
			var metadata map[string]json.RawMessage
			bareAccessJSON := !strings.Contains(startTag, "type=") &&
				json.Unmarshal([]byte(body), &metadata) == nil &&
				(metadata["isAccessibleForFree"] != nil || metadata["isaccessibleforfree"] != nil || metadata["paywall"] != nil)
			if strings.Contains(startTag, "application/ld+json") || bareAccessJSON {
				kept = append(kept, body)
			}
		}
		return " "
	})
	cleaned = styleBlockRE.ReplaceAllString(cleaned, " ")
	cleaned = htmlCommentRE.ReplaceAllString(cleaned, " ")
	return strings.ToLower(cleaned + " " + strings.Join(kept, " "))
}

// accessRestriction classifies explicit restriction evidence, never HTTP status alone.
// Free registration wins over generic subscription metadata, as in Axios newsletters.
func Restriction(raw string) string {
	lower := Source(raw)
	for _, marker := range []string{
		"sign up for", "register for free to continue reading", "create a free account to continue reading",
	} {
		if marker == "sign up for" {
			// Avoid calling unrelated newsletter advertisements an article registration wall.
			if strings.Contains(lower, "to continue reading for free") || strings.Contains(lower, "to read the rest for free") {
				return "registration"
			}
			continue
		}
		if strings.Contains(lower, marker) {
			return "registration"
		}
	}
	for _, marker := range []string{"paid subscription required", "subscribe for $", "become a paying subscriber"} {
		if strings.Contains(lower, marker) {
			return "paywall"
		}
	}
	for _, marker := range []string{
		"subscribe to continue reading", "subscribe to read the full article",
		"this article is for subscribers only", "subscriber-only article",
		"subscription required to read", `"isaccessibleforfree":false`,
		`"isaccessibleforfree": false`, `"paywall":true`, `"paywall": true`,
	} {
		if strings.Contains(lower, marker) {
			// A generic metadata or subscribe marker does not establish that payment is required.
			return "restricted"
		}
	}
	return ""
}

// restrictionEvidence explains which predefined phrase caused a restriction decision.
// It returns marker names, not article text, to keep server logs predictable.
func Evidence(raw string) string {
	lower := Source(raw)
	for _, rule := range []struct{ phrase, label string }{
		{"to continue reading for free", "free-reading-phrase"},
		{"to read the rest for free", "free-reading-phrase"},
		{"register for free to continue reading", "free-registration"},
		{"create a free account to continue reading", "free-account"},
		{"paid subscription required", "explicit-paid-subscription"},
		{"subscribe for $", "explicit-price"},
		{"become a paying subscriber", "explicit-paid-subscriber"},
		{"subscribe to continue reading", "subscription-prompt"},
		{"subscribe to read the full article", "subscription-prompt"},
		{"this article is for subscribers only", "subscriber-only"},
		{"subscriber-only article", "subscriber-only"},
		{"subscription required to read", "subscription-required"},
		{`"isaccessibleforfree":false`, "structured-paywall-metadata"},
		{`"isaccessibleforfree": false`, "structured-paywall-metadata"},
		{`"paywall":true`, "paywall-metadata"},
		{`"paywall": true`, "paywall-metadata"},
	} {
		if strings.Contains(lower, rule.phrase) {
			return rule.label
		}
	}
	return "unidentified"
}

// Backward compatibility for tests and callers expecting a boolean restriction signal.
func IsPaywallEvidence(raw string) bool { return Restriction(raw) != "" }
func Hostname(address string) string {
	u, err := url.Parse(address)
	if err != nil {
		return "publisher"
	}
	return u.Hostname()
}

