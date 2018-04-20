package console

import (
	"time"

	"github.com/gin-gonic/gin"
)

type hAPI struct{ *gin.RouterGroup }

func (h *hAPI) IndexAction(c *gin.Context) {
	c.JSON(200, map[string]interface{}{
		"uptime": upTime.Format(time.RFC3339),
	})
}

type hAPIPrivate struct{ *gin.RouterGroup }

func (h *hAPIPrivate) AuthAction(c *gin.Context) {
	//
}
