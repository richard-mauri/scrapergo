package ui

import (

	"strings"
	"testing"
)

func TestDesktopReaderWidth(t *testing.T) {
	if !strings.Contains(Template(), `.suite-banner,.suite-content{box-sizing:border-box;width:min(1180px,calc(100vw - 96px))`) {
		t.Fatal("shared desktop banner/content width missing")
	}
	if !strings.Contains(Template(), ".suite-banner,.suite-content{width:calc(100vw - 36px)}") {
		t.Fatal("shared mobile banner/content width missing")
	}
}
