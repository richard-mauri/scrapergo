package httpapi

// Publisher access restrictions and content completeness are independent.
// Neither accessType nor accessNote establishes that extracted editorial
// content was truncated. AccessNote is a generated caution, not source
// evidence. Until there is a distinct source-backed truncation signal,
// classification must not label a successfully read article as partial.
func classifyContentStatus(accessType, accessNote string) (string, string) {
 return "unknown", ""
}
