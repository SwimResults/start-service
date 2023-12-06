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

func heatController() {
	router.GET("/heat", getHeats)
	router.GET("/heat/:id", getHeat)

	router.GET("/heat/amount", getHeatsAmount)
	router.GET("/heat/meet/:meet_id/amount", getHeatsAmountByMeeting)

	router.GET("/heat/meet/:meet_id", getHeatsByMeeting)
	router.GET("/heat/meet/:meet_id/event_list", getHeatsByMeetingForEventList)
	router.GET("/heat/meet/:meet_id/info", getHeatInfoByMeeting)
	router.GET("/heat/meet/:meet_id/event/:event_id/info", getHeatInfoByMeetingAndEvent)
	router.GET("/heat/meet/:meet_id/current", getCurrentHeat)

	router.POST("/heat", addHeat)
	router.POST("/heat/import", importHeat)
	router.POST("/heat/:id/time", updateHeatTime)

	router.PUT("/heat", updateHeat)
	router.DELETE("/heat/:id", removeHeat)
}

func getHeats(c *gin.Context) {
	heats, err := service.GetHeats()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, heats)
}

func getHeat(c *gin.Context) {
	id, convErr := primitive.ObjectIDFromHex(c.Param("id"))
	if convErr != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "given id was not of type ObjectID"})
		return
	}

	heat, err := service.GetHeatById(id)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, heat)
}

func getHeatsAmount(c *gin.Context) {
	starts, err := service.GetHeatsAmount()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, starts)
}

func getHeatsAmountByMeeting(c *gin.Context) {
	meeting := c.Param("meet_id")

	if meeting == "" {
		c.String(http.StatusBadRequest, "no meeting id given")
		return
	}

	starts, err := service.GetHeatsAmountByMeeting(meeting)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, starts)
}

func getHeatsByMeeting(c *gin.Context) {
	meeting := c.Param("meet_id")

	if meeting == "" {
		c.String(http.StatusBadRequest, "no meeting id given")
		return
	}

	heats, err := service.GetHeatsByMeeting(meeting)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, heats)
}

func getHeatsByMeetingForEventList(c *gin.Context) {
	meeting := c.Param("meet_id")

	if meeting == "" {
		c.String(http.StatusBadRequest, "no meeting id given")
		return
	}
	events := c.QueryArray("events")

	fmt.Println(events)

	var info dto.MeetingHeatsEventListDto
	var err error

	if len(events) == 0 {
		info, err = service.GetHeatsByMeetingForEventList(meeting)
	} else {
		var eventNumbers []int
		for _, event := range events {
			i, err2 := strconv.Atoi(event)
			if err2 != nil {
				c.IndentedJSON(http.StatusNotFound, gin.H{"message": err2.Error()})
				return
			}
			eventNumbers = append(eventNumbers, i)
		}
		info, err = service.GetHeatsByMeetingForEventListEvents(meeting, eventNumbers)
	}

	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, info)
}

func getHeatInfoByMeeting(c *gin.Context) {
	meeting := c.Param("meet_id")

	if meeting == "" {
		c.String(http.StatusBadRequest, "no meeting id given")
		return
	}

	info, err := service.GetHeatInfoByMeeting(meeting)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, info)
}

func getHeatInfoByMeetingAndEvent(c *gin.Context) {
	meeting := c.Param("meet_id")

	if meeting == "" {
		c.String(http.StatusBadRequest, "no meeting id given")
		return
	}

	event, err1 := strconv.Atoi(c.Param("event_id"))
	if err1 != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "given event_id is not of type number"})
		return
	}

	info, err := service.GetHeatInfoByMeetingAndEvent(meeting, event)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, info)
}

func getCurrentHeat(c *gin.Context) {
	meeting := c.Param("meet_id")

	if meeting == "" {
		c.String(http.StatusBadRequest, "no meeting id given")
		return
	}

	heat, err := service.GetCurrentHeat(meeting)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, heat)
}

func removeHeat(c *gin.Context) {
	id, convErr := primitive.ObjectIDFromHex(c.Param("id"))
	if convErr != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "given id was not of type ObjectID"})
		return
	}

	err := service.RemoveHeatById(id)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusNoContent, "")
}

func addHeat(c *gin.Context) {
	var heat model.Heat
	if err := c.BindJSON(&heat); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	r, err := service.AddHeat(heat)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, r)
}

func importHeat(c *gin.Context) {
	var request dto.ImportHeatRequestDto
	if err := c.BindJSON(&request); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	heat, r, err := service.ImportHeat(request.Heat)
	if err != nil {
		fmt.Printf(err.Error())
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	if r {
		c.IndentedJSON(http.StatusCreated, heat)
	} else {
		c.IndentedJSON(http.StatusOK, heat)
	}

}

func updateHeat(c *gin.Context) {
	var heat model.Heat
	if err := c.BindJSON(&heat); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	r, err := service.UpdateHeat(heat)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, r)
}

func updateHeatTime(c *gin.Context) {
	id, convErr := primitive.ObjectIDFromHex(c.Param("id"))
	if convErr != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "given id was not of type ObjectID"})
		return
	}

	var request dto.HeatTimesRequestDto
	if err := c.BindJSON(&request); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	r, err := service.UpdateHeatTimes(id, request.Time, request.Type)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, r)
}
