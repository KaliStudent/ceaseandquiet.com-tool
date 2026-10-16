// cq-preview is a local-only, non-sending preview of the Cease & Quiet site.
package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/drumandbytes/eraser/internal/preview"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8787", "local listen address")
	flag.Parse()
	handler, err := preview.New(*addr)
	if err != nil {
		log.Fatal(err)
	}
	server := &http.Server{Addr: *addr, Handler: handler, ReadHeaderTimeout: 5 * time.Second, IdleTimeout: 30 * time.Second, MaxHeaderBytes: 16384}
	log.Printf("Cease & Quiet local preview: http://%s", *addr)
	log.Fatal(server.ListenAndServe())
}
