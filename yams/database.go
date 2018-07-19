package yams

import (
	"database/sql"
	"log"

	"github.com/elgris/sqrl"
	_ "github.com/lib/pq"
)

var DB *sql.DB
var QB sqrl.StatementBuilderType

func init() {
	var err error
	if DB, err = sql.Open("postgres", DSN); err != nil {
		log.Fatal(err)
	}
	QB = sqrl.StatementBuilder.PlaceholderFormat(sqrl.Dollar).RunWith(DB)
}
