// Package webguard bounds work and request sizes before website handlers run.
package webguard

import (
	"crypto/rand"
	"crypto/sha256"
	"net"
	"net/http"
	"sync"
	"time"
)

type Guard struct {
	mu     sync.Mutex
	window time.Time
	count  int
	peers  map[[32]byte]int
	salt   [32]byte
	now    func() time.Time
	slots  chan struct{}
}

func New() *Guard {
	g := &Guard{peers: make(map[[32]byte]int), now: time.Now, slots: make(chan struct{}, 32)}
	_, _ = rand.Read(g.salt[:])
	return g
}

func (g *Guard) allow(remote string) bool {
	// Only the TCP peer is trusted. Forwarded headers cannot select a new bucket.
	host, _, err := net.SplitHostPort(remote)
	if err != nil {
		host = remote
	}
	if ip := net.ParseIP(host); ip != nil {
		host = ip.String()
	}
	digest := sha256.Sum256(append(g.salt[:], []byte(host)...))
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.now().Sub(g.window) >= time.Minute {
		g.window = g.now()
		g.count = 0
		clear(g.peers)
	}
	if g.count >= 600 || g.peers[digest] >= 180 {
		return false
	}
	if _, known := g.peers[digest]; !known && len(g.peers) >= 1024 {
		return false
	}
	g.count++
	g.peers[digest]++
	return true
}

func (g *Guard) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		if !g.allow(r.RemoteAddr) {
			w.Header().Set("Retry-After", "60")
			http.Error(w, "Too many requests. Please wait one minute.", http.StatusTooManyRequests)
			return
		}
		select {
		case g.slots <- struct{}{}:
			defer func() { <-g.slots }()
		default:
			w.Header().Set("Retry-After", "5")
			http.Error(w, "Service busy. Please try again shortly.", http.StatusServiceUnavailable)
			return
		}
		if r.ContentLength > 16384 {
			http.Error(w, "Request too large.", http.StatusRequestEntityTooLarge)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 16384)
		next.ServeHTTP(w, r)
	})
}
