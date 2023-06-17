package service

import (
	"context"
	client2 "github.com/swimresults/athlete-service/client"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"os"
	"time"
)

var client *mongo.Client
var athleteClient *client2.AthleteClient

func Init(c *mongo.Client) {
	database := c.Database(os.Getenv("SR_START_MONGO_DATABASE"))
	client = c

	athleteServiceUrl := os.Getenv("SR_START_ATHLETE_URL")
	if athleteServiceUrl != "" {
		athleteClient = client2.NewAthleteClient(athleteServiceUrl)
	}

	startService(database)
	heatService(database)
	disqualificationService(database)
}

func PingDatabase() bool {

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	err := client.Ping(ctx, readpref.Primary())
	if err != nil {
		return false
	}

	return true
}
