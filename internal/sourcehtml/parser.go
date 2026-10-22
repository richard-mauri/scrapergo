// Package sourcehtml owns ScraperGo's legacy parsed-source tree and helpers.
// Parsing behavior is preserved as a compatibility boundary for staged migration.
package sourcehtml

import (
 "html"
 "regexp"
 "strings"
)

type Node struct {
	Tag    string
	Attr   map[string]string
	Parent *Node
	Kids   []*Node
	Text   string
}

var tagRE = regexp.MustCompile(`(?s)<!--.*?-->|<![^>]*>|<[^>]*>`)
var attrRE = regexp.MustCompile(`(?i)([a-zA-Z_:][-a-zA-Z0-9_:.]*)\s*=\s*(?:"([^"]*)"|'([^']*)'|([^\s>]+))`)
var wsRE = regexp.MustCompile(`\s+`)
var voidTags = map[string]bool{"meta": true, "link": true, "img": true, "input": true, "br": true, "hr": true, "source": true, "area": true, "base": true, "embed": true, "wbr": true}
var BlockedTags = map[string]bool{"script": true, "style": true, "noscript": true, "iframe": true, "svg": true, "form": true, "button": true, "nav": true, "footer": true, "aside": true, "header": true}

func Normalize(s string) string {
	return strings.TrimSpace(wsRE.ReplaceAllString(html.UnescapeString(s), " "))
}
func Parse(raw string) *Node {
	root := &Node{Tag: "root"}
	stack := []*Node{root}
	last := 0
	for _, loc := range tagRE.FindAllStringIndex(raw, -1) {
		if loc[0] > last {
			stack[len(stack)-1].Kids = append(stack[len(stack)-1].Kids, &Node{Tag: "#text", Text: raw[last:loc[0]], Parent: stack[len(stack)-1]})
		}
		tok := raw[loc[0]:loc[1]]
		last = loc[1]
		if strings.HasPrefix(tok, "<!--") || strings.HasPrefix(tok, "<!") {
			continue
		}
		if strings.HasPrefix(tok, "</") {
			tag := strings.ToLower(strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(tok, "</"), ">")))
			fields := strings.Fields(tag)
			if len(fields) == 0 {
				continue
			}
			tag = fields[0]
			for j := len(stack) - 1; j > 0; j-- {
				if stack[j].Tag == tag {
					stack = stack[:j]
					break
				}
			}
			continue
		}
		body := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(tok, "<"), ">"))
		if body == "" {
			continue
		}
		parts := strings.Fields(body)
		if len(parts) == 0 {
			continue
		}
		tag := strings.ToLower(strings.TrimRight(parts[0], "/"))
		if !regexp.MustCompile(`^[a-z][a-z0-9:-]*$`).MatchString(tag) {
			continue
		}
		p := stack[len(stack)-1]
		node := &Node{Tag: tag, Parent: p, Attr: map[string]string{}}
		for _, m := range attrRE.FindAllStringSubmatch(body, -1) {
			val := m[2]
			if val == "" {
				val = m[3]
			}
			if val == "" {
				val = m[4]
			}
			node.Attr[strings.ToLower(m[1])] = html.UnescapeString(val)
		}
		p.Kids = append(p.Kids, node)
		if !voidTags[tag] && !strings.HasSuffix(body, "/") {
			stack = append(stack, node)
		}
	}
	if last < len(raw) {
		p := stack[len(stack)-1]
		p.Kids = append(p.Kids, &Node{Tag: "#text", Text: raw[last:], Parent: p})
	}
	return root
}
func Find(root *Node, test func(*Node) bool) []*Node {
	var out []*Node
	var walk func(*Node)
	walk = func(n *Node) {
		if test(n) {
			out = append(out, n)
		}
		for _, k := range n.Kids {
			walk(k)
		}
	}
	walk(root)
	return out
}
func Text(n *Node) string {
	if n == nil {
		return ""
	}
	if n.Tag == "#text" {
		return n.Text
	}
	if BlockedTags[n.Tag] {
		return ""
	}
	var b strings.Builder
	for _, c := range n.Kids {
		b.WriteString(Text(c))
		b.WriteByte(' ')
	}
	return Normalize(b.String())
}
func Meta(root *Node, keys ...string) string {
	for _, n := range Find(root, func(n *Node) bool { return n.Tag == "meta" }) {
		key := strings.ToLower(n.Attr["property"])
		if key == "" {
			key = strings.ToLower(n.Attr["name"])
		}
		for _, wanted := range keys {
			if key == wanted && n.Attr["content"] != "" {
				return Normalize(n.Attr["content"])
			}
		}
	}
	return ""
}
