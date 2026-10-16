package readonlyapi

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestDirectoryPaginationAndFilters(t *testing.T) {
	api, err := New()
	if err != nil {
		t.Fatal(err)
	}
	type result struct {
		Items   []Item
		Total   int
		Offset  int
		HasMore bool
	}
	read := func(path string) result {
		t.Helper()
		w := httptest.NewRecorder()
		api.Handler().ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 200 {
			t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
		}
		var response result
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatal(err)
		}
		return response
	}
	first := read("/api/v1/brokers?limit=12")
	second := read("/api/v1/brokers?limit=12&offset=12")
	if len(first.Items) != 12 || !first.HasMore || first.Total != 723 || second.Offset != 12 {
		t.Fatalf("invalid pagination: %+v", first)
	}
	seen := map[string]bool{}
	for _, item := range first.Items {
		seen[item.ID] = true
	}
	for _, item := range second.Items {
		if seen[item.ID] {
			t.Fatal("duplicate across pages")
		}
	}
	filtered := read("/api/v1/brokers?region=eu&limit=100")
	for _, item := range filtered.Items {
		if item.Region != "eu" {
			t.Fatal("region filter leaked another region")
		}
	}
	empty := read("/api/v1/brokers?q=synthetic-no-match-12345")
	if len(empty.Items) != 0 || empty.Total != 0 || empty.HasMore {
		t.Fatal("invalid empty result")
	}
	for _, path := range []string{"/api/v1/brokers?offset=-1", "/api/v1/brokers?offset=99999", "/api/v1/brokers?region=unknown", "/api/v1/brokers?limit=0"} {
		w := httptest.NewRecorder()
		api.Handler().ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 400 {
			t.Fatalf("%s status=%d", path, w.Code)
		}
	}
}
