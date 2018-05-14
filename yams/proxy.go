package yams

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

const (
	ProxyHeaderStatus    = "x-yams-status"
	ProxyHeaderRouteId   = "x-yams-route-id"
	ProxyHeaderSessionId = "x-yams-session-id"

	ProxyStatusError       = "error"
	ProxyStatusProxy       = "proxy"
	ProxyStatusIntercepted = "intercepted"
)

func ReverseProxy(w http.ResponseWriter, r *http.Request, backend string, debug bool) {
	u, err := url.Parse(backend)
	if err != nil {
		panic(err)
	}
	r.Host = u.Host
	if debug {
		w.Header().Set(ProxyHeaderStatus, ProxyStatusProxy)
	}
	httputil.NewSingleHostReverseProxy(u).ServeHTTP(w, r)
}
