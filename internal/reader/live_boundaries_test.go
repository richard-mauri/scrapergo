package reader

import (
    "clean-article-reader/internal/article"
    "testing"
)

func TestRestoreFlattenedLiveUpdateAtCBSBoundary(t *testing.T) {
    paragraph := "Iran's Supreme Leader Mojtaba Khamenei has not been seen in public since being wounded in the strikes that killed his father on Feb. 28."
    headline := "Tanker hit by explosion in Strait of Hormuz, U.K. naval agency says"
    body := "A tanker was hit by an explosion Thursday in the Strait of Hormuz, a British navy agency said Friday."
    for _, url := range []string{"https://www.cbsnews.com/live-updates/iran-war/", "https://example.org/live/iran-war/"} {
        a := &article.Article{URL:url, Blocks:[]article.Block{
            {Kind:"p", Text:paragraph+" 56m ago"},
            {Kind:"p", Text:headline},
            {Kind:"p", Text:body},
            {Kind:"p", Text:"The tanker remained under its own power following the explosion."},
        }}
        restoreFlattenedLiveUpdates(a)
        if len(a.Blocks) != 5 {t.Fatalf("%s got %d blocks: %+v",url,len(a.Blocks),a.Blocks)}
        expectedKinds:=[]string{"p","live-time","live-heading","p","p"}
        expectedText:=[]string{paragraph,"56m ago",headline,body,"The tanker remained under its own power following the explosion."}
        for i := range expectedKinds {
            if a.Blocks[i].Kind!=expectedKinds[i] || a.Blocks[i].Text!=expectedText[i] {
                t.Fatalf("%s block %d: got %+v; want %s %q",url,i,a.Blocks[i],expectedKinds[i],expectedText[i])
            }
        }
    }
}

func TestRestoreFlattenedLiveUpdatePreservesLinkParts(t *testing.T) {
    text:="Iran's Supreme Leader was wounded in the strikes that killed his father on Feb. 28."
    a:=&article.Article{URL:"https://example.org/live-updates/story",Blocks:[]article.Block{
        {Kind:"p",Text:text+" 56m ago",Parts:[]article.TextPart{
            {Text:"Iran's Supreme Leader was wounded in the strikes that "},
            {Text:"killed his father",URL:"https://example.org/context"},
            {Text:" on Feb. 28. 56m ago"},
        }},
        {Kind:"p",Text:"Tanker hit by explosion in Strait of Hormuz, U.K. naval agency says"},
        {Kind:"p",Text:"A tanker was hit by an explosion Thursday in the Strait of Hormuz, a British navy agency said Friday."},
    }}
    restoreFlattenedLiveUpdates(a)
    if len(a.Blocks)!=4 || a.Blocks[0].Text!=text || a.Blocks[0].Parts[1].URL!="https://example.org/context" || a.Blocks[0].Parts[2].Text!=" on Feb. 28." {
        t.Fatalf("inline link or suffix split lost: %+v",a.Blocks)
    }
}

func TestRestoreFlattenedLiveUpdatesDeclinesAmbiguous(t *testing.T) {
    for _, tc := range []struct{url string; text string; heading string; body string}{
        {"https://example.org/news","The report ended. 56m ago","Tanker hit by explosion in Strait of Hormuz, naval agency says","A tanker was hit by an explosion Thursday near the coast of Oman."},
        {"https://example.org/live-updates/news","The report happened 56m ago","Tanker hit by explosion in Strait of Hormuz, naval agency says","A tanker was hit by an explosion Thursday near the coast of Oman."},
        {"https://example.org/live-updates/news","The report ended. 56m ago","Ordinary prose ends with a period.","A tanker was hit by an explosion Thursday near the coast of Oman."},
    } {
        a:=&article.Article{URL:tc.url,Blocks:[]article.Block{{Kind:"p",Text:tc.text},{Kind:"p",Text:tc.heading},{Kind:"p",Text:tc.body}}}
        restoreFlattenedLiveUpdates(a)
        if len(a.Blocks)!=3 || a.Blocks[0].Text!=tc.text || a.Blocks[1].Kind!="p" {
            t.Fatalf("false positive split for %+v: %+v",tc,a.Blocks)
        }
    }
}
