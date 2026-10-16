package staging

import (
	"crypto/sha256"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTLSStagingAccess(t *testing.T) {
	const host, user, password = "abcw.xyz:8443", "reviewer", "synthetic-test-password-only"
	handler, err := New(host, user, sha256.Sum256([]byte(password)))
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewTLSServer(handler)
	defer server.Close()
	for _, tc := range []struct {
		method, path, host, user, password string
		status                             int
	}{
		{"GET", "/", host, "", "", 401},
		{"GET", "/site.js", host, user, "wrong", 401},
		{"GET", "/api/v1/brokers", host, "other", password, 401},
		{"GET", "/", "other.example:8443", user, password, 421},
		{"GET", "/", host, user, password, 200},
		{"HEAD", "/styles.css", host, user, password, 200},
		{"GET", "/api/v1/brokers?q=spokeo", host, user, password, 200},
		{"POST", "/api/v1/brokers", host, user, password, 405},
		{"GET", "/config.yaml", host, user, password, 404},
		{"GET", "/?email=synthetic", host, user, password, 400},
	} {
		t.Run(tc.method+tc.path+tc.host+tc.user+tc.password, func(t *testing.T) {
			r, err := http.NewRequest(tc.method, server.URL+tc.path, nil)
			if err != nil {
				t.Fatal(err)
			}
			r.Host = tc.host
			if tc.user != "" {
				r.SetBasicAuth(tc.user, tc.password)
			}
			response, err := server.Client().Do(r)
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			body, err := io.ReadAll(response.Body)
			if err != nil {
				t.Fatal(err)
			}
			if response.StatusCode != tc.status {
				t.Fatalf("status %d want %d", response.StatusCode, tc.status)
			}
			if response.Header.Get("Cache-Control") != "no-store" || !strings.Contains(response.Header.Get("Content-Security-Policy"), "form-action 'none'") {
				t.Fatal("missing privacy headers")
			}
			if strings.Contains(string(body), password) {
				t.Fatal("credential leaked")
			}
			if tc.status == 401 && response.Header.Get("WWW-Authenticate") == "" {
				t.Fatal("missing authentication challenge")
			}
			if tc.method == "HEAD" && len(body) != 0 {
				t.Fatal("HEAD body")
			}
		})
	}
	r := httptest.NewRequest("GET", "http://"+host+"/", nil)
	r.SetBasicAuth(user, password)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != 421 {
		t.Fatal("accepted plaintext request")
	}
}

func TestRequiredConfiguration(t *testing.T) {
	for _, host := range []string{"", "abcw.xyz", "abcw.xyz:0", ":8443", "abcw.xyz:65536", "user@abcw.xyz:8443"} {
		if _, err := New(host, "reviewer", sha256.Sum256([]byte("synthetic"))); err == nil {
			t.Fatalf("accepted %q", host)
		}
	}
	if _, err := New("abcw.xyz:8443", "", sha256.Sum256([]byte("synthetic"))); err == nil {
		t.Fatal("accepted missing user")
	}
	if _, err := New("abcw.xyz:8443", "reviewer", [sha256.Size]byte{}); err == nil {
		t.Fatal("accepted missing digest")
	}
}
