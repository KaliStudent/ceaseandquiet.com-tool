// cq-preview is a local-only, non-sending preview of the Cease & Quiet site.
package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"time"

	"github.com/drumandbytes/eraser/internal/readonlyapi"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8787", "local listen address")
	site := flag.String("site", filepath.Join("ceaseandquiet", "site"), "staged static site directory")
	flag.Parse()
	host, _, err := net.SplitHostPort(*addr)
	if err != nil || (host != "localhost" && (net.ParseIP(host) == nil || !net.ParseIP(host).IsLoopback())) {
		log.Fatal("preview must bind to a loopback address")
	}
	api, err := readonlyapi.New()
	if err != nil {
		log.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.Handle("/api/", api.Handler())
	mux.Handle("/", http.FileServer(http.Dir(*site)))
	server := &http.Server{Addr: *addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	log.Printf("Cease & Quiet local preview: http://%s", *addr)
	log.Fatal(server.ListenAndServe())
}
