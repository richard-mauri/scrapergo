package googlenews

import (
 "clean-article-reader/internal/fetcher"
 "context"
 "fmt"
 "html"
 "io"
 "net/http"
 "strings"
 "time"
)

func ResolveURL(ctx context.Context, address string) (string, error) {
	id, isNews, err := ArticleID(address)
	if err != nil {
		return "", err
	}
	if !isNews {
		return address, nil
	}
	if legacy := LegacyPublisherURL(id); legacy != "" {
		if _, err := fetcher.ValidatedURL(ctx, legacy); err != nil {
			return "", err
		}
		return legacy, nil
	}
	// Fetch the Google News article's current signature and timestamp.
	// The endpoint is fixed: user-supplied data appears only in a validated path.
	endpoint := "https://news.google.com/rss/articles/" + id + "?hl=en-US&gl=US&ceid=US:en"
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	client := &http.Client{Timeout: 18 * time.Second, Transport: transport, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) > 0 {
			return http.ErrUseLastResponse
		}
		return nil
	}}
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 Safari/605.1.15")
	req.Header.Set("Cookie", "CONSENT=PENDING+987")
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Google News lookup failed: %w", err)
	}
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	resp.Body.Close()
	if readErr != nil {
		return "", readErr
	}
	signature := NewsSignatureRE.FindSubmatch(body)
	timestamp := NewsTimestampRE.FindSubmatch(body)
	if len(signature) < 2 || len(timestamp) < 2 {
		return "", fmt.Errorf("Google News signature lookup failed (HTTP %d; missing article signature or timestamp). The page may require consent or Google may have changed its format", resp.StatusCode)
	}
	values, err := RPCPayload(id, string(timestamp[1]), html.UnescapeString(string(signature[1])))
	if err != nil {
		return "", err
	}
	rpcReq, err := http.NewRequestWithContext(ctx, "POST", "https://news.google.com/_/DotsSplashUi/data/batchexecute?rpcids=Fbv4je", strings.NewReader(values.Encode()))
	if err != nil {
		return "", err
	}
	rpcReq.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
	rpcReq.Header.Set("Referer", "https://news.google.com/")
	rpcReq.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 Safari/605.1.15")
	result, err := client.Do(rpcReq)
	if err != nil {
		return "", fmt.Errorf("Google News resolution request failed: %w", err)
	}
	rpcData, readErr := io.ReadAll(io.LimitReader(result.Body, 2<<20))
	result.Body.Close()
	if readErr != nil {
		return "", readErr
	}
	if result.StatusCode != 200 {
		return "", fmt.Errorf("Google News Fbv4je RPC returned HTTP %d", result.StatusCode)
	}
	publisher, err := ParseRPCResponse(rpcData)
	if err != nil {
		return "", err
	}
	if _, err = fetcher.ValidatedURL(ctx, publisher); err != nil {
		return "", fmt.Errorf("invalid publisher destination: %w", err)
	}
	return publisher, nil
}
