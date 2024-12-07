package service

import (
	"context"
	client2 "github.com/swimresults/athlete-service/client"
	meetingClient "github.com/swimresults/meeting-service/client"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"os"
	"time"
)

var client *mongo.Client
var athleteClient *client2.AthleteClient
var teamClient *client2.TeamClient
var ageGroupClient *meetingClient.AgeGroupClient
var eventClient *meetingClient.EventClient

func Init(c *mongo.Client) {
	database := c.Database(os.Getenv("SR_START_MONGO_DATABASE"))
	client = c

	athleteServiceUrl := os.Getenv("SR_START_ATHLETE_URL")
	if athleteServiceUrl != "" {
		athleteClient = client2.NewAthleteClient(athleteServiceUrl)
		teamClient = client2.NewTeamClient(athleteServiceUrl)
	}

	meetingServiceUrl := os.Getenv("SR_START_MEETING_URL")
	if meetingServiceUrl != "" {
		ageGroupClient = meetingClient.NewAgeGroupClient(meetingServiceUrl)
		eventClient = meetingClient.NewEventClient(meetingServiceUrl)
	}

	startService(database)
	heatService(database)
	disqualificationService(database)
	registrationService(database)

	StartNotificationMainLoop()
}

func PingDatabase() bool {

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
	defer cancel()

	err := client.Ping(ctx, readpref.Primary())
	if err != nil {
		return false
	}

	return true
}
