package extraction

import (
 "net/url"
 "strings"

 "golang.org/x/net/html"
)

// DiscoverPDFLink locates a publisher-advertised PDF in a public HTML viewer.
// It does not fetch links. The caller must use the validated fetcher and
// enforce its independent redirect, network, size and access restrictions.
func DiscoverPDFLink(raw, source string) string {
 base, err := url.Parse(source)
 if err != nil || (base.Scheme != "http" && base.Scheme != "https") || base.Host == "" { return "" }
 root, err := html.Parse(strings.NewReader(raw))
 if err != nil { return "" }
 var candidate string
 var walk func(*html.Node)
 walk = func(n *html.Node) {
  if candidate != "" { return }
  if n.Type == html.ElementNode {
   tag := strings.ToLower(n.Data)
   if tag == "a" || tag == "iframe" || tag == "embed" || tag == "object" {
    attrName := "href"
    if tag == "iframe" || tag == "embed" { attrName = "src" }
    if tag == "object" { attrName = "data" }
    var ref string
    for _, a := range n.Attr { if strings.EqualFold(a.Key,attrName) { ref = strings.TrimSpace(a.Val); break } }
    if ref != "" {
     u, err := base.Parse(ref)
     if err == nil && u.User == nil && strings.EqualFold(u.Host,base.Host) && u.Scheme == base.Scheme {
      p := strings.ToLower(u.Path)
      declaredPDF := strings.HasSuffix(p,".pdf") || strings.Contains(p,"/article/download/")
      if tag == "a" {
       var label strings.Builder
       var collect func(*html.Node)
       collect = func(x *html.Node) {
        if x.Type == html.TextNode { label.WriteString(x.Data); label.WriteByte(' ') }
        for c := x.FirstChild; c != nil; c = c.NextSibling { collect(c) }
       }
       collect(n)
       declaredPDF = declaredPDF && (strings.Contains(strings.ToLower(label.String()),"pdf") || strings.HasSuffix(p,".pdf"))
      }
      if declaredPDF { u.Fragment = ""; candidate = u.String(); return }
     }
    }
   }
  }
  for c := n.FirstChild; c != nil; c = c.NextSibling { walk(c) }
 }
 walk(root)
 return candidate
}
