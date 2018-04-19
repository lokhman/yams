package proxy

import (
	"database/sql"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type profile struct {
	id      int
	host    string
	backend *string
	debug   bool
}

func (p *profile) proxy(w http.ResponseWriter, r *http.Request) {
	u, err := url.Parse(*p.backend)
	if err != nil {
		panic(err)
	}
	r.Host = u.Host

	if p.debug {
		w.Header().Set(headerStatus, statusProxy)
	}

	httputil.NewSingleHostReverseProxy(u).ServeHTTP(w, r)
}

func matchProfile(host string) *profile {
	p := &profile{host: host}

	q := `SELECT id, backend, debug FROM profiles WHERE $1 = ANY(hosts)`
	if err := DB.QueryRow(q, host).Scan(&p.id, &p.backend, &p.debug); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		panic(err)
	}

	return p
}