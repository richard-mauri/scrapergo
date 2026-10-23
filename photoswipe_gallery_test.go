package main

import "testing"

// The two photo URLs and their shared flex layout are based on the actual
// downloaded BMW markup, not publisher-specific gallery classes.
func TestGenericNestedResponsiveFlexPhotoPair(t *testing.T) {
 base:="https://cdn.arstechnica.net/wp-content/uploads/2026/10/"
 raw:=`<article><figure><img src="`+base+`BMW-iX5-60-xDrive-Tanzanite-Blue_43.jpg"></figure><p>Opening prose</p><div class="unrelated-photo-shell"><div class="flex flex-col flex-nowrap gap-5 py-5 md:flex-row"><div><div class="image-tile relative block h-full w-full"><a class="cursor-zoom-in"><img src="`+base+`BMW-iX5-60-xDrive-Pine-Grey_23-1024x683.jpg"></a></div></div><div class="flex-1"><div class="image-tile relative block h-full w-full"><a class="cursor-zoom-in"><img src="`+base+`BMW-iX5-60-xDrive-Pine-Grey_34-1024x683.jpg"></a></div></div></div></div><p>Later prose</p></article>`
 a:=Article{Blocks:[]Block{{Kind:"img",Role:"lead",URL:base+"BMW-iX5-60-xDrive-Tanzanite-Blue_43.jpg"},{Kind:"p",Text:"Before gallery"},{Kind:"img",URL:base+"BMW-iX5-60-xDrive-Pine-Grey_23-1024x683.jpg"},{Kind:"p",Text:"Intervening recovered paragraph"},{Kind:"img",URL:base+"BMW-iX5-60-xDrive-Pine-Grey_34-1024x683.jpg"}}}
 recovery.GroupArticleGalleryImages(&a,raw,"https://example.org/article")
 if len(a.Blocks)!=5 || a.Blocks[0].Role!="lead" {t.Fatalf("lead or article blocks changed: %+v",a.Blocks)}
 if a.Blocks[2].Gallery==""||a.Blocks[3].Gallery!=a.Blocks[2].Gallery {t.Fatalf("responsive flex layout not restored: %+v",a.Blocks)}
 if a.Blocks[4].Kind!="p" {t.Fatalf("paragraph lost: %+v",a.Blocks)}
}

func TestGenericGridPhotos(t *testing.T) {
 raw:=`<div class="grid grid-cols-2"><div><img src="/a.jpg"></div><div><img src="/b.jpg"></div></div>`
 a:=Article{Blocks:[]Block{{Kind:"img",URL:"https://example.org/a.jpg"},{Kind:"img",URL:"https://example.org/b.jpg"}}}
 recovery.GroupArticleGalleryImages(&a,raw,"https://example.org/news")
 if a.Blocks[0].Gallery==""||a.Blocks[1].Gallery!=a.Blocks[0].Gallery {t.Fatalf("grid pair not grouped: %+v",a.Blocks)}
}

func TestEditorialFlexContainerWithProseNotGallery(t *testing.T) {
 raw:=`<div class="flex md:flex-row"><div><img src="/a.jpg"></div><div><p>Article prose</p><img src="/b.jpg"></div></div>`
 a:=Article{Blocks:[]Block{{Kind:"img",URL:"https://example.org/a.jpg"},{Kind:"p",Text:"Article prose"},{Kind:"img",URL:"https://example.org/b.jpg"}}}
 recovery.GroupArticleGalleryImages(&a,raw,"https://example.org/news")
 if a.Blocks[0].Gallery!=""||a.Blocks[2].Gallery!="" {t.Fatalf("editorial flex misclassified: %+v",a.Blocks)}
}

func TestSingleImageLayoutNotGallery(t *testing.T) {
 raw:=`<div class="flex md:flex-row"><div><img src="/a.jpg"></div></div>`
 a:=Article{Blocks:[]Block{{Kind:"img",URL:"https://example.org/a.jpg"}}}
 recovery.GroupArticleGalleryImages(&a,raw,"https://example.org/news")
 if a.Blocks[0].Gallery!="" {t.Fatal("single image grouped")}
}
