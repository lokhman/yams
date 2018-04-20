package yams

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func init() {
	var err error
	if DB, err = sql.Open("postgres", DSN); err != nil {
		log.Fatal(err)
	}
}
