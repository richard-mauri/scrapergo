// Package article defines ScraperGo's transport-independent canonical article data.
// Extraction and rendering code should depend on these types rather than on HTTP.
package article

// InlineLink is a link recovered from article text.
type InlineLink struct {
 Text string `json:"text"`
 URL string `json:"url"`
}

// TextPart represents a segment of an article block, retaining inline formatting.
type TextPart struct {
 Text string `json:"text"`
 Bold bool `json:"bold,omitempty"`
 Italic bool `json:"italic,omitempty"`
 URL string `json:"url,omitempty"`
}

// Block is a paragraph, heading or media item in reading order.
type Block struct {
 Parts []TextPart `json:"parts,omitempty"`
 Kind string `json:"kind"`
 Role string `json:"role,omitempty"`
 Gallery string `json:"gallery,omitempty"`
 Links []InlineLink `json:"links,omitempty"`
 Text string `json:"text"`
 URL string `json:"url,omitempty"`
}

// Article is ScraperGo's canonical, transport-independent article document.
type Article struct {
 Title string `json:"title"`
 Byline string `json:"byline,omitempty"`
 Published string `json:"published,omitempty"`
 Site string `json:"site,omitempty"`
 URL string `json:"url"`
 Blocks []Block `json:"blocks"`
 WordCount int `json:"wordCount"`
 ImageURL string `json:"imageUrl,omitempty"`
 Method string `json:"method,omitempty"`
 AccessNote string `json:"accessNote,omitempty"`
 AccessType string `json:"accessType,omitempty"`
}
