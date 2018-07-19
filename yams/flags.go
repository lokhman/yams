package yams

import (
	"flag"
	"os"

	"github.com/gin-gonic/gin"
)

var (
	Mode        = *flag.String("mode", GetEnv("YAMS_MODE", gin.ReleaseMode), "Server mode")
	ProxyAddr   = *flag.String("proxy-addr", GetEnv("YAMS_PROXY_ADDR", ":8086"), "Proxy server address")
	ConsoleAddr = *flag.String("console-addr", GetEnv("YAMS_CONSOLE_ADDR", ":8087"), "Console server address")
	DSN         = *flag.String("dsn", GetEnv("DATABASE_URL", "postgres://localhost"), "Database connection URL")
)

func init() {
	gin.SetMode(Mode)
}

func GetEnv(key, value string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return value
}
