package proxy

import (
	"database/sql"
	"flag"
	"fmt"
	"net/http"

	"github.com/lokhman/yams/utils"
)

const (
	headerStatus    = "x-yams-status"
	headerRouteID   = "x-yams-route-id"
	headerSessionID = "x-yams-session-id"

	statusError       = "error"
	statusProxy       = "proxy"
	statusIntercepted = "intercepted"

	sessionIDSize = 24
)

var (
	Server = &server{Server: http.Server{
		Addr:    *flag.String("proxy-addr", utils.GetEnv("YAMS_PROXY_ADDR", ":8086"), "Proxy server address"),
		Handler: &handler{},
	}}
	DB *sql.DB
)

type server struct {
	http.Server
}

type handler struct{}

func (s *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var yr *route

	defer (func() {
		if err := recover(); err != nil {
			proxyError{http.StatusInternalServerError, fmt.Sprintf("%v", err), yr}.write(w)
		}
	})()

	yp := matchProfile(r.Host)
	if yp == nil {
		proxyError{http.StatusNotFound, fmt.Sprintf(`No profile configured for host "%s"`, r.Host), nil}.write(w)
		return
	}

	yr = matchRoute(yp, r.Method, r.URL.Path)
	if yr == nil {
		if yp.backend == nil {
			proxyError{http.StatusNotFound, fmt.Sprintf(`No route found for path "%s"`, r.URL.Path), nil}.write(w)
			return
		}

		yp.proxy(w, r)
		return
	}

	yr.execute(w, r)
}
