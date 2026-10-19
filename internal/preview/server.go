// Package preview serves the local beta without file browsing or full-profile intake.
package preview

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	cqassets "github.com/drumandbytes/eraser/ceaseandquiet"
	previewsite "github.com/drumandbytes/eraser/ceaseandquiet/site"
	"github.com/drumandbytes/eraser/internal/addressapi"
	"github.com/drumandbytes/eraser/internal/formsecurity"
	"github.com/drumandbytes/eraser/internal/privacynews"
	"github.com/drumandbytes/eraser/internal/privacyscan"
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
	formSecurity := formsecurity.FromEnvironment()
	mux.Handle("/api/v1/form-security/", formSecurity)
	mux.Handle("/api/", api.Handler())
	mux.Handle("/api/v1/privacy-news", privacynews.New())
	scan := &privacyscan.Handler{Auth: formSecurity}
	mux.Handle("/api/v1/privacy-scan", scan)
	mux.Handle("/api/v1/privacy-scan/status", scan)
	mux.Handle("/api/v1/address-suggestions", formSecurity.ProtectAddress(addressapi.New(os.Getenv("CQ_GEOAPIFY_API_KEY"))))
	for _, name := range []string{"index.html", "privacy-scan.html", "consumer-protection.html", "styles.css", "site.js", "home.js", "scan.js", "consumer.js", "affiliate-links.json", "scan-workflow.mjs", "address.mjs", "form-tools.mjs", "vault.mjs", "profile-vault.mjs", "form-security.mjs"} {
		var content []byte
		var err error
		if name == "address.mjs" || name == "vault.mjs" {
			content, err = cqassets.Files.ReadFile("extension/" + name)
		} else {
			content, err = previewsite.Files.ReadFile(name)
		}
		if err != nil {
			return nil, err
		}
		path := "/" + name
		if name == "index.html" {
			path = "/"
		}
		if name == "privacy-scan.html" {
			path = "/privacy-scan/"
		}
		if name == "consumer-protection.html" {
			path = "/consumer-protection/"
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
	for _, path := range []string{"/privacy-scan", "/consumer-protection"} {
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.RawQuery != "" {
				http.Error(w, "unexpected query", http.StatusBadRequest)
				return
			}
			http.Redirect(w, r, path+"/", http.StatusPermanentRedirect)
		})
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; script-src 'self'; style-src 'self'; connect-src 'self'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'; object-src 'none'")
		if os.Getenv("CQ_TURNSTILE_SITE_KEY") != "" && os.Getenv("CQ_TURNSTILE_SECRET") != "" {
			w.Header().Set("Content-Security-Policy", "default-src 'none'; script-src 'self' https://challenges.cloudflare.com; style-src 'self'; connect-src 'self'; frame-src https://challenges.cloudflare.com; base-uri 'none'; form-action 'none'; frame-ancestors 'none'; object-src 'none'")
		}
		if r.Host != addr {
			http.Error(w, "unexpected host", http.StatusMisdirectedRequest)
			return
		}
		addressLookup := r.URL.Path == "/api/v1/address-suggestions" && r.Method == http.MethodPost
		securityAction := strings.HasPrefix(r.URL.Path, "/api/v1/form-security/") && r.Method == http.MethodPost
		scanSubmission := r.URL.Path == "/api/v1/privacy-scan" && r.Method == http.MethodPost
		if r.Method != http.MethodGet && r.Method != http.MethodHead && !addressLookup && !securityAction && !scanSubmission {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "read-only preview", http.StatusMethodNotAllowed)
			return
		}
		mux.ServeHTTP(w, r)
	}), nil
}
