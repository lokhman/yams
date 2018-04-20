package proxy

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/lokhman/yams/yams"
)

const (
	headerStatus    = "x-yams-status"
	headerRouteID   = "x-yams-route-id"
	headerSessionID = "x-yams-session-id"

	statusError       = "error"
	statusProxy       = "proxy"
	statusIntercepted = "intercepted"

	maxMemory = 64 << 20
	sidSize   = 24
)

var Server = &http.Server{
	Addr:    yams.ProxyAddr,
	Handler: &handler{},
}

var db = yams.DB

type handler struct{}

func (s *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var yr *route

	defer (func() {
		if err := recover(); err != nil {
			perr(w, http.StatusInternalServerError, fmt.Sprintf("%v", err), yr)
		}
	})()

	yp := matchProfile(r.Host)
	if yp == nil {
		perr(w, http.StatusNotFound, fmt.Sprintf(`No profile configured for host "%s"`, r.Host), nil)
		return
	}

	yr = matchRoute(yp, r.Method, r.URL.Path)
	if yr == nil {
		if yp.backend == nil {
			perr(w, http.StatusNotFound, fmt.Sprintf(`No route found for path "%s"`, r.URL.Path), nil)
			return
		}

		yp.proxy(w, r)
		return
	}

	yr.execute(w, r)
}

func ip(r *http.Request) string {
	ip := r.Header.Get("x-forwarded-for")
	if index := strings.IndexByte(ip, ','); index >= 0 {
		ip = ip[0:index]
	}
	ip = strings.TrimSpace(ip)
	if ip != "" {
		return ip
	}
	ip = strings.TrimSpace(r.Header.Get("x-real-ip"))
	if ip != "" {
		return ip
	}
	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		return ip
	}
	return ""
}
