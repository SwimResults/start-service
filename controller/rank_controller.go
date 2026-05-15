package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/swimresults/service-core/security"
	"github.com/swimresults/start-service/dto"
	"github.com/swimresults/start-service/service"
)

func rankController() {
	security.Route(router, http.MethodPost, "/rank/import", security.PermissionMeeting, importRank)
}

func importRank(c *gin.Context) {
	var request dto.ImportRankRequestDto
	if err := c.BindJSON(&request); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	rank, r, err := service.ImportRank(request.Start, request.Rank)
	if err != nil {
		fmt.Println(err)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	if r {
		c.IndentedJSON(http.StatusCreated, rank)
	} else {
		c.IndentedJSON(http.StatusOK, rank)
	}
}
