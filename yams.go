package main

import (
	"flag"
	"log"

	"github.com/lokhman/yams/console"
	"github.com/lokhman/yams/proxy"
	"golang.org/x/sync/errgroup"
)

func main() {
	flag.Parse()

	var stack errgroup.Group
	stack.Go(proxy.Server.ListenAndServe)
	stack.Go(console.Server.ListenAndServe)

	if err := stack.Wait(); err != nil {
		log.Fatal(err)
	}
}
