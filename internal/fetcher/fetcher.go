// Package fetcher isolates ScraperGo's validated HTML retrieval and network security.
package fetcher

import (
 "context"
 "encoding/json"
 "html"
 "path"
 "errors"
 "fmt"
 "io"
 "log"
 "net"
 "net/http"
 "net/http/cookiejar"
 "net/url"
 "strings"
 "time"
)

const MaxHTML = 5 << 20

func SafeHost(ctx context.Context, host string) error {
	if host == "" {
		return errors.New("missing hostname")
	}
	if strings.EqualFold(host, "localhost") || strings.HasSuffix(strings.ToLower(host), ".localhost") {
		return errors.New("local addresses are not allowed")
	}
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return fmt.Errorf("DNS lookup failed: %w", err)
	}
	for _, ip := range ips {
		v := ip.IP
		if !v.IsGlobalUnicast() || v.IsPrivate() || v.IsLoopback() || v.IsLinkLocalUnicast() || v.IsLinkLocalMulticast() {
			return errors.New("private or reserved network addresses are not allowed")
		}
	}
	return nil
}
func ValidatedURL(ctx context.Context, s string) (*url.URL, error) {
	u, err := url.Parse(strings.TrimSpace(s))
	if err != nil {
		return nil, err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, errors.New("only HTTP and HTTPS URLs are supported")
	}
	if u.User != nil {
		return nil, errors.New("URL credentials are not supported")
	}
	if err := SafeHost(ctx, u.Hostname()); err != nil {
		return nil, err
	}
	return u, nil
}

// setBrowserHeaders uses ordinary Chrome-style navigation headers. These can
// improve compatibility, but cannot reproduce Safari's TLS fingerprint or cookies.
func SetBrowserHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.8")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
}

// setReadeckLikeHeaders provides a second, internally consistent Chromium
// request profile. This is an independently implemented approximation, not
// Readeck code, and does not reproduce its TLS fingerprint.
func SetReadeckLikeHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/jpeg,image/png,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Language", "en-US,en;q=0.8")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-CH-UA", `"Google Chrome";v="144", "Chromium";v="144"`)
	req.Header.Set("Sec-CH-UA-Mobile", "?0")
	req.Header.Set("Sec-CH-UA-Platform", `"Windows"`)
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
}

// alternateHeaderRetry applies only to a publisher's explicit refusal.
// A second 403 is returned normally rather than classified as a paywall.
func AlternateHeaderRetry(status int) bool { return status == http.StatusForbidden }

func downloadUncached(ctx context.Context, address string) (string, string, error) {
	u, err := ValidatedURL(ctx, address)
	if err != nil {
		return "", "", err
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}
		if err := SafeHost(ctx, host); err != nil {
			return nil, err
		}
		return (&net.Dialer{Timeout: 8 * time.Second}).DialContext(ctx, network, net.JoinHostPort(host, port))
	}
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Timeout: 18 * time.Second, Transport: transport, Jar: jar, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return errors.New("too many redirects")
		}
		_, err := ValidatedURL(req.Context(), req.URL.String())
		return err
	}}
	// The initial attempt preserves v0.2.14 behavior. Only a 403 triggers
	// one alternative request, with a fresh cookie jar and no publisher tab.
	started := time.Now()
	var res *http.Response
	for attempt := 0; attempt < 2; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			return "", "", err
		}
		profile := "existing"
		if attempt == 0 {
			SetBrowserHeaders(req)
		} else {
			profile = "readeck-like"
			SetReadeckLikeHeaders(req)
		}
		res, err = client.Do(req)
		if err != nil {
			log.Printf("fetch host=%s profile=%s error=%v elapsed=%s", u.Hostname(), profile, err, time.Since(started))
			return "", "", err
		}
		log.Printf("fetch host=%s profile=%s status=%d redirect_host=%s content_type=%q elapsed=%s",
			u.Hostname(), profile, res.StatusCode, res.Request.URL.Hostname(), res.Header.Get("Content-Type"), time.Since(started))
		if attempt == 0 && AlternateHeaderRetry(res.StatusCode) {
			_ = res.Body.Close()
			// A retry should not carry over cookies from the rejected response.
			freshJar, _ := cookiejar.New(nil)
			client.Jar = freshJar
			continue
		}
		break
	}
	defer res.Body.Close()
	challenge := false
	if res.StatusCode == http.StatusAccepted {
		// Capture bounded response characteristics without logging page text.
		sample, sampleErr := io.ReadAll(io.LimitReader(res.Body, 4096))
		if sampleErr != nil {
			log.Printf("fetch accepted-response host=%s sample_error=%v", res.Request.URL.Hostname(), sampleErr)
		} else {
			lower := strings.ToLower(string(sample))
			challenge = strings.Contains(lower, "challenge") || strings.Contains(lower, "cf-browser-verification") || strings.Contains(lower, "captcha")
			log.Printf("fetch accepted-response host=%s sample_bytes=%d html_hint=%t article_hint=%t challenge_hint=%t retry_after_present=%t",
				res.Request.URL.Hostname(), len(sample),
				strings.Contains(lower, "<html") || strings.Contains(lower, "<!doctype html"),
				strings.Contains(lower, "<article"),
				challenge,
				res.Header.Get("Retry-After") != "")
		}
	}
	if res.StatusCode == http.StatusAccepted && challenge {
		// A challenge is not article HTML. Try only the publisher's public
		// WordPress post API, without copying browser verification tokens.
		if fallbackHTML, ok := wordpressFallback(ctx, client, u); ok {
			log.Printf("fetch wordpress-fallback host=%s success=true", u.Hostname())
			return fallbackHTML, u.String(), nil
		}
	}
	if res.StatusCode != http.StatusOK {
		return "", "", &PublisherHTTPError{Host: res.Request.URL.Hostname(), StatusCode: res.StatusCode}
	}
	if c := res.Header.Get("Content-Type"); c != "" && !strings.Contains(strings.ToLower(c), "html") {
		return "", "", errors.New("URL did not return an HTML page")
	}
	data, err := io.ReadAll(io.LimitReader(res.Body, MaxHTML+1))
	if err != nil {
		return "", "", err
	}
	if len(data) > MaxHTML {
		return "", "", errors.New("page exceeds 5 MB limit")
	}
	return string(data), res.Request.URL.String(), nil
}

