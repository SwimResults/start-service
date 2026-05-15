package controller

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/swimresults/service-core/security"
	"github.com/swimresults/start-service/dto"
	"github.com/swimresults/start-service/model"
	"github.com/swimresults/start-service/service"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func rankingController() {
	router.GET("/ranking", getRankings)
	router.GET("/ranking/:id", getRanking)
	router.GET("/ranking/meet/:meet_id", getRankingsByMeeting)
	router.GET("/ranking/meet/:meet_id/event/:event_id", getRankingByMeetingAndEvent)

	security.Route(router, http.MethodPost, "/ranking", security.PermissionMeeting, addRanking)
	security.Route(router, http.MethodPost, "/ranking/import", security.PermissionMeeting, importRanking)

	security.Route(router, http.MethodDelete, "/ranking/:id", security.PermissionMeeting, removeRanking)

	security.Route(router, http.MethodPut, "/ranking", security.PermissionMeeting, updateRanking)
}

func getRankings(c *gin.Context) {
	rankings, err := service.GetRankings()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, rankings)
}

func getRanking(c *gin.Context) {
	id, convErr := primitive.ObjectIDFromHex(c.Param("id"))
	if convErr != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "given id was not of type ObjectID"})
		return
	}

	ranking, err := service.GetRankingById(id)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, ranking)
}

func getRankingsByMeeting(c *gin.Context) {
	id := c.Param("meet_id")
	if id == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "given meet_id is empty"})
		return
	}
	rankings, err := service.GetRankingsByMeeting(id)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, rankings)
}

func getRankingByMeetingAndEvent(c *gin.Context) {
	id := c.Param("meet_id")
	if id == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "given meet_id is empty"})
		return
	}

	eventId, err := strconv.Atoi(c.Param("event_id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "event_id is not of type number"})
		return
	}

	ranking, err := service.GetRankingsByMeetingAndEvent(id, eventId)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, ranking)
}

func removeRanking(c *gin.Context) {
	id, convErr := primitive.ObjectIDFromHex(c.Param("id"))
	if convErr != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "given id was not of type ObjectID"})
		return
	}

	err := service.RemoveRankingById(id)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusNoContent, "")
}

func importRanking(c *gin.Context) {
	var request dto.ImportRankingRequestDto
	if err := c.BindJSON(&request); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	ranking, r, err := service.ImportRanking(request.Ranking)
	if err != nil {
		fmt.Println(err)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	if r {
		c.IndentedJSON(http.StatusCreated, ranking)
	} else {
		c.IndentedJSON(http.StatusOK, ranking)
	}
}

func addRanking(c *gin.Context) {
	var ranking model.Ranking
	if err := c.BindJSON(&ranking); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	r, err := service.AddRanking(ranking)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, r)
}

func updateRanking(c *gin.Context) {
	var ranking model.Ranking
	if err := c.BindJSON(&ranking); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	r, err := service.UpdateRanking(ranking)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, r)
}
