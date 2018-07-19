package model

import (
	"database/sql"
	"strings"

	"github.com/lokhman/yams/yams"
)

type Profile struct {
	Id           int
	Host         string
	Backend      *string
	IsDebug      bool
	VarsLifetime int
}

func MatchProfile(host string) *Profile {
	// allow HTTP and HTTPS by default
	host = strings.TrimSuffix(host, ":80")
	host = strings.TrimSuffix(host, ":443")

	p := &Profile{Host: host}
	q := `SELECT id, backend, is_debug, vars_lifetime FROM profiles WHERE $1 = ANY(hosts)`
	if err := yams.DB.QueryRow(q, host).Scan(&p.Id, &p.Backend, &p.IsDebug, &p.VarsLifetime); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		panic(err)
	}
	return p
}
