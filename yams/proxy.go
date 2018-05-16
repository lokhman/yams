package yams

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

const (
	ProxyHeaderStatus    = "X-YAMS-Status"
	ProxyHeaderRouteId   = "X-YAMS-Route-Id"
	ProxyHeaderSessionId = "X-YAMS-Session-Id"

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
