package main

import (
 "clean-article-reader/internal/ui"
	"strings"
	"testing"
)

func TestRestrictionEvidenceDiagnostics(t *testing.T) {
	for _, tc := range []struct{ body, evidence, kind string }{
		{`Sign up for Axios Hill Leaders to continue reading for free.`, "free-reading-phrase", "registration"},
		{`<script>{"isAccessibleForFree":false}</script>`, "structured-paywall-metadata", "restricted"},
		{`Subscribe to continue reading`, "subscription-prompt", "restricted"},
		{`HTTP 403 Forbidden`, "unidentified", ""},
	} {
		if got := restrictionEvidence(tc.body); got != tc.evidence {
			t.Errorf("evidence %q: got %q", tc.body, got)
		}
		if got := accessRestriction(tc.body); got != tc.kind {
			t.Errorf("kind %q: got %q", tc.body, got)
		}
	}
	if !strings.Contains(ui.Template(), "paydiagnostic") {
		t.Fatal("publisher diagnostics missing from UI")
	}
}
