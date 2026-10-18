// Package formsecurity provides bounded, temporary human/email verification.
// Addresses, codes, tokens and provider responses are never logged or persisted.
package formsecurity

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"net/http"
	"net/mail"
	"strings"
	"sync"
	"time"
)

type Config struct {
	SiteKey    string
	CheckHuman func(context.Context, string) error
	SendCode   func(context.Context, string, string) error
}
type session struct {
	expires     time.Time
	email, code [32]byte
	codeExpires time.Time
	attempts    int
	verified    bool
}
type cooldown struct {
	since, last time.Time
	count       int
}
type Service struct {
	config     Config
	mu         sync.Mutex
	sessions   map[[32]byte]*session
	mailLimits map[[32]byte]cooldown
	salt       string
	now        func() time.Time
	mailWindow time.Time
	mailCount  int
}

func New(config Config) *Service {
	return &Service{config: config, sessions: make(map[[32]byte]*session), mailLimits: make(map[[32]byte]cooldown), salt: rand.Text(), now: time.Now}
}
func (s *Service) enabled() bool                   { return s.config.SiteKey != "" && s.config.CheckHuman != nil }
func digest(text string) [32]byte                  { return sha256.Sum256([]byte(text)) }
func (s *Service) emailHash(email string) [32]byte { return digest(s.salt + email) }
func canonicalEmail(value string) (string, error) {
	value = strings.TrimSpace(value)
	if len(value) > 254 || strings.ContainsAny(value, "\r\n\x00") {
		return "", errors.New("invalid email")
	}
	address, err := mail.ParseAddress(value)
	if err != nil || address.Name != "" || address.Address != value || !strings.Contains(value, "@") {
		return "", errors.New("invalid email")
	}
	index := strings.LastIndexByte(value, '@')
	return value[:index+1] + strings.ToLower(value[index+1:]), nil
}
func (s *Service) prune() {
	for key, value := range s.sessions {
		if !s.now().Before(value.expires) {
			delete(s.sessions, key)
		}
	}
	for key, value := range s.mailLimits {
		if s.now().Sub(value.since) >= time.Hour {
			delete(s.mailLimits, key)
		}
	}
}
func (s *Service) lookup(r *http.Request) (*session, string) {
	cookie, err := r.Cookie("cq_form_session")
	if err != nil || len(cookie.Value) > 128 {
		return nil, ""
	}
	entry := s.sessions[digest(cookie.Value)]
	if entry == nil || !s.now().Before(entry.expires) {
		return nil, ""
	}
	return entry, cookie.Value
}
func (s *Service) HasHuman(r *http.Request) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, _ := s.lookup(r)
	return entry != nil
}

