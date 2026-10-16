// cq-staging serves the synthetic-data preview behind an explicit TLS/auth gate.
package main

import (
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/drumandbytes/eraser/internal/staging"
)

func main() {
	host := os.Getenv("CQ_STAGING_HOST")
	_, port, err := net.SplitHostPort(host)
	if err != nil {
		log.Fatal("CQ_STAGING_HOST must specify the approved hostname:port")
	}
	digest, err := os.ReadFile(os.Getenv("CQ_STAGING_PASSWORD_SHA256_FILE"))
	if err != nil {
		log.Fatal("cannot read staging credential digest")
	}
	decoded, err := hex.DecodeString(strings.TrimSpace(string(digest)))
	if err != nil || len(decoded) != sha256.Size {
		log.Fatal("invalid staging credential digest")
	}
	var hash [sha256.Size]byte
	copy(hash[:], decoded)
	handler, err := staging.New(host, os.Getenv("CQ_STAGING_USER"), hash)
	if err != nil {
		log.Fatal(err)
	}
	cert, key := os.Getenv("CQ_STAGING_TLS_CERT_FILE"), os.Getenv("CQ_STAGING_TLS_KEY_FILE")
	if cert == "" || key == "" {
		log.Fatal("explicit trusted TLS certificate and key paths required")
	}
	server := &http.Server{
		Addr: ":" + port, Handler: handler,
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second,
		WriteTimeout: 20 * time.Second, IdleTimeout: 30 * time.Second,
		MaxHeaderBytes: 16384, TLSConfig: &tls.Config{MinVersion: tls.VersionTLS12},
	}
	log.Print("Cease and Quiet restricted staging starting; broker delivery disabled")
	log.Fatal(server.ListenAndServeTLS(cert, key))
}
