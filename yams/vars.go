package yams

import (
	"github.com/gin-gonic/gin"
)

const MaxScriptSize = 8 << 20 // 8MB
const MaxAssetSize = 64 << 20 // 64MB

var SecretKey = RandBytes(32)
var Debug = Mode == gin.DebugMode

const (
	RoleAdmin     = "admin"
	RoleManager   = "manager"
	RoleDeveloper = "developer"
)

var AnyRole = []string{
	RoleAdmin,
	RoleManager,
	RoleDeveloper,
}
