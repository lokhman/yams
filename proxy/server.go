package proxy

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/lokhman/yams/proxy/adapter"
	"github.com/lokhman/yams/proxy/model"
	"github.com/lokhman/yams/yams"
)

const sidLength = 24

var Server = &http.Server{
	Addr:    yams.ProxyAddr,
	Handler: &handler{},
}

type handler struct{}

func (s *handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	var r *model.Route

	defer (func() {
		if err := recover(); err != nil {
			perror(rw, http.StatusInternalServerError, fmt.Sprintf("%v", err), r, skipPanic)
		}
	})()

	p := model.MatchProfile(req.Host)
	if p == nil {
		perror(rw, http.StatusNotFound, fmt.Sprintf(`yams: no profile configured for host "%s"`, req.Host), nil, skipError)
		return
	}

	r = model.MatchRoute(p, req.Method, req.URL.Path)
	if r == nil || !r.IsEnabled {
		if p.Backend == nil {
			perror(rw, http.StatusNotFound, fmt.Sprintf(`yams: no route found for path "%s"`, req.URL.Path), nil, skipError)
			return
		}
		yams.ReverseProxy(rw, req, *p.Backend, p.IsDebug)
		return
	}

	sid := strings.TrimSpace(req.Header.Get(yams.ProxyHeaderSessionId))
	if len(sid) > sidLength {
		sid = sid[:sidLength]
	}
	if sid == "" {
		sid = yams.RandString(sidLength)
	}

	if r.Profile.IsDebug {
		rw.Header().Set(yams.ProxyHeaderStatus, yams.ProxyStatusIntercepted)
		rw.Header().Set(yams.ProxyHeaderRouteId, r.UUID)
		rw.Header().Set(yams.ProxyHeaderSessionId, sid)
	}

	if r.Timeout == 0 {
		w, ok := rw.(http.Hijacker)
		if !ok {
			panic("yams: unable to hijack response writer")
		}
		conn, buf, err := w.Hijack()
		if err != nil {
			panic(err)
		}
		buf.Flush()
		conn.Close()
		return
	}

	var err error
	switch r.Adapter {
	case yams.AdapterLua:
		err = adapter.NewLuaScript(r, rw, req, sid).Execute()
	default:
		panic(fmt.Sprintf(`yams: unknown adapter "%s"`, r.Adapter))
	}
	if err != nil {
		panic(err)
	}
}
