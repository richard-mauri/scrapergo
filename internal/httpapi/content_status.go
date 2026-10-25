package httpapi

import "strings"

// Unknown is the default: successful extraction alone cannot establish
// whether an article was delivered in full.
func classifyContentStatus(accessType, accessNote string) (string, string) {
 switch strings.ToLower(strings.TrimSpace(accessType)) {
 case "registration", "paywall", "restricted":
  return "partial", "publisher_access_restricted"
 }
 note := strings.ToLower(accessNote)
 if strings.Contains(note, "may be only a preview") || strings.Contains(note, "may be incomplete") {
  return "partial", "publisher_access_restricted"
 }
 return "unknown", ""
}
