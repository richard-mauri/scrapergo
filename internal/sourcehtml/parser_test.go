package sourcehtml

import "testing"

func TestParseTreeAndMetadata(t *testing.T) {
 root:=Parse(`<html><head><meta property="og:image" content="/lead.jpg"></head><body><article><p>Alpha <a href="/story">beta</a>.</p><img src="/photo.jpg"></article></body></html>`)
 if got:=Meta(root,"og:image");got!="/lead.jpg" {t.Fatalf("metadata: %q",got)}
 pars:=Find(root,func(n *Node)bool{return n.Tag=="p"})
 if len(pars)!=1 || Text(pars[0])!="Alpha beta ." {t.Fatalf("paragraphs: %+v, text: %q",pars,func()string{if len(pars)>0{return Text(pars[0])};return ""}())}
 images:=Find(root,func(n *Node)bool{return n.Tag=="img"})
 if len(images)!=1 || images[0].Attr["src"]!="/photo.jpg" {t.Fatalf("images: %+v",images)}
}
func TestNormalizeEntities(t *testing.T) {
 if got:=Normalize(" A &amp;  B ");got!="A & B" {t.Fatalf("normalized: %q",got)}
}
