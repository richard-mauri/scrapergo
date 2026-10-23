package main

import (
 "bytes"
 "clean-article-reader/internal/ui"
	"strings"
	"testing"
)

func TestVersionSingleSource(t *testing.T) {
	html := ui.Render(appVersion,appBuild)
	if !strings.Contains(html, "ScraperGo · v"+appVersion) {
		t.Fatalf("header version differs from appVersion %s", appVersion)
	}
	if !strings.Contains(html, appBuild) || !strings.Contains(appDisplayVersion(), appBuild) {
		t.Fatal("build label missing from UI or executable version")
	}
	if strings.Contains(html, "__APP_VERSION__") {
		t.Fatal("unexpanded version placeholder")
	}
	if strings.Contains(ui.Template(), "ScraperGo · v0.") {
		t.Fatal("hardcoded version in template")
	}
}

func TestTannerHeader(t *testing.T) {
	if !strings.Contains(ui.Template(), `src="/tanner.jpg"`) || !strings.Contains(ui.Template(), `<span>Tanner</span>`) {
		t.Fatal("missing Tanner portrait in header")
	}
	if !bytes.HasPrefix(tannerJPG, []byte{0xff, 0xd8, 0xff}) {
		t.Fatal("embedded Tanner photo is not JPEG")
	}
}
