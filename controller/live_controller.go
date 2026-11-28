package controller

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/swimresults/service-core/security"
	"github.com/swimresults/start-service/dto"
	"github.com/swimresults/start-service/service"
)

func liveController() {
	router.GET("/live/meet/:meet_id", getLiveStatus)

	security.Route(router, http.MethodPost, "/live/meet/:meet_id/data", security.PermissionMeeting, recordLiveData)
	security.Route(router, http.MethodPut, "/live/meet/:meet_id/message", security.PermissionMeeting, setLiveMessage)
	security.Route(router, http.MethodDelete, "/live/meet/:meet_id/message", security.PermissionMeeting, clearLiveMessage)
}

func getLiveStatus(c *gin.Context) {
	meeting := c.Param("meet_id")
	if meeting == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "no meeting id given"})
		return
	}

	status, err := service.GetLiveStatus(meeting)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, status)
}

func recordLiveData(c *gin.Context) {
	meeting := c.Param("meet_id")
	if meeting == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "no meeting id given"})
		return
	}

	var request dto.LiveDataRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	source := strings.TrimSpace(request.Source)
	if source == "" || len(source) > 64 {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "source must be 1–64 characters"})
		return
	}

	service.RecordLiveData(meeting, source)
	c.Status(http.StatusNoContent)
}

func setLiveMessage(c *gin.Context) {
	meeting := c.Param("meet_id")
	if meeting == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "no meeting id given"})
		return
	}

	var request dto.LiveMessageRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	text, err := service.ValidateLiveMessage(request.Text, request.Mode)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if err := service.SetLiveMessage(meeting, text, request.Mode, security.GetUserID(c)); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	status, err := service.GetLiveStatus(meeting)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, status)
}

func clearLiveMessage(c *gin.Context) {
	meeting := c.Param("meet_id")
	if meeting == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "no meeting id given"})
		return
	}

	if err := service.ClearLiveMessage(meeting); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
