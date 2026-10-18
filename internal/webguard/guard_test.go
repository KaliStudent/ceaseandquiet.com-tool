package webguard

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRateLimitCannotBeBypassedWithForwardedHeaders(t *testing.T) {
	g := New()
	now := time.Now()
	g.now = func() time.Time { return now }
	h := g.Wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	for i := range 181 {
		r := httptest.NewRequest("GET", "/", nil)
		r.Header.Set("X-Forwarded-For", fmt.Sprintf("192.0.2.%d", i))
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		want := http.StatusNoContent
		if i == 180 {
			want = http.StatusTooManyRequests
		}
		if w.Code != want {
			t.Fatalf("request %d status %d want %d", i, w.Code, want)
		}
	}
	now = now.Add(time.Minute)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
	if w.Code != http.StatusNoContent {
		t.Fatal("rate window did not reset")
	}
}

func TestBodyAndConcurrencyBounds(t *testing.T) {
	g := New()
	h := g.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := io.ReadAll(r.Body); err != nil {
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	for _, chunked := range []bool{false, true} {
		r := httptest.NewRequest("POST", "/", strings.NewReader(strings.Repeat("x", 16385)))
		if chunked {
			r.ContentLength = -1
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != http.StatusRequestEntityTooLarge {
			t.Fatal("oversized body accepted")
		}
	}
	for range cap(g.slots) {
		g.slots <- struct{}{}
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
	if w.Code != http.StatusServiceUnavailable {
		t.Fatal("concurrency cap bypassed")
	}
}

func TestGlobalBudgetAndPeerMemoryBound(t *testing.T) {
	g := New()
	for i := range 600 {
		if !g.allow(fmt.Sprintf("192.0.%d.%d:1234", i/250, i%250)) {
			t.Fatal("early global rejection")
		}
	}
	if g.allow("198.51.100.1:1234") {
		t.Fatal("global request budget bypassed")
	}
	g.count = 0
	for i := range 1024 {
		g.peers[shaKey(i)] = 1
	}
	if g.allow("203.0.113.1:1234") {
		t.Fatal("peer map memory cap bypassed")
	}
}

func shaKey(i int) [32]byte { var key [32]byte; key[0] = byte(i); key[1] = byte(i >> 8); return key }
