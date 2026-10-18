package main

import "testing"

func TestPaywallEvidence(t *testing.T) {
	for _, tc := range []struct {
		body string
		want bool
	}{{`<script>{"isAccessibleForFree":false}</script>`, true}, {`Subscribe to continue reading`, true}, {`HTTP 403 forbidden`, false}, {`Read our daily news coverage here`, false}} {
		if got := paywallEvidence(tc.body); got != tc.want {
			t.Errorf("%q: %v", tc.body, got)
		}
	}
}

func TestAccessRestriction(t *testing.T) {
	for _, tc := range []struct{ input, want string }{
		{`Sign up for Axios Hill Leaders to continue reading for free.`, "registration"},
		{`Register for free to continue reading`, "registration"},
		{`This article is for subscribers only`, "restricted"},
		{`Paid subscription required`, "paywall"},
		{`HTTP 403 Forbidden`, ""},
		{`Subscribe to our newsletter`, ""},
	} {
		if got := accessRestriction(tc.input); got != tc.want {
			t.Errorf("%q: got %q, want %q", tc.input, got, tc.want)
		}
	}
}

// The application's own access-warning copy is not evidence that its home
// page requires a subscription. This regression uses the actual embedded UI.
func TestAccessWarningInExecutableScriptIsNotPaywall(t *testing.T) {
	if got := accessRestriction(indexHTML); got != "" {
		t.Fatalf("ScraperGo homepage falsely classified as %q (%s)", got, restrictionEvidence(indexHTML))
	}
	for _, body := range []string{
		`<script>const title = "Paid Subscription Required";</script><article>Public content</article>`,
		`<script>const x = "subscribe to continue reading";</script>`,
		`<script>const config = {"isAccessibleForFree":false};</script>`,
		`<style>.x::after{content:"Paid subscription required"}</style>`,
		`<!-- paid subscription required --><article>Public</article>`,
	} {
		if got := accessRestriction(body); got != "" {
			t.Errorf("false positive for executable or non-visible text: %q", got)
		}
	}
}

func TestAccessRestrictionKeepsVisibleEvidenceAndJSONLD(t *testing.T) {
	cases := []struct{ source, kind, evidence string }{
		{`<article><div>Paid Subscription Required</div></article>`, "paywall", "explicit-paid-subscription"},
		{`<script type="application/ld+json">{"isAccessibleForFree":false}</script>`, "restricted", "structured-paywall-metadata"},
		{`<script type="text/javascript">let message='Paid Subscription Required'</script><div>Subscribe to continue reading</div>`, "restricted", "subscription-prompt"},
	}
	for _, tc := range cases {
		if got := accessRestriction(tc.source); got != tc.kind {
			t.Errorf("kind got %q, want %q", got, tc.kind)
		}
		if got := restrictionEvidence(tc.source); got != tc.evidence {
			t.Errorf("evidence got %q, want %q", got, tc.evidence)
		}
	}
}
