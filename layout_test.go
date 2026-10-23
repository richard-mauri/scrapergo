package main

import (
 "clean-article-reader/internal/ui"
	"strings"
	"testing"
)

func TestDesktopReaderWidth(t *testing.T) {
	if !strings.Contains(ui.Template(), "width:calc(100% - 48px);max-width:1200px") {
		t.Fatal("centered desktop reader width missing")
	}
	if !strings.Contains(ui.Template(), "width:calc(100% - 24px)") {
		t.Fatal("mobile reader width missing")
	}
}
