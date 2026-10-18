package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestTannerHeader(t *testing.T) {
	if !strings.Contains(indexHTML, `src="/tanner.jpg"`) || !strings.Contains(indexHTML, `<span>Tanner</span>`) {
		t.Fatal("missing Tanner portrait in header")
	}
	if !bytes.HasPrefix(tannerJPG, []byte{0xff, 0xd8, 0xff}) {
		t.Fatal("embedded Tanner photo is not JPEG")
	}
}
