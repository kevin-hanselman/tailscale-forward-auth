package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"encoding/json"
)

var (
	listenAddr = flag.String("addr", "127.0.0.1:", "the TCP address to listen on")
)

func main() {
	flag.Parse()
	if *listenAddr == "" {
		log.Fatal("listen address not set")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		e := json.NewEncoder(w)
		e.SetIndent("", "  ")
		e.Encode(r.Header)
	})

	ln, err := net.Listen("tcp", *listenAddr)
	if err != nil {
		log.Fatalf("can't listen on %s: %v", *listenAddr, err)
	}
	defer ln.Close()

	log.Printf("listening on %s", ln.Addr())
	log.Fatal(http.Serve(ln, mux))
}
