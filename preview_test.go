package main

import (
	"strings"
	"testing"
)

func TestAccessNotices(t *testing.T) {
	for _, kind := range []string{"registration", "paywall", "restricted"} {
		msg := accessNotice(kind)
		if !strings.Contains(strings.ToLower(msg), "may") {
			t.Fatalf("notice %q promises too much", msg)
		}
	}
}
func TestArticlePreviewJSONFields(t *testing.T) {
	a := Article{AccessType: "restricted", AccessNote: accessNotice("restricted")}
	if a.AccessType != "restricted" || a.AccessNote == "" {
		t.Fatal("preview metadata missing")
	}
}
