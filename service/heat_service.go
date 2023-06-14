package service

import (
	"context"
	"errors"
	"github.com/swimresults/start-service/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

var heatCollection *mongo.Collection

func heatService(database *mongo.Database) {
	heatCollection = database.Collection("heat")
}

func getHeatsByBsonDocument(d primitive.D) ([]model.Heat, error) {
	var heats []model.Heat

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := heatCollection.Find(ctx, d)
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

func GetHeats() ([]model.Heat, error) {
	return getHeatsByBsonDocument(bson.D{})
}

func GetHeatsByMeeting(id string) ([]model.Heat, error) {
	return getHeatsByBsonDocument(bson.D{{"meeting", id}})
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

func GetHeatByNumber(meeting string, event int, number int) (model.Heat, error) {
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

	return model.Heat{}, errors.New("no entry with given number and event found in meeting")
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

func ImportHeat(heat model.Heat) (model.Heat, bool, error) {
	if heat.Meeting == "" || heat.Event == 0 || heat.Number == 0 {
		return model.Heat{}, false, errors.New("missing arguments (meeting/event/heat is needed)")
	}

	existing, err := GetHeatByNumber(heat.Meeting, heat.Event, heat.Number)
	if err != nil {
		if err.Error() == "no entry with given number and event found in meeting" {
			newHeat, err2 := AddHeat(heat)
			if err2 != nil {
				return model.Heat{}, false, err2
			}
			return newHeat, true, nil
		}
		return model.Heat{}, false, err
	}
	return existing, false, nil
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
