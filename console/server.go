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

var db, qb = yams.DB, yams.QB
var upTime = time.Now()

func handler() http.Handler {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.LoadHTMLFiles("public/console/dist/index.html")

	// Index
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/web/")
	})

	// Web
	web := &hWeb{r.Group("/web")}
	web.GET("/*?", web.IndexAction)

	// Static
	r.StaticFS("/static", gin.Dir("static", false))
	r.StaticFile("/favicon.ico", "./static/favicon.ico")

	// API public
	api := &hAPI{r.Group("/api")}
	api.GET("", api.IndexAction)
	api.POST("/auth", api.AuthAction)

	// API private
	api = &hAPI{r.Group("/api")}
	api.GET("/auth", api.Auth(yams.AnyRole...), api.AuthUserAction)
	api.POST("/auth/refresh", api.Auth(yams.AnyRole...), api.AuthRefreshAction)

	api.GET("/users", api.Auth(yams.RoleAdmin), api.UsersAction)
	api.POST("/users", api.Auth(yams.RoleAdmin), api.UsersCreateAction)
	api.GET("/users/:id", api.Auth(yams.RoleAdmin), api.UsersViewAction)
	api.PUT("/users/:id", api.Auth(yams.RoleAdmin), api.UsersUpdateAction)
	api.DELETE("/users/:id", api.Auth(yams.RoleAdmin), api.UsersDeleteAction)

	api.GET("/profiles", api.Auth(yams.AnyRole...), api.ProfilesAction)
	api.POST("/profiles", api.Auth(yams.RoleAdmin, yams.RoleManager), api.ProfilesCreateAction)
	api.GET("/profiles/:id", api.Auth(yams.AnyRole...), api.ProfilesViewAction)
	api.PUT("/profiles/:id", api.Auth(yams.AnyRole...), api.ProfilesUpdateAction)
	api.DELETE("/profiles/:id", api.Auth(yams.RoleAdmin, yams.RoleManager), api.ProfilesDeleteAction)

	api.GET("/profiles/:id/routes", api.Auth(yams.AnyRole...), api.RoutesAction)
	api.POST("/profiles/:id/routes", api.Auth(yams.AnyRole...), api.RoutesCreateAction)
	api.GET("/routes/:id", api.Auth(yams.AnyRole...), api.RoutesViewAction)
	api.PUT("/routes/:id", api.Auth(yams.AnyRole...), api.RoutesUpdateAction)
	api.DELETE("/routes/:id", api.Auth(yams.AnyRole...), api.RoutesDeleteAction)
	api.GET("/routes/:id/script", api.Auth(yams.AnyRole...), api.RoutesScriptViewAction)
	api.PUT("/routes/:id/script", api.Auth(yams.AnyRole...), api.RoutesScriptUpdateAction)
	api.POST("/routes/:id/position", api.Auth(yams.AnyRole...), api.RoutesPositionAction)
	api.POST("/routes/:id/state", api.Auth(yams.AnyRole...), api.RoutesStateAction)

	api.GET("/profiles/:id/assets", api.Auth(yams.AnyRole...), api.AssetsAction)
	api.PUT("/profiles/:id/assets", api.Auth(yams.AnyRole...), api.AssetsUploadAction)
	api.PUT("/profiles/:id/assets/*path", api.Auth(yams.AnyRole...), api.AssetsUploadAction)
	api.GET("/profiles/:id/assets/*path", api.Auth(yams.AnyRole...), api.AssetsDownloadAction)
	api.DELETE("/profiles/:id/assets/*path", api.Auth(yams.AnyRole...), api.AssetsDeleteAction)

	return r
}

func code(statusCode, errorCode int) int {
	return int(errorCode<<16) | (statusCode & 0xFFFF)
}

func ucode(code int) (int, int) {
	return code & 0xFFFF, code >> 16
}
