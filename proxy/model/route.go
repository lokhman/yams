package model

import (
	"database/sql"
	"strconv"

	"github.com/lib/pq"
	"github.com/lokhman/yams/yams"
)

type Route struct {
	Profile *Profile
	UUID    string
	Method  string
	Path    string
	Adapter string
	Script  string
	Timeout int
	Args    map[string]string
}

func (r Route) Debug() [][2]string {
	return [][2]string{
		{"ID", r.UUID},
		{"Request", r.Method + " " + r.Path},
		{"Timeout", strconv.Itoa(r.Timeout)},
	}
}

func MatchRoute(p *Profile, method, path string) *Route {
	var pk, pv pq.StringArray
	r := &Route{Profile: p, Method: method}

	q := `SELECT uuid, path, path_args, regexp_matches($3, path_re), adapter, script, timeout FROM routes WHERE profile_id = $1 AND methods && $2 AND $3 ~ path_re ORDER BY position LIMIT 1`
	if err := yams.DB.QueryRow(q, p.Id, pq.StringArray{method, "*"}, path).Scan(&r.UUID, &r.Path, &pk, &pv, &r.Adapter, &r.Script, &r.Timeout); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		panic(err)
	}

	r.Args = make(map[string]string)
	for i, key := range pk {
		r.Args[key] = pv[i]
	}
	return r
}
