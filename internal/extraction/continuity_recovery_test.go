package extraction

import (
 articlemodel "clean-article-reader/internal/article"
 "clean-article-reader/internal/sourcehtml"
"net/url"
	"strings"
	"testing"
)

func TestEditorialContinuityRestoresGalleryAndFinalParagraphs(t *testing.T) {
	source := `<html><body><main><article><h1>Slate showroom</h1>
 <p>As I sat in the cab of the pickup truck and manually rolled down the windows, I thought about driving an old truck years ago.</p>
 <div class="advertisement"><p>Advertisement promoted deals that are not part of this editorial article.</p></div>
 <p>Being able to order online and have the truck delivered is a major attraction for buyers and shoppers who dislike dealerships.</p>
 <div class="gallery"><figure><img src="/pegboard.jpg" alt="Lights"/><figcaption>All the headlight and taillight options for the Slate Truck are on display.</figcaption></figure>
 <figure><img src="/models.jpg" alt="Interiors"/><figcaption>Miniature interior models help customers visualize what their Slate Truck will look like.</figcaption></figure></div>
 <p>Purdy told me that Slate is staying the course with the advertised sticker price and that deliveries are planned for next year.</p>
 <p>For service, Slate will be using <a href="https://repairpal.com/">RepairPal</a> to connect owners with shops able to maintain the trucks.</p>
 <p>Monday morning is not exactly prime time for dealership traffic, but a handful of prospective customers were in the showroom.</p>
 <p>But the sleek-looking showroom and an attractive prototype still face a production reality check. Now it is time to deliver.</p>
 </article></main></body></html>`
	base, _ := url.Parse("https://example.com/slate")
	selected := []articlemodel.Block{
		{Kind: "p", Text: "As I sat in the cab of the pickup truck and manually rolled down the windows, I thought about driving an old truck years ago."},
		{Kind: "p", Text: "Being able to order online and have the truck delivered is a major attraction for buyers and shoppers who dislike dealerships."},
	}
	got := RecoverEditorialContinuity(source, selected, base)
	if len(got) != 8 {
		t.Fatalf("want 8 article blocks, got %d: %+v", len(got), got)
	}
	if got[2].Kind != "img" || got[2].URL != "https://example.com/pegboard.jpg" || !strings.Contains(got[2].Text, "headlight") {
		t.Fatalf("first gallery image/caption missing: %+v", got[2])
	}
	if got[3].Kind != "img" || got[3].URL != "https://example.com/models.jpg" || !strings.Contains(got[3].Text, "Miniature") {
		t.Fatalf("second gallery image/caption missing: %+v", got[3])
	}
	if !strings.Contains(got[7].Text, "time to deliver") {
		t.Fatalf("final paragraph missing: %+v", got[7])
	}
	if strings.Contains(strings.Join(func() []string {
		a := []string{}
		for _, b := range got {
			a = append(a, b.Text)
		}
		return a
	}(), " "), "promoted deals") {
		t.Fatal("advertisement retained")
	}
	found := false
	for _, part := range got[5].Parts {
		if part.URL == "https://repairpal.com/" {
			found = true
		}
	}
	if !found {
		t.Fatal("RepairPal inline link lost")
	}
}
func TestEditorialContinuityRejectsUnverifiedContent(t *testing.T) {
	base, _ := url.Parse("https://example.org/a")
	raw := `<html><article><p>Unrelated first paragraph covering material absent from the selected article altogether and should not pass the anchoring test.</p><p>Unrelated second paragraph also covers different issues and is not part of the selected article for this test.</p><p>Unrelated third paragraph with sufficient content to make recovery appear tempting but must not be used.</p></article></html>`
	chosen := []articlemodel.Block{{Kind: "p", Text: "Selected article first paragraph with sufficient material to serve as an editorial reference for continuity."}, {Kind: "p", Text: "Selected article second paragraph with sufficient material for the required independent corroboration."}}
	got := RecoverEditorialContinuity(raw, chosen, base)
	if len(got) != len(chosen) || got[0].Text != chosen[0].Text {
		t.Fatalf("replaced unrelated article: %+v", got)
	}
}

