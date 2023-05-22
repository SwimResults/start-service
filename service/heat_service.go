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

var heatCollection *mongo.Collection

func heatService(database *mongo.Database) {
	heatCollection = database.Collection("heat")
}

func GetHeats() ([]model.Heat, error) {
	var heats []model.Heat

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := heatCollection.Find(ctx, bson.M{})
	if err != nil {
		return []model.Heat{}, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var heat model.Heat
		cursor.Decode(&heat)
		heats = append(heats, heat)
	}

	if err := cursor.Err(); err != nil {
		return []model.Heat{}, err
	}

	return heats, nil
}

func GetHeatById(id primitive.ObjectID) (model.Heat, error) {
	var heat model.Heat

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := heatCollection.Find(ctx, bson.D{{"_id", id}})
	if err != nil {
		return model.Heat{}, err
	}
	defer cursor.Close(ctx)

	if cursor.Next(ctx) {
		cursor.Decode(&heat)
		return heat, nil
	}

	return model.Heat{}, errors.New("no entry with given id found")
}

func GetHeatByNumber(meeting string, event string, number int) (model.Heat, error) {
	var heat model.Heat

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := heatCollection.Find(ctx, bson.D{{"meeting", meeting}, {"event", event}, {"number", number}})
	if err != nil {
		return model.Heat{}, err
	}
	defer cursor.Close(ctx)

	if cursor.Next(ctx) {
		cursor.Decode(&heat)
		return heat, nil
	}

	return model.Heat{}, errors.New("no entry with given id found")
}

func RemoveHeatById(id primitive.ObjectID) error {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := heatCollection.DeleteOne(ctx, bson.D{{"_id", id}})
	if err != nil {
		return err
	}
	return nil
}

func AddHeat(heat model.Heat) (model.Heat, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	r, err := heatCollection.InsertOne(ctx, heat)
	if err != nil {
		return model.Heat{}, err
	}

	return GetHeatById(r.InsertedID.(primitive.ObjectID))
}

func UpdateHeat(heat model.Heat) (model.Heat, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := heatCollection.ReplaceOne(ctx, bson.D{{"_id", heat.Identifier}}, heat)
	if err != nil {
		return model.Heat{}, err
	}

	return GetHeatById(heat.Identifier)
}
