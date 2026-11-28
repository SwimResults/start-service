package service

import (
	"context"
	"os"
	"time"

	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var client *mongo.Client

// clients for the other services, see clients.go; nil if their URL is not configured
var athleteClient *AthleteClient
var teamClient *TeamClient
var eventClient *EventClient

func Init(c *mongo.Client) {
	database := c.Database(os.Getenv("SR_START_MONGO_DATABASE"))
	client = c

	athleteServiceUrl := os.Getenv("SR_START_ATHLETE_URL")
	if athleteServiceUrl != "" {
		if api, err := newAthleteAPI(athleteServiceUrl); err != nil {
			fmt.Printf("invalid SR_START_ATHLETE_URL %q: %v\n", athleteServiceUrl, err)
		} else {
			athleteClient = &AthleteClient{api}
			teamClient = &TeamClient{api}
		}
	}

	meetingServiceUrl := os.Getenv("SR_START_MEETING_URL")
	if meetingServiceUrl != "" {
		if c, err := newEventClient(meetingServiceUrl); err != nil {
			fmt.Printf("invalid SR_START_MEETING_URL %q: %v\n", meetingServiceUrl, err)
		} else {
			eventClient = c
		}
	}

	startService(database)
	heatService(database)
	disqualificationService(database)
	registrationService(database)
	rankingService(database)
	liveService(database)

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
