package extraction

import (


	"net/url"
	"testing"
)

func TestSerializedCaptionImagesGeneric(t *testing.T) {
	base, _ := url.Parse("https://example.org/articles/guide")
	raw := `<script type="application/json">["https://cdn.example.org/media/photo.webp","A detailed caption about the main illustration in this technical guide."]</script>`
	got := SerializedCaptionImages(raw, base)
	caption := "A detailed caption about the main illustration in this technical guide."
	if got[caption] != "https://cdn.example.org/media/photo.webp" {
		t.Fatalf("image/caption association lost: %v", got)
	}
}