// PublisherHTTPError preserves the HTTP status without assuming why access was denied.
type PublisherHTTPError struct {
	Host       string
	StatusCode int
}

// PublisherStatusCode exposes the upstream response without coupling API adapters to fetcher.\nfunc (e *PublisherHTTPError) PublisherStatusCode() int { return e.StatusCode }\n\nfunc (e *PublisherHTTPError) Error() string {
	return fmt.Sprintf("article request to %s returned HTTP %d (%s)", e.Host, e.StatusCode, http.StatusText(e.StatusCode))
}

// A 401 or 403 is an access denial, not proof of a subscription requirement.
func PublisherAccessMessage(err error) (string, bool) {
	var denied *PublisherHTTPError
	if errors.As(err, &denied) && (denied.StatusCode == http.StatusUnauthorized || denied.StatusCode == http.StatusForbidden) {
		return "Publisher access restricted. The publisher declined ScraperGo's request. Open the original website to view any available preview, sign-in, or subscription options.", true
	}
	return "", false
}


type wpPublicPost struct {
 Slug string `json:"slug"`
 Link string `json:"link"`
 Status string `json:"status"`
 Title struct { Rendered string `json:"rendered"` } `json:"title"`
 Content struct { Rendered string `json:"rendered"`; Protected bool `json:"protected"` } `json:"content"`
}

func canonicalWPURL(u *url.URL) string {
 copy := *u
 copy.RawQuery = ""
 copy.Fragment = ""
 copy.Path = strings.TrimSuffix(copy.EscapedPath(), "/")
 copy.RawPath = ""
 return copy.String()
}

// decodePublicWPPost rejects mismatched slugs/canonical URLs and protected content.
func decodePublicWPPost(data []byte, original *url.URL, slug string) (string, bool) {
 var posts []wpPublicPost
 if err := json.Unmarshal(data, &posts); err != nil || len(posts) != 1 { return "", false }
 p := posts[0]
 link, err := url.Parse(p.Link)
 if err != nil || link.Scheme != original.Scheme || !strings.EqualFold(link.Host, original.Host) ||
 canonicalWPURL(link) != canonicalWPURL(original) ||
 p.Slug != slug || p.Status != "publish" || p.Content.Protected || strings.TrimSpace(p.Content.Rendered) == "" { return "", false }
 // Existing HTML extraction remains the single source of article rendering.
 return "<!doctype html><html><head><meta charset=\"utf-8\"><title>" +
 html.EscapeString(p.Title.Rendered) + "</title><link rel=\"canonical\" href=\"" +
 html.EscapeString(original.String()) + "\"></head><body><article><h1>" +
 html.EscapeString(p.Title.Rendered) + "</h1>" + p.Content.Rendered +
 "</article></body></html>", true
}

// wordpressFallback performs one same-origin, bounded lookup after a challenge.
// Its network client retains the existing SSRF, redirect and timeout safeguards.
func wordpressFallback(ctx context.Context, client *http.Client, original *url.URL) (string, bool) {
 slug := path.Base(strings.TrimSuffix(original.Path, "/"))
 if slug == "" || slug == "." || slug == "/" || strings.Contains(slug, ".") { return "", false }
 api := *original
 api.Path = "/wp-json/wp/v2/posts"
 api.RawPath = ""
 q := url.Values{}
 q.Set("slug",slug)
 q.Set("_fields","id,slug,link,status,title,content")
 api.RawQuery = q.Encode()
 api.Fragment = ""
 req, err := http.NewRequestWithContext(ctx, http.MethodGet, api.String(), nil)
 if err != nil { return "", false }
 req.Header.Set("Accept","application/json")
 response, err := client.Do(req)
 if err != nil { return "", false }
 defer response.Body.Close()
 if response.StatusCode != http.StatusOK || response.Request.URL.Scheme != original.Scheme ||
 !strings.EqualFold(response.Request.URL.Host, original.Host) ||
 !strings.Contains(strings.ToLower(response.Header.Get("Content-Type")), "application/json") { return "", false }
 payload, err := io.ReadAll(io.LimitReader(response.Body, MaxHTML+1))
 if err != nil || len(payload)>MaxHTML { return "", false }
 return decodePublicWPPost(payload, original, slug)
}
