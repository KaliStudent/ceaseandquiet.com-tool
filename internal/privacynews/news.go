// Package privacynews fetches public privacy headlines from a fixed RSS source.
package privacynews

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"
)

type Article struct {
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	Published time.Time `json:"published"`
	Source    string    `json:"source"`
}
type Handler struct {
	mu                  sync.Mutex
	endpoint            string
	client              *http.Client
	items               []Article
	updated, retryAfter time.Time
	now                 func() time.Time
}

func New() *Handler {
	return &Handler{endpoint: "https://www.eff.org/rss/updates.xml", client: &http.Client{Timeout: 5 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}, now: time.Now}
}
func (h *Handler) fetch() error {
	request, err := http.NewRequest(http.MethodGet, h.endpoint, nil)
	if err != nil {
		return err
	}
	request.Header.Set("User-Agent", "CeaseAndQuiet-PrivacyNews/1.0")
	response, err := h.client.Do(request)
	if err != nil {
		return errors.New("news unavailable")
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return errors.New("news unavailable")
	}
	var feed struct {
		Channel struct {
			Items []struct {
				Title       string `xml:"title"`
				Link        string `xml:"link"`
				Date        string `xml:"pubDate"`
				Description string `xml:"description"`
			} `xml:"item"`
		} `xml:"channel"`
	}
	if xml.NewDecoder(io.LimitReader(response.Body, 2<<20)).Decode(&feed) != nil {
		return errors.New("invalid news feed")
	}
	items := make([]Article, 0, 6)
	seen := map[string]bool{}
	for _, item := range feed.Channel.Items {
		link, err := url.Parse(strings.TrimSpace(item.Link))
		if err != nil || link.Scheme != "https" || (link.Host != "www.eff.org" && link.Host != "eff.org") || link.User != nil || link.RawQuery != "" || link.Fragment != "" {
			continue
		}
		title := strings.Join(strings.Fields(item.Title), " ")
		if title == "" || len(title) > 300 || seen[link.String()] {
			continue
		}
		published, err := time.Parse(time.RFC1123Z, item.Date)
		if err != nil {
			published, err = time.Parse(time.RFC1123, item.Date)
		}
		if err != nil || published.After(h.now().Add(24*time.Hour)) {
			continue
		}
		text := strings.ToLower(title + " " + item.Description)
		relevant := false
		for _, word := range []string{"privacy", "data broker", "surveillance", "personal data", "tracking", "data breach", "biometric", "facial recognition", "advertising"} {
			if strings.Contains(text, word) {
				relevant = true
				break
			}
		}
		if !relevant {
			continue
		}
		seen[link.String()] = true
		items = append(items, Article{title, link.String(), published.UTC(), "Electronic Frontier Foundation"})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Published.After(items[j].Published) })
	if len(items) > 6 {
		items = items[:6]
	}
	if len(items) == 0 {
		return errors.New("no privacy headlines")
	}
	h.items = items
	h.updated = h.now()
	return nil
}
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if r.URL.RawQuery != "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.now().After(h.retryAfter) {
		h.retryAfter = h.now().Add(time.Minute)
		if h.fetch() == nil {
			h.retryAfter = h.now().Add(15 * time.Minute)
		}
	}
	if len(h.items) == 0 {
		w.WriteHeader(http.StatusServiceUnavailable)
		if r.Method != http.MethodHead {
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "news temporarily unavailable"})
		}
		return
	}
	if r.Method == http.MethodHead {
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"items": h.items, "updatedAt": h.updated, "stale": h.now().Sub(h.updated) > 15*time.Minute, "sourceURL": "https://www.eff.org/updates"})
}
