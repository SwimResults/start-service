package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/swimresults/start-service/dto"
	"github.com/swimresults/start-service/service"
	"net/http"
)

func resultController() {
	router.POST("/result/import", importResult)
}

func importResult(c *gin.Context) {
	var request dto.ImportResultRequestDto
	if err := c.BindJSON(&request); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	result, _, err := service.ImportResult(request.Start, request.Result)
	if err != nil {
		fmt.Println(err)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusCreated, result)

}
