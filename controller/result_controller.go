package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/swimresults/service-core/security"
	"github.com/swimresults/start-service/dto"
	"github.com/swimresults/start-service/service"
)

func resultController() {
	security.Route(router, "POST", "/result/import", security.PermissionMeeting, importResult)
}

func importResult(c *gin.Context) {
	var request dto.ImportResultRequestDto
	if err := c.BindJSON(&request); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	result, r, err := service.ImportResult(request.Start, request.Result)
	if err != nil {
		fmt.Println(err)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	if r {
		c.IndentedJSON(http.StatusCreated, result)
	} else {
		c.IndentedJSON(http.StatusOK, result)
	}

}
