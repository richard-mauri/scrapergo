// Package googlenews implements Google News article identity and RPC protocol decoding.
// It does not fetch resources or validate publisher destinations.
package googlenews

import (
 "encoding/base64"
 "encoding/json"
 "errors"
 "fmt"
 "net/url"
 "regexp"
 "strconv"
 "strings"
)

// Google's News RPC is undocumented. Keep this resolver separate from article
// extraction so a Google change produces a useful error, not a 0-word article.
var NewsIDRE = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
var NewsSignatureRE = regexp.MustCompile(`data-n-a-sg=["']([^"']+)["']`)
var NewsTimestampRE = regexp.MustCompile(`data-n-a-ts=["']([0-9]+)["']`)
var legacyNewsURLRE = regexp.MustCompile(`https?://[^\x00-\x20"'<>]+`)

func ArticleID(address string) (string, bool, error) {
	u, err := url.Parse(strings.TrimSpace(address))
	if err != nil {
		return "", false, err
	}
	if u.Hostname() != "news.google.com" {
		return "", false, nil
	}
	path := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(path) < 2 || (path[len(path)-2] != "read" && path[len(path)-2] != "articles") {
		return "", false, nil
	}
	id := path[len(path)-1]
	if len(id) < 12 || len(id) > 2048 || !NewsIDRE.MatchString(id) {
		return "", true, errors.New("invalid Google News article identifier")
	}
	return id, true, nil
}

func LegacyPublisherURL(id string) string {
	bytes, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(id, "="))
	if err != nil {
		return ""
	}
	for _, candidate := range legacyNewsURLRE.FindAllString(string(bytes), -1) {
		candidate = strings.TrimRight(candidate, "\\.,;)")
		u, e := url.Parse(candidate)
		if e == nil && (u.Scheme == "https" || u.Scheme == "http") && u.Hostname() != "news.google.com" {
			return candidate
		}
	}
	return ""
}

func RPCPayload(id, timestamp, signature string) (url.Values, error) {
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil || ts <= 0 {
		return nil, errors.New("invalid Google News timestamp")
	}
	inner := []any{"garturlreq", []any{[]any{"X", "X", []any{"X", "X"}, nil, nil, 1, 1, "US:en", nil, 1, nil, nil, nil, nil, nil, 0, 1}, "X", "X", 1, []any{1, 1, 1}, 1, 1, nil, 0, 0, nil, 0}, id, ts, signature}
	data, err := json.Marshal(inner)
	if err != nil {
		return nil, err
	}
	outer, err := json.Marshal([]any{[]any{[]any{"Fbv4je", string(data)}}})
	if err != nil {
		return nil, err
	}
	return url.Values{"f.req": {string(outer)}}, nil
}

func ParseRPCResponse(body []byte) (string, error) {
	// Google sends an XSSI prefix, optional length lines, and RPC frames.
	// The standard frame is ["wrb.fr", "Fbv4je", "[null,\"https://...\"]", ...].
	// Some versions omit the "wrb.fr" prefix.
	data := string(body)
	start := strings.IndexByte(data, '[')
	if start < 0 {
		return "", errors.New("Google News returned no RPC frames (possibly consent or blocked response)")
	}
	decoder := json.NewDecoder(strings.NewReader(data[start:]))
	var frames []json.RawMessage
	if err := decoder.Decode(&frames); err != nil {
		return "", fmt.Errorf("Google News returned malformed RPC frames: %w", err)
	}
	for _, raw := range frames {
		var frame []json.RawMessage
		if json.Unmarshal(raw, &frame) != nil || len(frame) < 3 {
			continue
		}
		var first, second string
		_ = json.Unmarshal(frame[0], &first)
		_ = json.Unmarshal(frame[1], &second)
		payloadIndex := -1
		if first == "wrb.fr" && second == "Fbv4je" {
			payloadIndex = 2
		}
		if first == "Fbv4je" {
			payloadIndex = 2
		}
		if payloadIndex < 0 {
			continue
		}
		var payload string
		if json.Unmarshal(frame[payloadIndex], &payload) != nil {
			continue
		}
		var decoded []json.RawMessage
		if json.Unmarshal([]byte(payload), &decoded) != nil || len(decoded) < 2 {
			continue
		}
		var candidate string
		if json.Unmarshal(decoded[1], &candidate) != nil {
			continue
		}
		u, err := url.Parse(candidate)
		if err == nil && (u.Scheme == "https" || u.Scheme == "http") && u.Hostname() != "" && !strings.EqualFold(u.Hostname(), "news.google.com") && u.User == nil {
			return candidate, nil
		}
	}
	return "", errors.New("Google News RPC returned no publisher URL (Fbv4je response missing or changed)")
}

