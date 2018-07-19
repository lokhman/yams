package console

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type hWeb struct{ *gin.RouterGroup }

func (h *hWeb) IndexAction(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}
