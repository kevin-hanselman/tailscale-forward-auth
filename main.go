// This code is forked from Tailscale codebase which is governed by
// a BSD-style licence. See https://github.com/tailscale/tailscale
//
// The link below is the code from which this code originates:
// https://github.com/tailscale/tailscale/blob/741ae9956e674177687062b5499a80db83505076/cmd/nginx-auth/nginx-auth.go

package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"

	"tailscale.com/client/tailscale"
)

var (
	listenProto      = flag.String("network", "tcp", "type of network to listen on, defaults to tcp")
	listenAddr       = flag.String("addr", "127.0.0.1:", "address to listen on, defaults to 127.0.0.1:")
	headerRemoteIP   = flag.String("remote-ip-header", "X-Forwarded-For", "HTTP header field containing the remote IP")
	headerRemotePort = flag.String("remote-port-header", "X-Forwarded-Port", "HTTP header field containing the remote port")
	debug            = flag.Bool("debug", false, "enable debug logging")
)

func main() {
	flag.Parse()
	if *listenAddr == "" {
		log.Fatal("listen address not set")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if *debug {
			log.Printf("received request with header %+v", r.Header)
		}

		remoteHost := r.Header.Get(*headerRemoteIP)
		if remoteHost == "" {
			w.WriteHeader(http.StatusBadRequest)
			log.Printf("missing header %s", *headerRemoteIP)
			return
		}

		remotePort := r.Header.Get(*headerRemotePort)
		if remotePort == "" {
			w.WriteHeader(http.StatusBadRequest)
			log.Printf("missing header %s", *headerRemotePort)
			return
		}

		remoteAddr, err := netip.ParseAddrPort(net.JoinHostPort(remoteHost, remotePort))
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			log.Printf("remote address and port are not valid: %v", err)
			return
		}

		client := &tailscale.LocalClient{}
		info, err := client.WhoIs(r.Context(), remoteAddr.String())
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			log.Printf("can't look up %s: %v", remoteAddr, err)
			return
		}

		if len(info.Node.Tags) != 0 {
			w.WriteHeader(http.StatusForbidden)
			log.Printf("node %s is tagged", info.Node.Hostinfo.Hostname())
			return
		}

		// tailnet of connected node. When accessing shared nodes, this
		// will be empty because the tailnet of the sharee is not exposed.
		var tailnet string

		if !info.Node.Hostinfo.ShareeNode() {
			var ok bool
			_, tailnet, ok = strings.Cut(info.Node.Name, info.Node.ComputedName+".")
			if !ok {
				w.WriteHeader(http.StatusUnauthorized)
				log.Printf("can't extract tailnet name from hostname %q", info.Node.Name)
				return
			}
			tailnet = strings.TrimSuffix(tailnet, ".beta.tailscale.net")
		}

		if expectedTailnet := r.Header.Get("Expected-Tailnet"); expectedTailnet != "" && expectedTailnet != tailnet {
			w.WriteHeader(http.StatusForbidden)
			log.Printf("user is part of tailnet %s, wanted: %s", tailnet, url.QueryEscape(expectedTailnet))
			return
		}

		h := w.Header()
		h.Set("Tailscale-Login", strings.Split(info.UserProfile.LoginName, "@")[0])
		h.Set("Tailscale-User", info.UserProfile.LoginName)
		h.Set("Tailscale-Name", info.UserProfile.DisplayName)
		h.Set("Tailscale-Profile-Picture", info.UserProfile.ProfilePicURL)
		h.Set("Tailscale-Tailnet", tailnet)
		w.WriteHeader(http.StatusNoContent)
	})

	ln, err := net.Listen(*listenProto, *listenAddr)
	if err != nil {
		log.Fatalf("can't listen on %s: %v", *listenAddr, err)
	}
	defer ln.Close()

	log.Printf("listening on %s", ln.Addr())
	log.Fatal(http.Serve(ln, mux))
}
