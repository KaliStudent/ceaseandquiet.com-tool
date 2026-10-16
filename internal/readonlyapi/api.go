// Package readonlyapi exposes public broker directory metadata only.
// It deliberately has no mail, profile, history, or delivery dependency.
package readonlyapi

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/drumandbytes/eraser/data"
	"github.com/drumandbytes/eraser/internal/broker"
)

type Item struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Region        string `json:"region"`
	Category      string `json:"category,omitempty"`
	ChannelStatus string `json:"channelStatus"`
}

type API struct{ items []Item }

func New() (*API, error) {
	db, err := broker.Parse(data.BrokersYAML)
	if err != nil {
		return nil, err
	}
	a := &API{items: make([]Item, 0, len(db.Brokers))}
	for _, b := range db.Brokers {
		a.items = append(a.items, Item{ID: b.ID, Name: b.Name, Region: b.Region, Category: b.Category, ChannelStatus: "unverified"})
	}
	sort.Slice(a.items, func(i, j int) bool { return a.items[i].Name < a.items[j].Name })
	return a, nil
}

func (a *API) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/brokers", a.list)
	mux.HandleFunc("GET /api/v1/brokers/{id}", a.one)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Cache-Control", "no-store")
		mux.ServeHTTP(w, r)
	})
}

func respond(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (a *API) list(w http.ResponseWriter, r *http.Request) {
	limit := 25
	if raw := r.URL.Query().Get("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 || n > 100 {
			respond(w, 400, map[string]string{"error": "limit must be 1–100"})
			return
		}
		limit = n
	}
	query := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	if len(query) > 100 {
		respond(w, 400, map[string]string{"error": "query too long"})
		return
	}
	items := make([]Item, 0, limit)
	for _, item := range a.items {
		if query != "" && !strings.Contains(strings.ToLower(item.Name), query) {
			continue
		}
		items = append(items, item)
		if len(items) == limit {
			break
		}
	}
	respond(w, 200, map[string]any{"items": items, "count": len(items), "notice": "Directory entries and contact channels have not been independently verified."})
}

func (a *API) one(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	for _, item := range a.items {
		if item.ID == id {
			respond(w, 200, item)
			return
		}
	}
	respond(w, 404, map[string]string{"error": "broker not found"})
}
