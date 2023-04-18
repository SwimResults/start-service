package service

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"sr-start/start-service/model"
	"time"
)

var collection *mongo.Collection

func startService(database *mongo.Database) {
	collection = database.Collection("start")
}

func GetStarts() ([]model.Start, error) {
	var starts []model.Start

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return []model.Start{}, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var start model.Start
		cursor.Decode(&start)
		starts = append(starts, start)
	}

	if err := cursor.Err(); err != nil {
		return []model.Start{}, err
	}

	return starts, nil
}

func GetStartById(id primitive.ObjectID) (model.Start, error) {
	var start model.Start

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.D{{"_id", id}})
	if err != nil {
		return model.Start{}, err
	}
	defer cursor.Close(ctx)

	if cursor.Next(ctx) {
		cursor.Decode(&start)
		return start, nil
	}

	return model.Start{}, errors.New("no entry with given id found")
}

func RemoveStartById(id primitive.ObjectID) error {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.DeleteOne(ctx, bson.D{{"_id", id}})
	if err != nil {
		return err
	}
	return nil
}

func AddStart(start model.Start) (model.Start, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	r, err := collection.InsertOne(ctx, start)
	if err != nil {
		return model.Start{}, err
	}

	return GetStartById(r.InsertedID.(primitive.ObjectID))
}

func UpdateStart(start model.Start) (model.Start, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.ReplaceOne(ctx, bson.D{{"_id", start.Identifier}}, start)
	if err != nil {
		return model.Start{}, err
	}

	return GetStartById(start.Identifier)
}
