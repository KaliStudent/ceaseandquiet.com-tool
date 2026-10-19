package addressapi

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func query(h *Handler, method, origin, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, "/api/v1/address-suggestions", strings.NewReader(body))
	r.Header.Set("Origin", origin)
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func TestNormalizeSuggestionsAndLimitProviderInput(t *testing.T) {
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if len(q) != 5 || q.Get("text") != "100 Exam" || q.Get("apiKey") != "synthetic-key" || q.Get("limit") != "5" || q.Get("filter") != "countrycode:us" {
			t.Error("unexpected provider input")
		}
		_, _ = io.WriteString(w, `{"results":[{"formatted":"100 Example Street, Example City, TS 00000","housenumber":"100","street":"Example Street","city":"Example City","state_code":"TS","postcode":"00000","country_code":"us","lat":1,"lon":2,"private":"omitted"},{"formatted":"Foreign address","street":"Foreign street","country_code":"ca"}]}`)
	}))
	defer provider.Close()
	h := New("synthetic-key")
	h.endpoint = provider.URL
	w := query(h, "POST", "https://abcw.xyz:8443", `{"text":"100 Exam"}`)
	if w.Code != 200 {
		t.Fatalf("status %d", w.Code)
	}
	var result struct {
		Items []suggestion `json:"items"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Items) != 1 || result.Items[0].Address != "100 Example Street" || result.Items[0].State != "TS" || result.Items[0].PostalCode != "00000" {
		t.Fatal("address normalization failed")
	}
	for _, prohibited := range []string{"synthetic-key", `"lat"`, `"lon"`, "private"} {
		if strings.Contains(w.Body.String(), prohibited) {
			t.Fatal("unnecessary provider data exposed")
		}
	}
	if w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("address response may be cached")
	}
}

func TestRejectInputsAndDisabledProvider(t *testing.T) {
	for _, tc := range []struct {
		method, origin, body string
		status               int
	}{
		{"POST", "https://abcw.xyz:8443", `{"text":"100 Exam"}`, 503},
		{"POST", "https://unrelated.example", `{"text":"100 Exam"}`, 403},
		{"POST", "", `{"text":"100 Exam"}`, 403},
		{"GET", "https://abcw.xyz:8443", "", 405},
		{"POST", "https://abcw.xyz:8443", `{"text":"100 Exam","email":"not-accepted"}`, 400},
		{"POST", "https://abcw.xyz:8443", `{"text":"abc"}`, 400},
		{"POST", "https://abcw.xyz:8443", `{"text":"line\nbreak"}`, 400},
		{"POST", "https://abcw.xyz:8443", `{"text":"100 Exam"} {}`, 400},
		{"POST", "https://abcw.xyz:8443", `{"text":"` + strings.Repeat("x", 1100) + `"}`, 400},
	} {
		w := query(New(""), tc.method, tc.origin, tc.body)
		if w.Code != tc.status {
			t.Errorf("status %d want %d", w.Code, tc.status)
		}
	}
	r := httptest.NewRequest("POST", "/api/v1/address-suggestions?text=not-accepted", strings.NewReader(`{"text":"100 Exam"}`))
	r.Header.Set("Origin", "https://abcw.xyz:8443")
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	New("").ServeHTTP(w, r)
	if w.Code != 400 {
		t.Fatal("accepted query string")
	}
}

func TestProviderFailuresAndRateLimit(t *testing.T) {
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Error(w, "synthetic-key and typed address", 500) }))
	defer provider.Close()
	h := New("synthetic-key")
	h.endpoint = provider.URL
	w := query(h, "POST", "https://abcw.xyz:8443", `{"text":"100 Exam"}`)
	if w.Code != 502 || strings.Contains(w.Body.String(), "synthetic-key") || strings.Contains(w.Body.String(), "typed address") {
		t.Fatal("provider failure leaked input")
	}
	h.window = time.Now()
	h.count = 60
	w = query(h, "POST", "https://abcw.xyz:8443", `{"text":"100 Exam"}`)
	if w.Code != 429 || w.Header().Get("Retry-After") != "60" {
		t.Fatal("missing provider rate limit")
	}
}
