package service

import (
	"context"
	"errors"
	"github.com/swimresults/start-service/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

var registrationCollection *mongo.Collection

func registrationService(database *mongo.Database) {
	registrationCollection = database.Collection("registration")
}

func getRegistrationsByBsonDocument(d interface{}) ([]model.Registration, error) {
	queryOptions := options.FindOptions{}
	queryOptions.SetSort(bson.D{{"added_at", 1}})

	return getRegistrationsByBsonDocumentWithOptions(d, &queryOptions)
}

func getRegistrationsByBsonDocumentWithOptions(d interface{}, queryOptions *options.FindOptions) ([]model.Registration, error) {
	var registrations []model.Registration

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := registrationCollection.Find(ctx, d, queryOptions)
	if err != nil {
		return []model.Registration{}, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var registration model.Registration
		cursor.Decode(&registration)
		registrations = append(registrations, registration)
	}

	if err := cursor.Err(); err != nil {
		return []model.Registration{}, err
	}

	return registrations, nil
}

func getRegistrationByBsonDocument(d interface{}) (model.Registration, error) {
	queryOptions := options.FindOptions{}
	queryOptions.SetSort(bson.D{{"added_at", 1}})

	return getRegistrationByBsonDocumentWithOptions(d, &queryOptions)
}

func getRegistrationByBsonDocumentWithOptions(d interface{}, queryOptions *options.FindOptions) (model.Registration, error) {
	registrations, err := getRegistrationsByBsonDocumentWithOptions(d, queryOptions)

	if err != nil {
		return model.Registration{}, err
	}

	if len(registrations) > 0 {
		return registrations[0], nil
	}

	return model.Registration{}, errors.New("no entry found")
}

func GetRegistrationById(id primitive.ObjectID) (model.Registration, error) {
	return getRegistrationByBsonDocument(bson.D{{"_id", id}})
}

func GetRegistrationsByMeeting(meeting string) ([]model.Registration, error) {
	return getRegistrationsByBsonDocument(bson.D{{"meeting", meeting}})
}

func GetRegistrationByMeetingAndUser(meeting string, userId primitive.ObjectID) (model.Registration, error) {
	return getRegistrationByBsonDocument(bson.D{{"meeting", meeting}, {"creator_user_id", userId}})
}

func RemoveRegistrationById(id primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := registrationCollection.DeleteOne(ctx, bson.D{{"_id", id}})
	if err != nil {
		return err
	}
	return nil
}

func AddRegistration(registration model.Registration) (model.Registration, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	registration.AddedAt = time.Now()
	registration.UpdatedAt = time.Now()

	r, err := registrationCollection.InsertOne(ctx, registration)
	if err != nil {
		return model.Registration{}, err
	}

	return GetRegistrationById(r.InsertedID.(primitive.ObjectID))
}

func UpdateRegistration(registration model.Registration) (model.Registration, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	registration.UpdatedAt = time.Now()

	_, err := registrationCollection.ReplaceOne(ctx, bson.D{{"_id", registration.Identifier}}, registration)
	if err != nil {
		return model.Registration{}, err
	}

	return GetRegistrationById(registration.Identifier)
}
