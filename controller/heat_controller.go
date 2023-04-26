package controller

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"sr-start/start-service/model"
	"sr-start/start-service/service"
)

func heatController() {
	router.GET("/heat", getHeats)
	router.GET("/heat/:id", getHeat)
	router.DELETE("/heat/:id", removeHeat)
	router.POST("/heat", addHeat)
	router.PUT("/heat", updateHeat)
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