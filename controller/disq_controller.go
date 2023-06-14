package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func disqualificationController() {
	router.POST("/disqualification/import", importDisqualification)
}

func importDisqualification(c *gin.Context) {
	c.Status(http.StatusNotImplemented)
}
