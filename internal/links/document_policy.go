package links

import (
 "net/url"
 "strings"
)

// IsDocumentURL recognizes link destinations eligible for inline article links.
func IsDocumentURL(destination string) bool {
	u, err := url.Parse(destination)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Hostname() == "" || u.User != nil {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(u.Path[strings.LastIndex(u.Path, "/")+1:])) {
	case "":
		return true
	}
	name := strings.ToLower(u.Path)
	for _, suffix := range []string{".pdf", ".zip", ".gz", ".png", ".jpg", ".jpeg", ".gif", ".webp", ".svg", ".mp4", ".mp3", ".mov", ".avi", ".doc", ".docx", ".xls", ".xlsx"} {
		if strings.HasSuffix(name, suffix) {
			return false
		}
	}
	return true
}

