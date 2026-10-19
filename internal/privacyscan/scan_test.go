package privacyscan

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type testAuth struct{ human, verified bool }

func (a testAuth) HasHuman(*http.Request) bool { return a.human }
func (a testAuth) RequireVerifiedEmail(_ *http.Request, email string) bool {
	return a.verified && email == "taylor@example.test"
}
func sample() Submission {
	return Submission{Profile: Profile{GivenName: "Taylor", FamilyName: "Example", Email: "taylor@example.test", Address: "100 Example Street", City: "Example City", Region: "CA", PostalCode: "00000"}, Country: "US", USResident: true, OwnInformation: true, AuthorizeScanAndRemoval: true, ConsentVersion: ConsentVersion}
}
func submit(h *Handler, input Submission) *httptest.ResponseRecorder {
	body, _ := json.Marshal(input)
	r := httptest.NewRequest(http.MethodPost, "/api/v1/privacy-scan", strings.NewReader(string(body)))
	r.Header.Set("Origin", "https://abcw.xyz:8443")
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}
func TestUSResidencyOwnershipAndBothActionsAreRequired(t *testing.T) {
	calls := 0
	h := &Handler{Auth: testAuth{true, true}, Start: func(_ context.Context, input Submission) (string, error) {
		calls++
		if !input.AuthorizeScanAndRemoval {
			t.Fatal("consent bypass")
		}
		return "synthetic-test-job", nil
	}}
	for _, change := range []func(*Submission){func(s *Submission) { s.Country = "CA" }, func(s *Submission) { s.USResident = false }, func(s *Submission) { s.OwnInformation = false }, func(s *Submission) { s.AuthorizeScanAndRemoval = false }, func(s *Submission) { s.ConsentVersion = "old" }, func(s *Submission) { s.Region = "ON" }, func(s *Submission) { s.PostalCode = "foreign" }, func(s *Submission) { s.Website = "spam" }, func(s *Submission) { s.GivenName = strings.Repeat("a", 101) }} {
		input := sample()
		change(&input)
		if w := submit(h, input); w.Code != http.StatusBadRequest {
			t.Fatalf("invalid submission status %d", w.Code)
		}
	}
	if calls != 0 {
		t.Fatal("invalid request reached workflow")
	}
	if w := submit(h, sample()); w.Code != http.StatusAccepted || calls != 1 {
		t.Fatal("authorized workflow did not start")
	}
}
func TestUnverifiedOrUnavailableWorkflowCannotClaimResults(t *testing.T) {
	calls := 0
	h := &Handler{Auth: testAuth{true, false}, Start: func(context.Context, Submission) (string, error) { calls++; return "synthetic", nil }}
	if w := submit(h, sample()); w.Code != http.StatusForbidden || calls != 0 {
		t.Fatal("unverified email reached workflow")
	}
	h.Auth = testAuth{false, true}
	if w := submit(h, sample()); w.Code != http.StatusForbidden || calls != 0 {
		t.Fatal("bot check bypassed")
	}
	h.Auth = testAuth{true, true}
	h.Start = nil
	w := submit(h, sample())
	if w.Code != http.StatusServiceUnavailable || strings.Contains(w.Body.String(), "jobId") {
		t.Fatal("unavailable workflow fabricated a scan")
	}
}
