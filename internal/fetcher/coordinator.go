package fetcher

import (
	"context"
	"sync"
	"time"
)

// Successful, unauthenticated HTML is retained briefly so classification,
// thumbnail generation and reading can share work. Failed fetches are never
// retained. No cookies or request-specific headers enter this cache.
const (
	sharedHTMLTTL   = 30 * time.Second
	sharedHTMLItems = 12
	sharedHTMLBytes = 12 << 20
)

type htmlResult struct {
	html, finalURL string
	err            error
}

type htmlEntry struct {
	html, finalURL string
	expires        time.Time
}

type inFlightHTML struct {
	done chan struct{}
	result htmlResult
}

type htmlCoordinator struct {
	mu       sync.Mutex
	entries  map[string]htmlEntry
	pending  map[string]*inFlightHTML
}

func newHTMLCoordinator() *htmlCoordinator {
	return &htmlCoordinator{entries: make(map[string]htmlEntry), pending: make(map[string]*inFlightHTML)}
}

var articleHTML = newHTMLCoordinator()

// Download applies the public-network validation on every request, including
// cache hits. Redirect and response checks remain owned by downloadUncached.
func Download(ctx context.Context, address string) (string, string, error) {
	u, err := ValidatedURL(ctx, address)
	if err != nil {
		return "", "", err
	}
	return articleHTML.get(ctx, u.String(), downloadUncached)
}

func (c *htmlCoordinator) get(ctx context.Context, key string, fetch func(context.Context, string) (string, string, error)) (string, string, error) {
	if err := ctx.Err(); err != nil {
		return "", "", err
	}
	c.mu.Lock()
	if hit, ok := c.entries[key]; ok {
		if time.Now().Before(hit.expires) {
			c.mu.Unlock()
			return hit.html, hit.finalURL, nil
		}
		delete(c.entries, key)
	}
	if pending := c.pending[key]; pending != nil {
		c.mu.Unlock()
		select {
		case <-pending.done:
			if err := ctx.Err(); err != nil {
				return "", "", err
			}
			return pending.result.html, pending.result.finalURL, pending.result.err
		case <-ctx.Done():
			return "", "", ctx.Err()
		}
	}
	pending := &inFlightHTML{done: make(chan struct{})}
	c.pending[key] = pending
	c.mu.Unlock()

	html, finalURL, err := fetch(ctx, key)
	c.mu.Lock()
	pending.result = htmlResult{html: html, finalURL: finalURL, err: err}
	delete(c.pending, key)
	// An in-memory cache is bounded by item count and byte count; zero-length
	// results, errors and canceled operations must not survive.
	if err == nil && ctx.Err() == nil && len(html) > 0 && len(html) <= MaxHTML {
		now := time.Now()
		total := 0
		for k, v := range c.entries {
			if !now.Before(v.expires) {
				delete(c.entries, k)
			} else {
				total += len(v.html)
			}
		}
		// Simple bounded FIFO-independent eviction; cache efficiency is
		// secondary to memory limits and avoiding long-term retention.
		for len(c.entries) >= sharedHTMLItems || total+len(html) > sharedHTMLBytes {
			if len(c.entries) == 0 {
				break
			}
			for k, v := range c.entries {
				total -= len(v.html)
				delete(c.entries, k)
				break
			}
		}
		c.entries[key] = htmlEntry{html: html, finalURL: finalURL, expires: now.Add(sharedHTMLTTL)}
	}
	close(pending.done)
	c.mu.Unlock()
	return html, finalURL, err
}

