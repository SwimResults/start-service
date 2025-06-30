package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/semihalev/gin-stats"
	"github.com/swimresults/start-service/service"
	ginprometheus "github.com/zsais/go-gin-prometheus"
	"net/http"
	"os"
)

var router = gin.Default()

func Run() {

	port := os.Getenv("SR_START_PORT")

	if port == "" {
		fmt.Println("no application port given! Please set SR_START_PORT.")
		return
	}

	p := ginprometheus.NewWithConfig(ginprometheus.Config{
		Subsystem: "gin",
	})
	p.Use(router)

	router.Use(stats.RequestStats())

	startController()
	heatController()
	disqualificationController()
	resultController()
	registrationController()

	router.GET("/actuator", actuator)
	router.GET("/stats", func(c *gin.Context) {
		c.JSON(http.StatusOK, stats.Report())
	})

	err := router.Run(":" + port)
	if err != nil {
		fmt.Println("Unable to start application on port " + port)
		return
	}
}

func actuator(c *gin.Context) {

	state := "OPERATIONAL"

	if !service.PingDatabase() {
		state = "DATABASE_DISCONNECTED"
	}
	c.String(http.StatusOK, state)
}
