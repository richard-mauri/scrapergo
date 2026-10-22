package sourcehtml

import (
"net/url"
"strings"
)

// ResolveURL resolves an article-selected source link without permitting unsafe schemes or userinfo.
func ResolveURL(raw string, base *url.URL) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	resolved := base.ResolveReference(u)
	if resolved.Scheme != "http" && resolved.Scheme != "https" {
		return ""
	}
	if resolved.Hostname() == "" || resolved.User != nil {
		return ""
	}
	return resolved.String()
}

