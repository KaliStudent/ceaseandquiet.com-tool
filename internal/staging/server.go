// Package staging adds a TLS-only access gate to the read-only preview.
package staging

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/drumandbytes/eraser/internal/preview"
)

// New requires one explicit public hostname and port. It exposes no sender or
// personal-data endpoints; credentials are checked before serving any asset/API.
func New(host, username string, passwordHash [sha256.Size]byte) (http.Handler, error) {
	name, port, err := net.SplitHostPort(host)
	n, portErr := strconv.Atoi(port)
	if err != nil || portErr != nil || n < 1 || n > 65535 || name == "" || strings.ContainsAny(name, "/\\@?# \t\r\n") || username == "" || passwordHash == [sha256.Size]byte{} {
		return nil, fmt.Errorf("explicit hostname:port and staging credentials required")
	}
	app, err := preview.New("127.0.0.1:8787")
	if err != nil {
		return nil, err
	}
	userHash := sha256.Sum256([]byte(username))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Robots-Tag", "noindex, nofollow, noarchive")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'; form-action 'none'")
		if r.TLS == nil || r.Host != host {
			http.Error(w, "unexpected staging destination", http.StatusMisdirectedRequest)
			return
		}
		user, password, ok := r.BasicAuth()
		u := sha256.Sum256([]byte(user))
		p := sha256.Sum256([]byte(password))
		valid := subtle.ConstantTimeCompare(u[:], userHash[:]) & subtle.ConstantTimeCompare(p[:], passwordHash[:])
		if !ok || valid != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="Cease and Quiet staging", charset="UTF-8"`)
			http.Error(w, "staging access required", http.StatusUnauthorized)
			return
		}
		// The exact external Host was checked above. Adapt the request for the
		// existing loopback handler in memory, without an exposed upstream port.
		local := r.Clone(r.Context())
		local.Host = "127.0.0.1:8787"
		local.Header.Del("Authorization")
		app.ServeHTTP(w, local)
	}), nil
}
