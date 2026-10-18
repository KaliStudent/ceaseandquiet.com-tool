package formsecurity

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func request(s *Service, path, body string, cookie *http.Cookie) *httptest.ResponseRecorder {
	r := httptest.NewRequest(http.MethodPost, "/api/v1/form-security/"+path, strings.NewReader(body))
	r.TLS = &tls.ConnectionState{}
	r.Header.Set("Origin", "https://abcw.xyz:8443")
	r.Header.Set("Content-Type", "application/json")
	if cookie != nil {
		r.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	s.ServeHTTP(w, r)
	return w
}
func createSession(t *testing.T, s *Service) *http.Cookie {
	t.Helper()
	w := request(s, "session", `{"token":"synthetic-human-token"}`, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("session status %d", w.Code)
	}
	cookies := w.Result().Cookies()
	if len(cookies) != 1 || !cookies[0].Secure || !cookies[0].HttpOnly || cookies[0].SameSite != http.SameSiteStrictMode {
		t.Fatal("unsafe form cookie")
	}
	return cookies[0]
}
func testConfig() Config {
	return Config{SiteKey: "synthetic-public-site-key", CheckHuman: func(_ context.Context, token string) error {
		if token != "synthetic-human-token" {
			return errors.New("invalid")
		}
		return nil
	}}
}

func TestEmailOwnershipMustBeConfirmedAndIsBoundToAddress(t *testing.T) {
	cfg := testConfig()
	var code string
	cfg.SendCode = func(_ context.Context, address, value string) error {
		if address != "taylor@example.test" {
			t.Fatal("unexpected address")
		}
		code = value
		return nil
	}
	s := New(cfg)
	cookie := createSession(t, s)
	if w := request(s, "email/start", `{"email":"taylor@example.test"}`, nil); w.Code != http.StatusForbidden {
		t.Fatal("email sent without human session")
	}
	if w := request(s, "email/start", `{"email":"taylor@example.test"}`, cookie); w.Code != http.StatusAccepted {
		t.Fatal("code not sent")
	}
	if len(code) != 6 {
		t.Fatal("invalid code size")
	}
	r := httptest.NewRequest("POST", "/api/v1/requests", nil)
	r.AddCookie(cookie)
	if s.RequireVerifiedEmail(r, "taylor@example.test") {
		t.Fatal("unconfirmed email trusted")
	}
	body, _ := json.Marshal(map[string]string{"email": "taylor@example.test", "code": code})
	if w := request(s, "email/confirm", string(body), cookie); w.Code != http.StatusOK {
		t.Fatal("confirmation failed")
	}
	if !s.RequireVerifiedEmail(r, "taylor@example.test") || s.RequireVerifiedEmail(r, "different@example.test") || s.RequireVerifiedEmail(r, "Taylor@example.test") {
		t.Fatal("email ownership binding failed")
	}
	if w := request(s, "email/confirm", string(body), cookie); w.Code != http.StatusUnauthorized {
		t.Fatal("code replay accepted")
	}
	if w := request(s, "email/status", `{"email":"different@example.test"}`, cookie); strings.Contains(w.Body.String(), `true`) {
		t.Fatal("changed email remained verified")
	}
}

func TestExpiryAttemptCapAndSendCooldown(t *testing.T) {
	cfg := testConfig()
	var code string
	count := 0
	cfg.SendCode = func(_ context.Context, _, value string) error { code = value; count++; return nil }
	s := New(cfg)
	now := time.Now()
	s.now = func() time.Time { return now }
	cookie := createSession(t, s)
	request(s, "email/start", `{"email":"taylor@example.test"}`, cookie)
	if w := request(s, "email/start", `{"email":"TAYLOR@example.test"}`, cookie); w.Code != http.StatusTooManyRequests || count != 1 {
		t.Fatal("cooldown bypassed by case change")
	}
	wrong := "000000"
	if code == wrong {
		wrong = "111111"
	}
	for range 5 {
		body, _ := json.Marshal(map[string]string{"email": "taylor@example.test", "code": wrong})
		request(s, "email/confirm", string(body), cookie)
	}
	body, _ := json.Marshal(map[string]string{"email": "taylor@example.test", "code": code})
	if w := request(s, "email/confirm", string(body), cookie); w.Code != http.StatusUnauthorized {
		t.Fatal("attempt limit bypassed")
	}
	now = now.Add(time.Minute)
	request(s, "email/start", `{"email":"taylor@example.test"}`, cookie)
	now = now.Add(11 * time.Minute)
	body, _ = json.Marshal(map[string]string{"email": "taylor@example.test", "code": code})
	if w := request(s, "email/confirm", string(body), cookie); w.Code != http.StatusUnauthorized {
		t.Fatal("expired code accepted")
	}
	now = now.Add(20 * time.Minute)
	if s.HasHuman(func() *http.Request { r := httptest.NewRequest("POST", "/", nil); r.AddCookie(cookie); return r }()) {
		t.Fatal("expired human session accepted")
	}
}

func TestInputRejectionAndUnavailableMailFailClosed(t *testing.T) {
	s := New(testConfig())
	cookie := createSession(t, s)
	for _, body := range []string{`{"email":"Taylor <taylor@example.test>"}`, `{"email":"taylor@example.test","website":"spam"}`, `{"email":"taylor@example.test","profile":"not-accepted"}`, `{"email":"taylor@example.test"} {}`, `{"email":"` + strings.Repeat("x", 5000) + `"}`} {
		if w := request(s, "email/start", body, cookie); w.Code != http.StatusBadRequest {
			t.Fatalf("invalid input accepted: status %d", w.Code)
		}
	}
	if w := request(s, "email/start", `{"email":"taylor@example.test"}`, cookie); w.Code != http.StatusServiceUnavailable {
		t.Fatal("unconfigured sender claimed success")
	}
	if w := request(s, "session", `{"token":"invalid-token"}`, nil); w.Code != http.StatusForbidden {
		t.Fatal("invalid bot token accepted")
	}
	r := httptest.NewRequest("POST", "/api/v1/form-security/session", strings.NewReader(`{"token":"synthetic-human-token"}`))
	r.Header.Set("Origin", "https://unrelated.example")
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, r)
	if w.Code != http.StatusForbidden {
		t.Fatal("cross-origin request accepted")
	}
	disabled := New(Config{})
	if w := request(disabled, "session", `{"token":"anything"}`, nil); w.Code != http.StatusServiceUnavailable {
		t.Fatal("disabled bot check bypassed")
	}
}

func TestProtectedAddressAndMailFailure(t *testing.T) {
	cfg := testConfig()
	cfg.SendCode = func(context.Context, string, string) error { return errors.New("private provider details") }
	s := New(cfg)
	cookie := createSession(t, s)
	next := s.ProtectAddress(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	r := httptest.NewRequest("POST", "/api/v1/address-suggestions", nil)
	w := httptest.NewRecorder()
	next.ServeHTTP(w, r)
	if w.Code != http.StatusForbidden {
		t.Fatal("address provider called without human session")
	}
	r.AddCookie(cookie)
	w = httptest.NewRecorder()
	next.ServeHTTP(w, r)
	if w.Code != http.StatusNoContent {
		t.Fatal("human address lookup refused")
	}
	w = request(s, "email/start", `{"email":"taylor@example.test"}`, cookie)
	if w.Code != http.StatusServiceUnavailable || strings.Contains(w.Body.String(), "private") {
		t.Fatal("mail failure leaked details or claimed success")
	}
}
