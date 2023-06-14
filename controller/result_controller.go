package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func resultController() {
	router.POST("/result/import", importResult)
}

func importResult(c *gin.Context) {
	c.Status(http.StatusNotImplemented)
}
