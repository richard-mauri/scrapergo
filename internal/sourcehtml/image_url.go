package sourcehtml

import (
 "net/url"
 "strings"
)

// ImageURL resolves valid editorial image sources, preferring lazy-loaded assets.
func ImageURL(n *Node, baseURL string) string {
	if n.Tag != "img" {
		return ""
	}
	// Prefer actual lazy-loaded URLs over transparent placeholder src values.
	candidates := []string{n.Attr["data-src"], n.Attr["data-original"], n.Attr["data-lazy-src"], n.Attr["src"]}
	for _, ss := range []string{n.Attr["data-srcset"], n.Attr["srcset"]} {
		if ss != "" {
			fields := strings.Fields(strings.TrimSpace(strings.Split(ss, ",")[0]))
			if len(fields) > 0 {
				candidates = append(candidates, fields[0])
			}
		}
	}
	base, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}
	for _, raw := range candidates {
		u, err := url.Parse(strings.TrimSpace(raw))
		if err != nil || u.String() == "" {
			continue
		}
		dest := base.ResolveReference(u)
		if (dest.Scheme == "https" || dest.Scheme == "http") && dest.Host != "" && dest.User == nil && !strings.HasSuffix(strings.ToLower(dest.Path), ".svg") {
			return dest.String()
		}
	}
	return ""
}
