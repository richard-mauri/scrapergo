package main

import "testing"

// Source has duplicate lightbox renditions and separate PhotoSwipe slides.
// Only slide siblings are one gallery, never their larger lightbox copies.
func TestPhotoSwipeBMWGallerySlides(t *testing.T) {
 base:="https://cdn.arstechnica.net/wp-content/uploads/2026/10/"
 raw:=`<article><div class="ars-lightbox"><div class="ars-lightbox-item"><a><img src="`+base+`BMW-iX5-60-xDrive-Pine-Grey_23.jpg"/></a></div><div class="ars-lightbox-item"><a><img src="`+base+`BMW-iX5-60-xDrive-Pine-Grey_34.jpg"/></a></div></div><div class="pswp"><div class="pswp__container"><div class="pswp__item"><div class="pswp__zoom-wrap"><img src="`+base+`BMW-iX5-60-xDrive-Pine-Grey_23-1024x683.jpg"/></div></div><div class="pswp__item"><div class="pswp__zoom-wrap"><img src="`+base+`BMW-iX5-60-xDrive-Pine-Grey_34-1024x683.jpg"/></div></div></div></div></article>`
 a:=Article{Blocks:[]Block{{Kind:"img",Role:"lead",URL:base+"BMW-iX5-60-xDrive-Tanzanite-Blue_43.jpg"},{Kind:"p",Text:"Before gallery"},{Kind:"img",URL:base+"BMW-iX5-60-xDrive-Pine-Grey_23-1024x683.jpg"},{Kind:"p",Text:"Readability paragraph between gallery slides"},{Kind:"img",URL:base+"BMW-iX5-60-xDrive-Pine-Grey_34-1024x683.jpg"}}}
 groupArticleGalleryImages(&a,raw,"https://arstechnica.com/cars/article")
 if len(a.Blocks)!=5 || a.Blocks[0].Role!="lead" {t.Fatalf("lead or article blocks changed: %+v",a.Blocks)}
 if a.Blocks[2].Gallery=="" || a.Blocks[3].Gallery!=a.Blocks[2].Gallery {t.Fatalf("PhotoSwipe images not adjacent gallery members: %+v",a.Blocks)}
 if a.Blocks[4].Kind!="p" {t.Fatalf("paragraph not preserved: %+v",a.Blocks)}
}

func TestPhotoSwipeSingleSlideNotGallery(t *testing.T) {
 raw:=`<div class="pswp__container"><div class="pswp__item"><img src="/a.jpg"/></div></div>`
 a:=Article{Blocks:[]Block{{Kind:"img",URL:"https://example.org/a.jpg"}}}
 groupArticleGalleryImages(&a,raw,"https://example.org/story")
 if a.Blocks[0].Gallery!="" {t.Fatal("single slide classified as gallery")}
}
