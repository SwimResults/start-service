package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/swimresults/service-core/security"
	"github.com/swimresults/start-service/model"
	"github.com/swimresults/start-service/service"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func registrationController() {
	security.Route(router, http.MethodGet, "/registration/:id", security.PermissionMeeting, getRegistration)

	security.Route(router, http.MethodGet, "/registration/meet/:meet_id", security.PermissionMeeting, getRegistrationsByMeeting)
	security.Route(router, http.MethodGet, "/registration/meet/:meet_id/me", security.PermissionMeeting, getRegistrationsByMeetingForMe)

	security.Route(router, http.MethodPost, "/registration", security.PermissionAdmin, addRegistration)

	security.Route(router, http.MethodPut, "/registration", security.PermissionAdmin, updateRegistration)
	security.Route(router, http.MethodDelete, "/registration/:id", security.PermissionAdmin, removeRegistration)
}

func getRegistration(c *gin.Context) {
	id, convErr := primitive.ObjectIDFromHex(c.Param("id"))
	if convErr != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "given id was not of type ObjectID"})
		return
	}

	registration, err := service.GetRegistrationById(id)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, registration)
}

func getRegistrationsByMeeting(c *gin.Context) {
	meeting := c.Param("meet_id")

	if meeting == "" {
		c.String(http.StatusBadRequest, "no meeting id given")
		return
	}

	registrations, err := service.GetRegistrationsByMeeting(meeting)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, registrations)
}

func getRegistrationsByMeetingForMe(c *gin.Context) {
	meeting := c.Param("meet_id")

	if meeting == "" {
		c.String(http.StatusBadRequest, "no meeting id given")
		return
	}

	id, convErr := primitive.ObjectIDFromHex(c.Query("user_id"))
	if convErr != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "given id was not of type ObjectID"})
		return
	}

	registrations, err := service.GetRegistrationByMeetingAndUser(meeting, id)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, registrations)
}

func removeRegistration(c *gin.Context) {
	id, convErr := primitive.ObjectIDFromHex(c.Param("id"))
	if convErr != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "given id was not of type ObjectID"})
		return
	}

	err := service.RemoveRegistrationById(id)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusNoContent, "")
}

func addRegistration(c *gin.Context) {
	var registration model.Registration
	if err := c.BindJSON(&registration); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	r, err := service.AddRegistration(registration)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, r)
}

func updateRegistration(c *gin.Context) {
	var registration model.Registration
	if err := c.BindJSON(&registration); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	r, err := service.UpdateRegistration(registration)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, r)
}
