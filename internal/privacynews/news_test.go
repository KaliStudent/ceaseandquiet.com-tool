package privacynews

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestPrivacyFeedIsFilteredCachedAndServedAsStaleOnFailure(t *testing.T) {
	requests := 0
	broken := false
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests++
		if broken {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = fmt.Fprint(w, `<rss><channel><item><title>Privacy story</title><link>https://www.eff.org/deeplinks/privacy-example</link><pubDate>Mon, 19 Oct 2026 00:00:00 +0000</pubDate><description>Data broker privacy</description></item><item><title>Unrelated</title><link>https://www.eff.org/deeplinks/unrelated</link><pubDate>Mon, 19 Oct 2026 00:00:00 +0000</pubDate><description>Other topic</description></item><item><title>Unsafe privacy link</title><link>javascript:alert(1)</link><pubDate>Mon, 19 Oct 2026 00:00:00 +0000</pubDate></item><item><title>Privacy destination outside source</title><link>https://unrelated.example/</link><pubDate>Mon, 19 Oct 2026 00:00:00 +0000</pubDate></item></channel></rss>`)
	}))
	defer provider.Close()
	h := New()
	h.endpoint = provider.URL
	now := time.Date(2026, 10, 19, 1, 0, 0, 0, time.UTC)
	h.now = func() time.Time { return now }
	get := func() *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/privacy-news", nil))
		return w
	}
	w := get()
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "Privacy story") || strings.Contains(w.Body.String(), "Unrelated") || strings.Contains(w.Body.String(), "javascript") {
		t.Fatal("news content boundary failed")
	}
	get()
	if requests != 1 {
		t.Fatal("feed cache missed")
	}
	broken = true
	now = now.Add(16 * time.Minute)
	w = get()
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"stale":true`) {
		t.Fatal("cached headlines not identified as stale")
	}
	get()
	if requests != 2 {
		t.Fatal("failed feed retried without backoff")
	}
}

func TestUnavailableFeedDoesNotInventHeadlines(t *testing.T) {
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusServiceUnavailable) }))
	defer provider.Close()
	h := New()
	h.endpoint = provider.URL
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/privacy-news", nil))
	if w.Code != http.StatusServiceUnavailable || strings.Contains(w.Body.String(), `"items"`) {
		t.Fatal("unavailable news reported as real headlines")
	}
}
