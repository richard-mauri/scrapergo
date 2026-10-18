package main

import (
	"strings"
	"testing"
)

func TestDesktopReaderWidth(t *testing.T) {
	if !strings.Contains(indexHTML, "width:calc(100vw - 48px)") {
		t.Fatal("desktop reader width missing")
	}
	if !strings.Contains(indexHTML, "width:calc(100% - 24px)") {
		t.Fatal("mobile reader width missing")
	}
}
