package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/swimresults/start-service/dto"
	"github.com/swimresults/start-service/model"
	"github.com/swimresults/start-service/service"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"strconv"
)

func startController() {
	router.GET("/start", getStarts)
	router.GET("/start/:id", getStart)

	router.GET("/start/amount", getStartsAmount)

	router.GET("/start/meet/:meet_id", getStartsByMeeting)
	router.GET("/start/meet/:meet_id/event/:event_id/heat/:heat_id", getStartsByMeetingAndEventAndHeat)
	router.GET("/start/meet/:meet_id/event/:event_id/heat/:heat_id/lane/:lane_number", getStartByMeetingAndEventAndHeatAndLane)
	router.GET("/start/meet/:meet_id/event/:event_id", getStartsByMeetingAndEvent)
	router.GET("/start/meet/:meet_id/athlete/:ath_id", getStartsByMeetingAndAthlete)
	router.GET("/start/meet/:meet_id/current", getCurrentStarts)
	router.GET("/start/meet/:meet_id/livestream", getLivestreamData)
	router.GET("/start/athlete/:ath_id", getStartsByAthlete)

	router.POST("/start", addStart)
	router.POST("/start/import", importStart)

	router.DELETE("/start/:id", removeStart)
	router.PUT("/start", updateStart)
}

func getStarts(c *gin.Context) {
	starts, err := service.GetStarts()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, starts)
}

func getStartsAmount(c *gin.Context) {
	starts, err := service.GetStartsAmount()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, starts)
}

func getStart(c *gin.Context) {
	id, convErr := primitive.ObjectIDFromHex(c.Param("id"))
	if convErr != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "given id was not of type ObjectID"})
		return
	}

	start, err := service.GetStartById(id)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, start)
}

func getStartsByMeeting(c *gin.Context) {
	meeting := c.Param("meet_id")

	if meeting == "" {
		c.String(http.StatusBadRequest, "no meeting id given")
		return
	}

	start, err := service.GetStartsByMeeting(meeting)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, start)
}

func getStartsByMeetingAndEventAndHeat(c *gin.Context) {
	meeting := c.Param("meet_id")

	if meeting == "" {
		c.String(http.StatusBadRequest, "no meeting id given")
		return
	}

	event, err := strconv.Atoi(c.Param("event_id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "given event_id is not of type number"})
		return
	}

	heat, convErr := strconv.Atoi(c.Param("heat_id"))
	if convErr != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "given heat_id was not an int"})
		return
	}
	start, err := service.GetStartsByMeetingAndEventAndHeat(meeting, event, heat)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, start)
}

func getStartByMeetingAndEventAndHeatAndLane(c *gin.Context) {
	meeting := c.Param("meet_id")

	if meeting == "" {
		c.String(http.StatusBadRequest, "no meeting id given")
		return
	}

	event, err := strconv.Atoi(c.Param("event_id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "given event_id is not of type number"})
		return
	}

	heat, convErr := strconv.Atoi(c.Param("heat_id"))
	if convErr != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "given heat_id was not an int"})
		return
	}

	lane, convErr2 := strconv.Atoi(c.Param("lane_number"))
	if convErr2 != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "given lane_number was not an int"})
		return
	}
	start, err := service.GetStartByMeetingAndEventAndHeatAndLane(meeting, event, heat, lane)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, start)
}

func getStartsByMeetingAndEvent(c *gin.Context) {
	meeting := c.Param("meet_id")

	if meeting == "" {
		c.String(http.StatusBadRequest, "no meeting id given")
		return
	}

	event, err := strconv.Atoi(c.Param("event_id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "given event_id is not of type number"})
		return
	}

	start, err := service.GetStartsByMeetingAndEvent(meeting, event)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, start)
}

func getStartsByMeetingAndAthlete(c *gin.Context) {
	meeting := c.Param("meet_id")

	if meeting == "" {
		c.String(http.StatusBadRequest, "no meeting id given")
		return
	}

	athlete, convErr := primitive.ObjectIDFromHex(c.Param("ath_id"))
	if convErr != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "given ath_id was not of type ObjectID"})
		return
	}

	start, err := service.GetStartsByMeetingAndAthlete(meeting, athlete)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, start)
}

func getStartsByAthlete(c *gin.Context) {
	athlete, convErr := primitive.ObjectIDFromHex(c.Param("ath_id"))
	if convErr != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "given ath_id was not of type ObjectID"})
		return
	}

	start, err := service.GetStartsByAthlete(athlete)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, start)
}

func getCurrentStarts(c *gin.Context) {
	meeting := c.Param("meet_id")

	if meeting == "" {
		c.String(http.StatusBadRequest, "no meeting id given")
		return
	}

	start, err := service.GetCurrentStarts(meeting)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, start)
}

func getLivestreamData(c *gin.Context) {
	meeting := c.Param("meet_id")

	if meeting == "" {
		c.String(http.StatusBadRequest, "no meeting id given")
		return
	}

	data, err := service.GetLivestreamData(meeting)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, data)
}

func removeStart(c *gin.Context) {
	id, convErr := primitive.ObjectIDFromHex(c.Param("id"))
	if convErr != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "given id was not of type ObjectID"})
		return
	}

	err := service.RemoveStartById(id)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusNoContent, "")
}

func addStart(c *gin.Context) {
	var start model.Start
	if err := c.BindJSON(&start); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	r, err := service.AddStart(start)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, r)
}

func importStart(c *gin.Context) {
	var request dto.ImportStartRequestDto
	if err := c.BindJSON(&request); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	start, r, err := service.ImportStart(request.Start)
	if err != nil {
		fmt.Printf(err.Error())
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	if r {
		c.IndentedJSON(http.StatusCreated, start)
	} else {
		c.IndentedJSON(http.StatusOK, start)
	}

}

func updateStart(c *gin.Context) {
	var start model.Start
	if err := c.BindJSON(&start); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	r, err := service.UpdateStart(start)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, r)
}