// RequireVerifiedEmail must be called by every future broker-delivery handler.
// The submitted email is bound to a server-confirmed session, never a UI flag.
func (s *Service) RequireVerifiedEmail(r *http.Request, email string) bool {
	value, err := canonicalEmail(email)
	if err != nil {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, _ := s.lookup(r)
	return entry != nil && entry.verified && entry.email == s.emailHash(value)
}
func (s *Service) ProtectAddress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.enabled() && !s.HasHuman(r) {
			http.Error(w, "Complete the bot check first.", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
func jsonReply(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func fail(w http.ResponseWriter, status int) { jsonReply(w, status, map[string]bool{"ok": false}) }

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if r.URL.Path == "/api/v1/form-security/config" && r.Method == http.MethodGet && r.URL.RawQuery == "" {
		jsonReply(w, http.StatusOK, map[string]any{"botEnabled": s.enabled(), "humanVerified": s.enabled() && s.HasHuman(r), "emailEnabled": s.enabled() && s.config.SendCode != nil, "siteKey": s.config.SiteKey})
		return
	}
	if r.Method != http.MethodPost {
		fail(w, http.StatusMethodNotAllowed)
		return
	}
	origin := r.Header.Get("Origin")
	if r.URL.RawQuery != "" || (origin != "https://abcw.xyz:8443" && origin != "http://127.0.0.1:8787") {
		fail(w, http.StatusForbidden)
		return
	}
	if !s.enabled() {
		fail(w, http.StatusServiceUnavailable)
		return
	}
	if strings.Split(r.Header.Get("Content-Type"), ";")[0] != "application/json" {
		fail(w, http.StatusUnsupportedMediaType)
		return
	}
	var input struct {
		Email   string `json:"email"`
		Token   string `json:"token"`
		Code    string `json:"code"`
		Website string `json:"website"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&input) != nil || decoder.Decode(&struct{}{}) != io.EOF || input.Website != "" {
		fail(w, http.StatusBadRequest)
		return
	}
	if r.URL.Path == "/api/v1/form-security/session" {
		if len(input.Token) == 0 || len(input.Token) > 2048 || input.Email != "" || input.Code != "" {
			fail(w, http.StatusBadRequest)
			return
		}
		if s.config.CheckHuman(r.Context(), input.Token) != nil {
			fail(w, http.StatusForbidden)
			return
		}
		s.mu.Lock()
		s.prune()
		if len(s.sessions) >= 2048 {
			s.mu.Unlock()
			fail(w, http.StatusTooManyRequests)
			return
		}
		token := rand.Text()
		expires := s.now().Add(30 * time.Minute)
		s.sessions[digest(token)] = &session{expires: expires}
		s.mu.Unlock()
		http.SetCookie(w, &http.Cookie{Name: "cq_form_session", Value: token, Path: "/api/v1/", Secure: r.TLS != nil, HttpOnly: true, SameSite: http.SameSiteStrictMode, MaxAge: 1800})
		jsonReply(w, http.StatusOK, map[string]bool{"ok": true})
		return
	}
	if input.Token != "" {
		fail(w, http.StatusBadRequest)
		return
	}
	address, err := canonicalEmail(input.Email)
	if err != nil {
		fail(w, http.StatusBadRequest)
		return
	}
	emailHash := s.emailHash(address)
	s.mu.Lock()
	s.prune()
	entry, token := s.lookup(r)
	if entry == nil {
		s.mu.Unlock()
		fail(w, http.StatusForbidden)
		return
	}
	switch r.URL.Path {
	case "/api/v1/form-security/email/status":
		verified := entry.verified && entry.email == emailHash
		s.mu.Unlock()
		jsonReply(w, http.StatusOK, map[string]bool{"verified": verified})
	case "/api/v1/form-security/email/confirm":
		if len(input.Code) != 6 || strings.Trim(input.Code, "0123456789") != "" || !s.now().Before(entry.codeExpires) || entry.attempts >= 5 || entry.email != emailHash || entry.verified {
			s.mu.Unlock()
			fail(w, http.StatusUnauthorized)
			return
		}
		entry.attempts++
		candidate := digest(token + ":" + input.Code)
		valid := subtle.ConstantTimeCompare(candidate[:], entry.code[:]) == 1
		if valid {
			entry.verified = true
			entry.code = [32]byte{}
		}
		s.mu.Unlock()
		if !valid {
			fail(w, http.StatusUnauthorized)
			return
		}
		jsonReply(w, http.StatusOK, map[string]bool{"verified": true})
	case "/api/v1/form-security/email/start":
		if s.config.SendCode == nil {
			s.mu.Unlock()
			fail(w, http.StatusServiceUnavailable)
			return
		}
		if input.Code != "" {
			s.mu.Unlock()
			fail(w, http.StatusBadRequest)
			return
		}
		if s.now().Sub(s.mailWindow) >= time.Minute {
			s.mailWindow = s.now()
			s.mailCount = 0
		}
		limitHash := digest(s.salt + "quota:" + strings.ToLower(address))
		limit, known := s.mailLimits[limitHash]
		if s.mailCount >= 20 || (known && (s.now().Sub(limit.last) < time.Minute || limit.count >= 3)) || (!known && len(s.mailLimits) >= 2048) {
			s.mu.Unlock()
			w.Header().Set("Retry-After", "60")
			fail(w, http.StatusTooManyRequests)
			return
		}
		if !known {
			limit.since = s.now()
		}
		limit.last = s.now()
		limit.count++
		s.mailLimits[limitHash] = limit
		s.mailCount++
		number, randomErr := rand.Int(rand.Reader, big.NewInt(1000000))
		if randomErr != nil {
			s.mu.Unlock()
			fail(w, http.StatusServiceUnavailable)
			return
		}
		code := number.String()
		code = strings.Repeat("0", 6-len(code)) + code
		entry.email = emailHash
		entry.code = digest(token + ":" + code)
		entry.codeExpires = s.now().Add(10 * time.Minute)
		entry.attempts = 0
		entry.verified = false
		s.mu.Unlock()
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		if s.config.SendCode(ctx, address, code) != nil {
			s.mu.Lock()
			if entry.code == digest(token+":"+code) {
				entry.code = [32]byte{}
			}
			s.mu.Unlock()
			fail(w, http.StatusServiceUnavailable)
			return
		}
		jsonReply(w, http.StatusAccepted, map[string]bool{"ok": true})
	default:
		s.mu.Unlock()
		fail(w, http.StatusNotFound)
	}
}
