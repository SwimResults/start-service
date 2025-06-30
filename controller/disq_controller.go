package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/swimresults/start-service/dto"
	"github.com/swimresults/start-service/service"
	"net/http"
)

func disqualificationController() {
	router.POST("/disqualification/import", importDisqualification)
}

func importDisqualification(c *gin.Context) {
	var request dto.ImportDisqualificationRequestDto
	if err := c.BindJSON(&request); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	disqualification, _, err := service.ImportDisqualification(request.Start, request.Disqualification)
	if err != nil {
		fmt.Println(err)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, disqualification)

}
