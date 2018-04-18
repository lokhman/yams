package main

import (
	"database/sql"
	"flag"
	"log"

	_ "github.com/lib/pq"
	"github.com/lokhman/yams/console"
	"github.com/lokhman/yams/proxy"
	"github.com/lokhman/yams/utils"
	"golang.org/x/sync/errgroup"
)

func main() {
	flag.Parse()

	db, err := sql.Open("postgres", *flag.String("db", utils.GetEnv("DATABASE_URL", "postgres://localhost"), "Database connection URL"))
	if err != nil {
		log.Fatal(err)
	}

	proxy.DB = db
	console.DB = db

	var stack errgroup.Group
	stack.Go(proxy.Server.ListenAndServe)
	stack.Go(console.Server.ListenAndServe)

	if err = stack.Wait(); err != nil {
		log.Fatal(err)
	}
}
