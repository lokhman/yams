package console

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lokhman/yams/yams"
)

const (
	aclAdmin     = "admin"
	aclManager   = "manager"
	aclDeveloper = "developer"
)

const (
	errCodeUnknown = (iota + 0xFF) & 0xFF
	_              // errCodeUndefined
	errCodeAuthNoHeader
	errCodeAuthBadToken
	errCodeAuthNoUser
	errCodeAuthNoACL
	errCodeBadCredentials
	errCodeInvalidIdentifier
	errCodeProfileHostExists
	errCodeInvalidAdapter
)

var Server = &http.Server{
	Addr:    yams.ConsoleAddr,
	Handler: handler(),
}

var _, qb = yams.DB, yams.QB
var upTime = time.Now()

func handler() http.Handler {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/", func(c *gin.Context) {
		c.String(200, "test")
	})

	// API public
	api := &hAPI{r.Group("/api")}
	api.GET("", api.IndexAction)
	api.POST("/auth", api.AuthAction)

	// API private
	api = &hAPI{r.Group("/api")}
	api.Use(api.Authentication)
	api.GET("/auth/acl", api.AuthACLAction)
	api.POST("/auth/refresh", api.AuthRefreshAction)
	api.GET("/profiles", api.ACL(aclManager, aclDeveloper), api.ProfilesAction)
	api.POST("/profiles", api.ACL(aclManager), api.ProfilesCreateAction)
	api.GET("/profiles/:id", api.ACL(aclManager, aclDeveloper), api.ProfilesViewAction)
	api.PUT("/profiles/:id", api.ACL(aclManager, aclDeveloper), api.ProfilesUpdateAction)
	api.DELETE("/profiles/:id", api.ACL(aclManager), api.ProfilesDeleteAction)
	api.GET("/profiles/:id/routes", api.ACL(aclDeveloper), api.RoutesAction)
	api.POST("/profiles/:id/routes", api.ACL(aclDeveloper), api.RoutesCreateAction)
	api.GET("/routes/:id", api.ACL(aclDeveloper), api.RoutesViewAction)
	api.PUT("/routes/:id", api.ACL(aclDeveloper), api.RoutesUpdateAction)
	api.DELETE("/routes/:id", api.ACL(aclDeveloper), api.RoutesDeleteAction)
	api.PUT("/routes/:id/script", api.ACL(aclDeveloper), api.RoutesScriptAction)

	return r
}

func code(statusCode, errorCode int) int {
	return int(errorCode<<16) | (statusCode & 0xFFFF)
}

func ucode(code int) (int, int) {
	return code & 0xFFFF, code >> 16
}
