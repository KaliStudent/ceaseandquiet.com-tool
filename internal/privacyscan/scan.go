// Package privacyscan enforces the US consumer scan/removal authorization boundary.
// A broker directory is never treated as evidence that a person was found.
package privacyscan

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/mail"
	"regexp"
	"strings"
	"time"
)

const ConsentVersion = "us-scan-removal-v1"

var zipPattern = regexp.MustCompile(`^\d{5}(-\d{4})?$`)

type Profile struct {
	GivenName   string `json:"givenName"`
	FamilyName  string `json:"familyName"`
	Email       string `json:"email"`
	Address     string `json:"address"`
	City        string `json:"city"`
	Region      string `json:"region"`
	PostalCode  string `json:"postalCode"`
	DateOfBirth string `json:"dateOfBirth"`
	Phone       string `json:"phone"`
}
type Submission struct {
	Profile
	Country                 string `json:"country"`
	USResident              bool   `json:"usResident"`
	OwnInformation          bool   `json:"ownInformation"`
	AuthorizeScanAndRemoval bool   `json:"authorizeScanAndRemoval"`
	ConsentVersion          string `json:"consentVersion"`
	Website                 string `json:"website"`
}
type Authorizer interface {
	HasHuman(*http.Request) bool
	RequireVerifiedEmail(*http.Request, string) bool
}
type Handler struct {
	Auth Authorizer
	// Start must perform the authorized presence scan and removal workflow.
	// There is no production adapter yet; nil means unavailable, never a fake job.
	Start func(context.Context, Submission) (string, error)
}

func valid(p Profile) bool {
	for _, field := range []struct {
		value string
		max   int
	}{{p.GivenName, 100}, {p.FamilyName, 100}, {p.Email, 254}, {p.Address, 200}, {p.City, 100}, {p.Region, 2}, {p.PostalCode, 10}, {p.DateOfBirth, 10}, {p.Phone, 32}} {
		if len(field.value) > field.max || strings.IndexFunc(field.value, func(c rune) bool { return c < 32 || c == 127 }) >= 0 {
			return false
		}
	}
	if strings.TrimSpace(p.GivenName) == "" || strings.TrimSpace(p.FamilyName) == "" || strings.TrimSpace(p.Address) == "" || strings.TrimSpace(p.City) == "" || !zipPattern.MatchString(p.PostalCode) {
		return false
	}
	address, err := mail.ParseAddress(p.Email)
	if err != nil || address.Address != p.Email || address.Name != "" {
		return false
	}
	if !strings.Contains(" AL AK AZ AR CA CO CT DE DC FL GA HI ID IL IN IA KS KY LA ME MD MA MI MN MS MO MT NE NV NH NJ NM NY NC ND OH OK OR PA RI SC SD TN TX UT VT VA WA WV WI WY AS GU MP PR VI ", " "+p.Region+" ") {
		return false
	}
	if p.DateOfBirth != "" {
		date, err := time.Parse("2006-01-02", p.DateOfBirth)
		if err != nil || date.After(time.Now()) {
			return false
		}
	}
	if p.Phone != "" {
		digits := 0
		for _, c := range p.Phone {
			if c >= '0' && c <= '9' {
				digits++
			} else if !strings.ContainsRune(" +().-", c) {
				return false
			}
		}
		if digits < 7 || digits > 15 {
			return false
		}
	}
	return true
}
func reply(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.RawQuery != "" {
		reply(w, http.StatusBadRequest, map[string]string{"error": "query not accepted"})
		return
	}
	if r.URL.Path == "/api/v1/privacy-scan/status" && r.Method == http.MethodGet {
		reply(w, http.StatusOK, map[string]any{"available": h.Start != nil, "consentVersion": ConsentVersion, "country": "US"})
		return
	}
	if r.URL.Path != "/api/v1/privacy-scan" || r.Method != http.MethodPost {
		reply(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not supported"})
		return
	}
	if origin := r.Header.Get("Origin"); origin != "https://abcw.xyz:8443" && origin != "http://127.0.0.1:8787" {
		reply(w, http.StatusForbidden, map[string]string{"error": "origin rejected"})
		return
	}
	if h.Auth == nil || !h.Auth.HasHuman(r) {
		reply(w, http.StatusForbidden, map[string]string{"error": "complete the bot check"})
		return
	}
	// Do not intake personal details while there is no real scan/removal adapter.
	if h.Start == nil {
		reply(w, http.StatusServiceUnavailable, map[string]string{"error": "Scanning and removal requests are not yet available. Nothing was submitted."})
		return
	}
	if strings.Split(r.Header.Get("Content-Type"), ";")[0] != "application/json" {
		reply(w, http.StatusUnsupportedMediaType, map[string]string{"error": "JSON required"})
		return
	}
	var input Submission
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&input) != nil || decoder.Decode(&struct{}{}) != io.EOF || input.Website != "" || !valid(input.Profile) {
		reply(w, http.StatusBadRequest, map[string]string{"error": "check your profile details"})
		return
	}
	if input.Country != "US" || !input.USResident || !input.OwnInformation || !input.AuthorizeScanAndRemoval || input.ConsentVersion != ConsentVersion {
		reply(w, http.StatusBadRequest, map[string]string{"error": "US residency and authorization for both scanning and removal are required"})
		return
	}
	if !h.Auth.RequireVerifiedEmail(r, input.Email) {
		reply(w, http.StatusForbidden, map[string]string{"error": "verify this email address first"})
		return
	}
	job, err := h.Start(r.Context(), input)
	if err != nil || job == "" {
		reply(w, http.StatusServiceUnavailable, map[string]string{"error": "The scan could not be started. No successful submission is being reported."})
		return
	}
	reply(w, http.StatusAccepted, map[string]string{"jobId": job, "status": "accepted"})
}
