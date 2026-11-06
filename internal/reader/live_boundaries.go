package reader

import (
	"clean-article-reader/internal/article"
	"net/url"
	"regexp"
	"strings"
)

// CBS-style Readability output can append a live update's relative time to
// the previous update's last paragraph and flatten the following headline.
// Recover the boundary from the sequence, without removing editorial text.
var liveTimeSuffix = regexp.MustCompile(`(?i)(?:^|\s)(\d{1,3}\s*(?:m|h|d|w)\s+ago|just now)$`)

func restoreFlattenedLiveUpdates(a *article.Article) {
	u, err := url.Parse(a.URL)
	if err != nil || !(strings.Contains(u.Path, "/live/") || strings.Contains(u.Path, "/live-updates/")) {
		return
	}
	var result []article.Block
	for i, b := range a.Blocks {
		// Require a following headline-shaped paragraph AND its body copy.
		if b.Kind != "p" || i+2 >= len(a.Blocks) || !liveHeadlineCandidate(a.Blocks[i+1], a.Blocks[i+2]) {
			result = append(result, b)
			continue
		}
		text := strings.TrimSpace(b.Text)
		match := liveTimeSuffix.FindStringSubmatchIndex(text)
		if match == nil {
			result = append(result, b)
			continue
		}
		timeText := strings.TrimSpace(text[match[2]:match[3]])
		prefix := strings.TrimSpace(text[:match[0]])
		// Never split ordinary sentences such as "It happened 56m ago".
		// The preceding editorial paragraph must be complete.
		if prefix != "" && !strings.ContainsAny(prefix[len(prefix)-1:], ".!?\u201d\"") {
			result = append(result, b)
			continue
		}
		if prefix != "" {
			b.Text = prefix
			// Keep original inline links when changing the text. If the inline
			// segments cannot be safely shortened, fall back to link ranges.
			if len(b.Parts) != 0 {
				joined := ""
				for _, p := range b.Parts { joined += p.Text }
				if joined == text {
					trim := len(joined)-len(prefix)
					for j := len(b.Parts)-1; j >= 0 && trim > 0; j-- {
						n := len(b.Parts[j].Text)
						if n <= trim { b.Parts[j].Text = ""; trim -= n } else {
							b.Parts[j].Text = b.Parts[j].Text[:n-trim]
							trim = 0
						}
					}
				} else { b.Parts = nil }
			}
			result = append(result, b)
		}
		result = append(result, article.Block{Kind:"live-time", Text:timeText})
		// Mark the heading on the following iteration, keeping source order.
		continue
	}
	// Apply heading identity to the paragraph immediately following each time.
	for i := 0; i+1 < len(result); i++ {
		if result[i].Kind == "live-time" && (result[i+1].Kind == "p" || result[i+1].Kind == "h3") {
			result[i+1].Kind = "live-heading"
		}
	}
	a.Blocks = result
}

func liveHeadlineCandidate(head, body article.Block) bool {
	if (head.Kind != "p" && head.Kind != "h3") || body.Kind != "p" {
		return false
	}
	label := strings.TrimSpace(head.Text)
	if len(label) < 20 || len(label) > 220 || strings.Contains(label, "\n") || strings.HasSuffix(label, ".") {
		return false
	}
	return len(strings.TrimSpace(body.Text)) >= 45
}
