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
