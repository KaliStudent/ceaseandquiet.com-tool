// Package preview serves the local beta without file browsing or personal intake.
package preview

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"time"

	previewsite "github.com/drumandbytes/eraser/ceaseandquiet/site"
	"github.com/drumandbytes/eraser/internal/readonlyapi"
)

func New(addr string) (http.Handler, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil || port == "" || (host != "localhost" && (net.ParseIP(host) == nil || !net.ParseIP(host).IsLoopback())) {
		return nil, fmt.Errorf("preview must bind to a loopback address")
	}
	api, err := readonlyapi.New()
	if err != nil {
		return nil, err
	}
	mux := http.NewServeMux()
	mux.Handle("/api/", api.Handler())
	for _, name := range []string{"index.html", "styles.css", "site.js"} {
		content, err := previewsite.Files.ReadFile(name)
		if err != nil {
			return nil, err
		}
		path := "/" + name
		if name == "index.html" {
			path = "/"
		}
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != path {
				http.NotFound(w, r)
				return
			}
			if r.URL.RawQuery != "" {
				http.Error(w, "unexpected query", http.StatusBadRequest)
				return
			}
			http.ServeContent(w, r, name, time.Time{}, bytes.NewReader(content))
		})
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; script-src 'self'; style-src 'self'; connect-src 'self'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'; object-src 'none'")
		if r.Host != addr {
			http.Error(w, "unexpected host", http.StatusMisdirectedRequest)
			return
		}
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "read-only preview", http.StatusMethodNotAllowed)
			return
		}
		mux.ServeHTTP(w, r)
	}), nil
}