func TestEditorialContinuityMixedWrapperAndCaptionOnce(t *testing.T) {
	base, _ := url.Parse("https://example.org/story")
	raw := `<html><main><h1>Article</h1>
<p>First editorial paragraph contains enough distinctive information to establish that it belongs to this long article, and matches the reader selection exactly.</p>
<div class="advertisement-wrapper">
<div class="advertisement"><p>Sponsored offers and deals that should not appear in reading view.</p></div>
<p>Second editorial paragraph contains distinct narrative substance confirming its order within the selected article despite the surrounding advertising layout.</p>
<p>Third editorial paragraph is a continuation within the same mixed wrapper and should not disappear when advertisement wrappers are filtered.</p>
<figure><img src="/gallery.jpg" alt="model"/><figcaption>Miniature interior models help customers visualize their Slate Truck.</figcaption></figure>
</div>
<p>Closing editorial paragraph provides the final production observations and states that now it is time to deliver to customers.</p></main></html>`
	selected := []articlemodel.Block{
		{Kind: "p", Text: "First editorial paragraph contains enough distinctive information to establish that it belongs to this long article, and matches the reader selection exactly."},
		{Kind: "p", Text: "Second editorial paragraph contains distinct narrative substance confirming its order within the selected article despite the surrounding advertising layout."},
	}
	got := RecoverEditorialContinuity(raw, selected, base)
	var captions, images int
	for _, b := range got {
		if b.Kind == "img" {
			images++
			if !strings.Contains(b.Text, "Miniature interior models") {
				t.Fatalf("missing caption: %+v", b)
			}
		} else if strings.Contains(b.Text, "Miniature interior models") {
			captions++
		}
		if strings.Contains(b.Text, "Sponsored offers") {
			t.Fatalf("advertisement retained: %+v", b)
		}
	}
	thirdPresent := false
	for _, b := range got {
		if strings.Contains(b.Text, "Third editorial paragraph is a continuation") {
			thirdPresent = true
		}
	}
	if images != 1 || captions != 0 || !thirdPresent || !strings.Contains(got[len(got)-1].Text, "time to deliver") {
		t.Fatalf("gallery or final content omitted, or duplicate caption emitted: %+v", got)
	}
}

func TestEditorialTailRecoveryAfterSelectedParagraph(t *testing.T) {
	base, _ := url.Parse("https://example.com/article")
	a := "The customization section discusses choosing the interior configuration and describes the truck options available to customers in this showroom."
	raw := `<html><main><article><p>` + a + `</p><div class="advertisement-wrapper"><div class="advertisement"><p>Sponsored deals for everyone.</p></div><p>The next editorial paragraph explains how the staff can assist people who wish to customize their vehicle in person at the center.</p><figure><img src="/pegboard.jpg"/><figcaption>Headlight and taillight options on display at the showroom.</figcaption></figure></div><p>The service discussion explains maintenance arrangements with a network of shops for future truck buyers.</p><p>The final production paragraph examines the manufacturing challenges and concludes it is time to deliver vehicles.</p></article></main></html>`
	chosen := []articlemodel.Block{{Kind: "p", Text: a}}
	got := extraction.RecoverAnchoredEditorialTail(sourcehtml.Parse(raw), chosen, base)
	if len(got) != 5 {
		t.Fatalf("want five blocks, got %d: %+v", len(got), got)
	}
	if got[2].Kind != "img" || got[2].URL != "https://example.com/pegboard.jpg" {
		t.Fatalf("gallery missing: %+v", got[2])
	}
	if !strings.Contains(got[4].Text, "time to deliver") {
		t.Fatalf("ending missing: %+v", got[4])
	}
}
