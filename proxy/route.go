package proxy

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/lib/pq"
	"github.com/lokhman/yams/yams"
)

type route struct {
	profile *profile
	uuid    string
	method  string
	path    string
	script  string
	timeout int
	args    map[string]string
}

func (r route) Debug() [][2]string {
	return [][2]string{
		{"ID", r.uuid},
		{"Request", r.method + " " + r.path},
		{"Timeout", strconv.Itoa(r.timeout)},
	}
}

func (r *route) execute(rw http.ResponseWriter, req *http.Request) {
	sid := strings.TrimSpace(req.Header.Get(headerSessionID))
	if len(sid) > sidSize {
		sid = sid[:sidSize]
	}
	if sid == "" {
		sid = yams.RandString(sidSize)
	}

	if r.profile.debug {
		rw.Header().Set(headerStatus, statusIntercepted)
		rw.Header().Set(headerRouteID, r.uuid)
		rw.Header().Set(headerSessionID, sid)
	}

	if err := newScript(r, rw, req, sid).execute(); err != nil {
		panic(err)
	}
}

func matchRoute(p *profile, method, path string) *route {
	var pk, pv pq.StringArray
	r := &route{profile: p, method: method}

	q := `SELECT uuid, path, path_args, regexp_matches($3, path_re), script, timeout FROM routes WHERE profile_id = $1 AND $2 = ANY(methods) AND $3 ~ path_re ORDER BY position LIMIT 1`
	if err := db.QueryRow(q, p.id, method, path).Scan(&r.uuid, &r.path, &pk, &pv, &r.script, &r.timeout); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		panic(err)
	}

	r.args = make(map[string]string)
	for i, key := range pk {
		r.args[key] = pv[i]
	}

	return r
}
