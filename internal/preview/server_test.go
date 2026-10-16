package preview

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPreviewBoundaries(t *testing.T) {
	handler, err := New("127.0.0.1:8787")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		method, path, host string
		status             int
	}{
		{"GET", "/", "127.0.0.1:8787", 200},
		{"HEAD", "/styles.css", "127.0.0.1:8787", 200},
		{"GET", "/site.js", "127.0.0.1:8787", 200},
		{"GET", "/", "attacker.example:8787", 421},
		{"POST", "/", "127.0.0.1:8787", 405},
		{"POST", "/api/v1/brokers", "127.0.0.1:8787", 405},
		{"DELETE", "/api/v1/brokers/spokeo", "127.0.0.1:8787", 405},
		{"GET", "/extension/popup.js", "127.0.0.1:8787", 404},
		{"GET", "/embed.go", "127.0.0.1:8787", 404},
		{"GET", "/?email=synthetic", "127.0.0.1:8787", 400},
	}
	for _, tt := range tests {
		t.Run(tt.method+tt.path+tt.host, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.path, nil)
			request.Host = tt.host
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != tt.status {
				t.Fatalf("status=%d want=%d", response.Code, tt.status)
			}
			if !strings.Contains(response.Header().Get("Content-Security-Policy"), "form-action 'none'") {
				t.Fatal("missing browser form submission protection")
			}
			if response.Header().Get("Cache-Control") != "no-store" {
				t.Fatal("preview may be cached")
			}
			if response.Header().Get("Access-Control-Allow-Origin") != "" {
				t.Fatal("unexpected cross-origin access")
			}
			if tt.method == "HEAD" && response.Body.Len() != 0 {
				t.Fatal("HEAD returned a body")
			}
		})
	}
}

func TestRejectPublicListener(t *testing.T) {
	for _, addr := range []string{":8787", "0.0.0.0:8787", "192.0.2.1:8787", "example.test:8787"} {
		if _, err := New(addr); err == nil {
			t.Fatalf("accepted public listener %q", addr)
		}
	}
}
