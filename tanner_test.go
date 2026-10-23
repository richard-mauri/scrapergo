package main

import (
 "clean-article-reader/internal/ui"
	"bytes"
	"strings"
	"testing"
)

func TestTannerHeader(t *testing.T) {
	if !strings.Contains(ui.Template(), `src="/tanner.jpg"`) || !strings.Contains(ui.Template(), `<span>Tanner</span>`) {
		t.Fatal("missing Tanner portrait in header")
	}
	if !bytes.HasPrefix(tannerJPG, []byte{0xff, 0xd8, 0xff}) {
		t.Fatal("embedded Tanner photo is not JPEG")
	}
}
