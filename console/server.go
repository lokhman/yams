package console

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lokhman/yams/yams"
)

var Server = &http.Server{
	Addr:    yams.ConsoleAddr,
	Handler: handler(),
}

var db = yams.DB
var upTime = time.Now()

func handler() http.Handler {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/", func(c *gin.Context) {
		c.String(200, "test")
	})

	api := &hAPI{r.Group("/api")}
	api.GET("", api.IndexAction)

	return r
}
