package api

import (
	"net"
	"net/http"
	"strings"
)

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ip := xff
		if i := strings.Index(xff, ","); i != -1 {
			ip = xff[:i]
		}
		ip = strings.TrimSpace(ip)
		if ip != "" {
			return ip
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
