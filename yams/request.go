package yams

import (
	"net"
	"net/http"
	"strings"
)

func ClientIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if index := strings.IndexByte(ip, ','); index >= 0 {
		ip = ip[0:index]
	}
	ip = strings.TrimSpace(ip)
	if ip != "" {
		return ip
	}
	ip = strings.TrimSpace(r.Header.Get("X-Real-IP"))
	if ip != "" {
		return ip
	}
	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		return ip
	}
	return ""
}
