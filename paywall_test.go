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
