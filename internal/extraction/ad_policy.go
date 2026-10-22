package extraction

import "regexp"

var sentenceAdRE = regexp.MustCompile(`(?i)(enter your email|free report|stocks to buy right now|our analysts combed|click here to|grab your free|sign up for|subscribe to|advertisement|read more:|get our free|limited.time offer)`)

// IsAdvertisingText reports the existing publisher-neutral solicitation patterns.
func IsAdvertisingText(s string) bool { return sentenceAdRE.MatchString(s) }
