// Package addressapi proxies explicitly requested partial-address suggestions.
// It does not accept or persist a personal profile and never logs queries/keys.
package addressapi

import (
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type Handler struct {
	key      string
	endpoint string
	client   *http.Client
	mu       sync.Mutex
	window   time.Time
	count    int
}

func New(key string) *Handler {
	return &Handler{key: key, endpoint: "https://api.geoapify.com/v1/geocode/autocomplete", client: &http.Client{Timeout: 4 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
}

type suggestion struct {
	Label      string `json:"label"`
	Address    string `json:"address"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postalCode"`
}

func clean(s string, max int) bool { return len(s) <= max && !strings.ContainsAny(s, "\r\n\x00") }

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"POST required"}`, http.StatusMethodNotAllowed)
		return
	}
	if r.URL.RawQuery != "" {
		http.Error(w, `{"error":"query parameters not accepted"}`, http.StatusBadRequest)
		return
	}
	origin := r.Header.Get("Origin")
	if origin != "https://abcw.xyz:8443" && origin != "http://127.0.0.1:8787" {
		http.Error(w, `{"error":"origin rejected"}`, http.StatusForbidden)
		return
	}
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		http.Error(w, `{"error":"JSON required"}`, http.StatusUnsupportedMediaType)
		return
	}
	var input struct {
		Text string `json:"text"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		http.Error(w, `{"error":"invalid address query"}`, http.StatusBadRequest)
		return
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		http.Error(w, `{"error":"invalid address query"}`, http.StatusBadRequest)
		return
	}
	input.Text = strings.TrimSpace(input.Text)
	if len(input.Text) < 4 || !clean(input.Text, 200) {
		http.Error(w, `{"error":"enter 4-200 address characters"}`, http.StatusBadRequest)
		return
	}
	if h.key == "" {
		http.Error(w, `{"error":"suggestions not configured; enter address manually"}`, http.StatusServiceUnavailable)
		return
	}
	h.mu.Lock()
	if time.Since(h.window) >= time.Minute {
		h.window = time.Now()
		h.count = 0
	}
	limited := h.count >= 60
	if !limited {
		h.count++
	}
	h.mu.Unlock()
	if limited {
		w.Header().Set("Retry-After", "60")
		http.Error(w, `{"error":"try again later or enter manually"}`, http.StatusTooManyRequests)
		return
	}
	query := url.Values{"text": {input.Text}, "format": {"json"}, "limit": {"5"}, "apiKey": {h.key}, "filter": {"countrycode:us"}}
	request, err := http.NewRequestWithContext(r.Context(), "GET", h.endpoint+"?"+query.Encode(), nil)
	if err != nil {
		http.Error(w, `{"error":"suggestions unavailable"}`, http.StatusBadGateway)
		return
	}
	response, err := h.client.Do(request)
	if err != nil {
		http.Error(w, `{"error":"suggestions unavailable"}`, http.StatusBadGateway)
		return
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		http.Error(w, `{"error":"suggestions unavailable"}`, http.StatusBadGateway)
		return
	}
	var data struct {
		Results []struct {
			Formatted   string `json:"formatted"`
			HouseNumber string `json:"housenumber"`
			Street      string `json:"street"`
			City        string `json:"city"`
			State       string `json:"state"`
			StateCode   string `json:"state_code"`
			Postcode    string `json:"postcode"`
			CountryCode string `json:"country_code"`
		} `json:"results"`
	}
	if json.NewDecoder(io.LimitReader(response.Body, 65536)).Decode(&data) != nil {
		http.Error(w, `{"error":"suggestions unavailable"}`, http.StatusBadGateway)
		return
	}
	items := make([]suggestion, 0, 5)
	for _, result := range data.Results {
		if strings.ToLower(result.CountryCode) != "us" {
			continue
		}
		state := result.StateCode
		if state == "" {
			state = result.State
		}
		street := strings.TrimSpace(result.HouseNumber + " " + result.Street)
		if street == "" || strings.TrimSpace(result.Formatted) == "" || !clean(result.Formatted, 500) || !clean(street, 200) || !clean(result.City, 100) || !clean(state, 100) || !clean(result.Postcode, 20) {
			continue
		}
		items = append(items, suggestion{result.Formatted, street, result.City, state, result.Postcode})
		if len(items) == 5 {
			break
		}
	}
	_ = json.NewEncoder(w).Encode(struct {
		Items []suggestion `json:"items"`
	}{items})
}
