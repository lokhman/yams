package model

import (
	"database/sql"

	"github.com/lokhman/yams/yams"
)

type Profile struct {
	Id           int
	Host         string
	Backend      *string
	Debug        bool
	VarsLifetime int
}

func MatchProfile(host string) *Profile {
	p := &Profile{Host: host}
	q := `SELECT id, backend, debug, vars_lifetime FROM profiles WHERE $1 = ANY(hosts)`
	if err := yams.DB.QueryRow(q, host).Scan(&p.Id, &p.Backend, &p.Debug, &p.VarsLifetime); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		panic(err)
	}
	return p
}
