package readonlyapi

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPublicAPIExposesNoContactOrPersonalFields(t *testing.T) {
	a, err := New()
	if err != nil {
		t.Fatal(err)
	}
	if len(a.items) != 723 {
		t.Fatalf("embedded directory count = %d, want 723", len(a.items))
	}
	for _, path := range []string{"/api/v1/brokers?limit=2", "/api/v1/brokers?limit=101", "/api/v1/brokers/missing"} {
		r := httptest.NewRecorder()
		a.Handler().ServeHTTP(r, httptest.NewRequest("GET", path, nil))
		if strings.Contains(r.Body.String(), "@") || strings.Contains(r.Body.String(), "opt_out_url") {
			t.Fatalf("contact field leaked: %s", path)
		}
	}
	r := httptest.NewRecorder()
	a.Handler().ServeHTTP(r, httptest.NewRequest("POST", "/api/v1/brokers", nil))
	if r.Code != 405 {
		t.Fatalf("POST status = %d, want 405", r.Code)
	}
	r = httptest.NewRecorder()
	a.Handler().ServeHTTP(r, httptest.NewRequest("GET", "/api/v1/brokers?q=spokeo", nil))
	if r.Code != 200 || !strings.Contains(r.Body.String(), `"name":"Spokeo"`) || strings.Contains(r.Body.String(), `"email"`) {
		t.Fatalf("unexpected filtered response: %s", r.Body.String())
	}
}
