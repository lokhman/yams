package console

import (
	"database/sql"
	"flag"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lokhman/yams/utils"
)

var (
	Server = &server{Server: http.Server{
		Addr:    *flag.String("console-addr", utils.GetEnv("YAMS_CONSOLE_ADDR", ":8087"), "Console server address"),
		Handler: handler(),
	}}
	DB *sql.DB
)

type server struct {
	http.Server
}

func handler() http.Handler {
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/", func(c *gin.Context) {
		c.String(200, "test")
	})

	return r
}
