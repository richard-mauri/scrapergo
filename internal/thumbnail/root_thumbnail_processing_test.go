package thumbnail_test

import (
 "clean-article-reader/internal/thumbnail"
 "bytes"
 "image"
 "image/color"
 "image/jpeg"
 "strconv"
 "strings"
 "testing"
)

func TestThumbnailResizeDimensions(t *testing.T) {
	source := image.NewRGBA(image.Rect(0, 0, 1920, 1080))
	for y := 0; y < 1080; y++ {
		for x := 0; x < 1920; x++ {
			source.Set(x, y, color.RGBA{R: 100, G: 140, B: 200, A: 255})
		}
	}
	var in bytes.Buffer
	if err := jpeg.Encode(&in, source, nil); err != nil {
		t.Fatal(err)
	}
	out, err := thumbnail.CropAndEncode(in.Bytes(), 256)
	if err != nil {
		t.Fatal(err)
	}
	cfg, format, err := image.DecodeConfig(bytes.NewReader(out))
	if err != nil {
		t.Fatal(err)
	}
	if format != "jpeg" || cfg.Width != 256 || cfg.Height != 176 {
		t.Fatalf("wrong thumbnail: %s %dx%d", format, cfg.Width, cfg.Height)
	}
}

func TestThumbnailCache(t *testing.T) {
	store := thumbnail.NewStore()
	store.Put("one", []byte("valid"))
	v, ok := store.Get("one")
	if !ok || string(v) != "valid" {
		t.Fatal("cache round-trip")
	}
	for i := 0; i < thumbnail.CacheEntries+2; i++ {
		store.Put(strconv.Itoa(i), []byte{1})
	}
	if store.Len() != thumbnail.CacheEntries {
		t.Fatal("cache capacity exceeded")
	}
}


func TestThumbnailSourceType(t *testing.T) {
	cases := []struct {
		data []byte
		want string
	}{
		{[]byte("RIFF\x00\x00\x00\x00WEBPVP8 "), "webp"},
		{[]byte("\x00\x00\x00\x18ftypavif"), "avif"},
		{[]byte("nonsense"), "unknown"},
	}
	for _, tc := range cases {
		if got := thumbnail.SourceType(tc.data); got != tc.want {
			t.Fatalf("got %q; want %q", got, tc.want)
		}
	}
}

func TestUnsupportedThumbnailFormatExplained(t *testing.T) {
	_, err := thumbnail.CropAndEncode([]byte("\x00\x00\x00\x18ftypavif"), 256)
	if err == nil || !strings.Contains(err.Error(), "avif") {
		t.Fatalf("expected AVIF diagnostic, got %v", err)
	}
}
