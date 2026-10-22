package main

import (
 "strings"
 "testing"
)

func TestSourceGalleryPreservesBMWImagePair(t *testing.T) {
 base:="https://example.org/story"
 raw:=`<article><figure><img src="/lead.jpg" alt="Lead"></figure><p>Opening prose in the article body.</p><div class="photo-gallery"><figure><img src="/driving-front.jpg" alt="Front angle"></figure><figure><img src="/driving-rear.jpg" alt="Rear angle"></figure></div><p>Later prose.</p></article>`
 a:=Article{ImageURL:"https://example.org/lead.jpg",Blocks:[]Block{
  {Kind:"img",Role:"lead",URL:"https://example.org/lead.jpg"},
  {Kind:"p",Text:"Opening prose in the article body."},
  {Kind:"img",URL:"https://example.org/driving-front.jpg",Text:"Front angle"},
  {Kind:"p",Text:"Paragraph inserted by text recovery."},
  {Kind:"img",URL:"https://example.org/driving-rear.jpg",Text:"Rear angle"},
  {Kind:"p",Text:"Later prose."},
 }}
 groupArticleGalleryImages(&a,raw,base)
 if len(a.Blocks)!=6 {t.Fatalf("block count changed: %+v",a.Blocks)}
 if a.Blocks[0].Role!="lead" || a.Blocks[2].Gallery=="" || a.Blocks[2].Gallery!=a.Blocks[3].Gallery {t.Fatalf("gallery not restored and lead not preserved: %+v",a.Blocks)}
 if !strings.Contains(a.Blocks[3].URL,"driving-rear") || a.Blocks[4].Kind!="p" {t.Fatalf("gallery split by paragraph: %+v",a.Blocks)}
}

func TestIndependentFiguresNotGrouped(t *testing.T) {
 raw:=`<article><figure><img src="/a.jpg"/></figure><p>Independent text</p><figure><img src="/b.jpg"/></figure></article>`
 a:=Article{Blocks:[]Block{{Kind:"img",URL:"https://example.org/a.jpg"},{Kind:"p",Text:"Independent text"},{Kind:"img",URL:"https://example.org/b.jpg"}}}
 groupArticleGalleryImages(&a,raw,"https://example.org/story")
 if len(a.Blocks)!=3 || a.Blocks[0].Gallery!="" || a.Blocks[2].Gallery!="" {t.Fatalf("unrelated images grouped: %+v",a.Blocks)}
}

func TestGalleryRequiresTwoPresentImages(t *testing.T) {
 raw:=`<article><div class="gallery"><img src="/one.jpg"/><img src="/two.jpg"/></div></article>`
 a:=Article{Blocks:[]Block{{Kind:"img",URL:"https://example.org/one.jpg"}}}
 groupArticleGalleryImages(&a,raw,"https://example.org/story")
 if a.Blocks[0].Gallery!="" {t.Fatalf("singleton gallery: %+v",a.Blocks)}
}

func TestVerticalInlineFlexDoesNotGroupEditorialImages(t *testing.T) {
 raw:=`<article><div style="display: flex; flex-direction: column"><div><img src="/a.jpg"></div><div><img src="/b.jpg"></div></div></article>`
 a:=Article{Blocks:[]Block{{Kind:"img",URL:"https://example.org/a.jpg"},{Kind:"p",Text:"Intervening text"},{Kind:"img",URL:"https://example.org/b.jpg"}}}
 groupArticleGalleryImages(&a,raw,"https://example.org/story")
 if a.Blocks[0].Gallery!=""||a.Blocks[2].Gallery!="" {t.Fatalf("vertical layout should preserve separate editorial figures: %+v",a.Blocks)}
}

func TestHorizontalInlineFlexGroupsPhotographs(t *testing.T) {
 raw:=`<article><div style="display: flex; flex-direction: row"><div><img src="/a.jpg"></div><div><img src="/b.jpg"></div></div></article>`
 a:=Article{Blocks:[]Block{{Kind:"img",URL:"https://example.org/a.jpg"},{Kind:"p",Text:"Recovered paragraph"},{Kind:"img",URL:"https://example.org/b.jpg"}}}
 groupArticleGalleryImages(&a,raw,"https://example.org/story")
 if a.Blocks[0].Gallery==""||a.Blocks[1].Gallery!=a.Blocks[0].Gallery {t.Fatalf("horizontal photo layout not restored: %+v",a.Blocks)}
}
