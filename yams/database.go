package yams

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"github.com/lokhman/sqrl"
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

func MapRows(rows *sql.Rows) ([]map[string]interface{}, error) {
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	r := make([]map[string]interface{}, 0)
	for rows.Next() {
		keys := make([]interface{}, len(cols))
		values := make([]interface{}, len(cols))
		for i, _ := range keys {
			values[i] = &keys[i]
		}
		if err = rows.Scan(values...); err != nil {
			return nil, err
		}
		p := make(map[string]interface{})
		for i, n := range cols {
			p[n] = *values[i].(*interface{})
		}
		r = append(r, p)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return r, nil
}
