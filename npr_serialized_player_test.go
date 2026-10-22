package main

import ("testing"
 "clean-article-reader/internal/structured"
)

func TestCapturedNPRSerializedPlayerBlock(t *testing.T) {
 markup := "\n <iframe src=\"https://www.npr.org/player/embed/nx-s1-5977837/nx-s1-9973214\" width=\"100%\" height=\"290\" frameborder=\"0\" scrolling=\"no\" title=\"NPR embedded audio player\">\n "
 player:=Block{Kind:"li",Text:"Embed\n \n \n\n "+markup,Parts:[]TextPart{{Text:"\n \n\n \n \n \n "},{Text:"Embed",Bold:true},{Text:"\n \n \n\n "},{Text:markup,Bold:true},{Text:"\n\n \n "}}}
 caption:="Dried grass in Greenwich Park, London, Aug. 11 during the UK’s hottest summer on record."
 a:=Article{Blocks:[]Block{player,{Kind:"img",Role:"lead",Text:caption,URL:"https://example.org/lead.jpg"},{Kind:"p",Text:caption+" Dan Kitwood/Getty Images Europe hide caption"},{Kind:"p",Text:"Actual editorial paragraph."}}}
 structured.RemoveSerializedPlayerBlocks(&a)
 if len(a.Blocks)!=3 {t.Fatalf("want 3 preserved article blocks, got %+v",a.Blocks)}
 if a.Blocks[0].Kind!="img" || a.Blocks[0].Role!="lead" || a.Blocks[0].Text!=caption {t.Fatal("lead or its caption was changed")}
 if a.Blocks[1].Kind!="p" || a.Blocks[2].Text!="Actual editorial paragraph." {t.Fatal("editorial text lost")}
}
func TestPreserveRealListAndMixedProse(t *testing.T) {
 blocks:=[]Block{{Kind:"li",Text:"Embed this example in your report."},{Kind:"li",Text:"A technical discussion of <iframe src=\"https://example.org/example\"> with valid prose."},{Kind:"li",Text:"An ordinary numbered item."}}
 a:=Article{Blocks:blocks}
 structured.RemoveSerializedPlayerBlocks(&a)
 if len(a.Blocks)!=3 {t.Fatalf("removed non-widget content: %+v",a.Blocks)}
}
