package main

import "clean-article-reader/internal/accesspolicy"

// Legacy regression fixtures continue to exercise the new package owner.
func accessNotice(kind string) string { return accesspolicy.Notice(kind) }
func restrictionSource(raw string) string { return accesspolicy.Source(raw) }
func accessRestriction(raw string) string { return accesspolicy.Restriction(raw) }
func restrictionEvidence(raw string) string { return accesspolicy.Evidence(raw) }
func paywallEvidence(raw string) bool { return accesspolicy.IsPaywallEvidence(raw) }
func mustHostname(raw string) string { return accesspolicy.Hostname(raw) }
