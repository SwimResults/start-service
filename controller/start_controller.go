package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"sr-start/start-service/model"
	"sr-start/start-service/service"
	"strconv"
)

func startController() {
	router.GET("/start", getStarts)
	router.GET("/start/:id", getStart)

	router.GET("/start/meet/:meet_id", getStartsByMeeting)
	router.GET("/start/meet/:meet_id/event/:event_id/heat/:heat_id", getStartsByMeetingAndEventAndHeat)
	router.GET("/start/meet/:meet_id/event/:event_id/heat/:heat_id/lane/:lane_number", getStartsByMeetingAndEventAndHeatAndLane)
	router.GET("/start/meet/:meet_id/event/:event_id", getStartsByMeetingAndEvent)
	router.GET("/start/meet/:meet_id/athlete/:ath_id", getStartsByMeetingAndAthlete)
	router.GET("/start/athlete/:ath_id", getStartsByAthlete)

	router.DELETE("/start/:id", removeStart)
	router.POST("/start", addStart)
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
	fmt.Printf("getting with: %s %s %d", meeting, event, heat)
	start, err := service.GetStartsByMeetingAndEventAndHeat(meeting, event, heat)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, start)
}

func getStartsByMeetingAndEventAndHeatAndLane(c *gin.Context) {
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
	fmt.Printf("getting with: %s %s %d", meeting, event, heat)
	start, err := service.GetStartsByMeetingAndEventAndHeatAndLane(meeting, event, heat, lane)
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
