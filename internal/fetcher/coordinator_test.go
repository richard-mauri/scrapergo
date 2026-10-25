package fetcher

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestHTMLCoordinatorSharesConcurrentFetch(t *testing.T) {
	c := newHTMLCoordinator()
	var calls atomic.Int32
	started := make(chan struct{})
	release := make(chan struct{})
	fetch := func(context.Context, string) (string, string, error) {
		if calls.Add(1) == 1 { close(started) }
		<-release
		return "<html>ok</html>", "https://example.com/story", nil
	}
	const readers = 12
	var wg sync.WaitGroup
	errs := make(chan error, readers)
	wg.Add(1)
	go func() {
		defer wg.Done()
		html, _, err := c.get(context.Background(), "https://example.com/story", fetch)
		if err != nil || html != "<html>ok</html>" { errs <- errors.New("first read failed") }
	}()
	<-started
	// Wait until the leader is held in the publisher fetch, then ensure the
	// other callers observe either its in-flight operation or its cached result.
	for i := 1; i < readers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			html, _, err := c.get(context.Background(), "https://example.com/story", fetch)
			if err != nil || html != "<html>ok</html>" { errs <- errors.New("shared read failed") }
		}()
	}
	close(release)
	wg.Wait()
	close(errs)
	for err := range errs { t.Fatal(err) }
	if got := calls.Load(); got != 1 { t.Fatalf("publisher fetches: got %d, want 1", got) }
}

func TestHTMLCoordinatorDoesNotCacheFailures(t *testing.T) {
	c := newHTMLCoordinator()
	var calls int
	fetch := func(context.Context, string) (string, string, error) {
		calls++
		if calls == 1 { return "", "", errors.New("HTTP 403") }
		return "<html>recovered</html>", "https://example.com/story", nil
	}
	if _, _, err := c.get(context.Background(), "story", fetch); err == nil { t.Fatal("expected denial") }
	html, _, err := c.get(context.Background(), "story", fetch)
	if err != nil || html != "<html>recovered</html>" || calls != 2 { t.Fatalf("recovery failed: %q, %v, calls=%d", html, err, calls) }
	_, _, err = c.get(context.Background(), "story", fetch)
	if err != nil || calls != 2 { t.Fatalf("successful HTML not reused: %v, calls=%d", err, calls) }
}

func TestHTMLCoordinatorWaitingCancellation(t *testing.T) {
	c := newHTMLCoordinator()
	started := make(chan struct{})
	release := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _, _ = c.get(context.Background(), "story", func(context.Context, string) (string, string, error) {
			close(started)
			<-release
			return "<html>ok</html>", "story", nil
		})
	}()
	<-started
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	begin := time.Now()
	_, _, err := c.get(ctx, "story", func(context.Context, string) (string, string, error) { t.Error("waiter fetched independently"); return "", "", nil })
	if !errors.Is(err, context.Canceled) || time.Since(begin) > time.Second { t.Fatalf("canceled waiter: %v", err) }
	close(release)
	wg.Wait()
}

func TestHTMLCoordinatorDoesNotCacheCanceledOrOversize(t *testing.T) {
	c := newHTMLCoordinator()
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	fetch := func(context.Context, string) (string, string, error) {
		calls++
		if calls == 1 { cancel(); return "<html>not reusable</html>", "story", nil }
		return "<html>ok</html>", "story", nil
	}
	_, _, _ = c.get(ctx, "story", fetch)
	_, _, err := c.get(context.Background(), "story", fetch)
	if err != nil || calls != 2 { t.Fatalf("canceled response cached; calls=%d err=%v", calls, err) }

	c2 := newHTMLCoordinator()
	sizeCalls := 0
	big := make([]byte, MaxHTML+1)
	oversize := func(context.Context, string) (string, string, error) { sizeCalls++; return string(big), "story", nil }
	_, _, _ = c2.get(context.Background(), "story", oversize)
	_, _, _ = c2.get(context.Background(), "story", oversize)
	if sizeCalls != 2 { t.Fatalf("oversize cached; calls=%d", sizeCalls) }
}
