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

func getStartsByBsonDocument(d primitive.D) ([]model.Start, error) {
	var starts []model.Start

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, d)
	if err != nil {
		return []model.Start{}, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var start model.Start
		cursor.Decode(&start)
		if !start.DisqualificationId.IsZero() {
			start.Disqualification, _ = GetDisqualificationById(start.DisqualificationId)
		}
		start.Heat, _ = GetHeatById(start.HeatId)
		starts = append(starts, start)
	}

	if err := cursor.Err(); err != nil {
		return []model.Start{}, err
	}

	return starts, nil
}

func GetStartById(id primitive.ObjectID) (model.Start, error) {
	starts, err := getStartsByBsonDocument(bson.D{{"_id", id}})
	if err != nil {
		return model.Start{}, err
	}

	if len(starts) > 0 {
		return starts[0], nil
	}

	return model.Start{}, errors.New("no entry with given id found")
}

func GetStarts() ([]model.Start, error) {
	return getStartsByBsonDocument(bson.D{})
}

func GetStartsByMeeting(meeting string) ([]model.Start, error) {
	return getStartsByBsonDocument(bson.D{{"meeting", meeting}})
}

func GetStartsByMeetingAndAthlete(meeting string, athlete primitive.ObjectID) ([]model.Start, error) {
	return getStartsByBsonDocument(bson.D{{"meeting", meeting}, {"athlete", athlete}})
}

func GetStartsByMeetingAndEvent(meeting string, event primitive.ObjectID) ([]model.Start, error) {
	var result []model.Start
	starts, err := getStartsByBsonDocument(bson.D{{"meeting", meeting}})
	if err != nil {
		return []model.Start{}, err
	}
	for _, start := range starts {
		if start.Heat.EventId == event {
			result = append(result, start)
		}
	}
	return result, nil
}

func GetStartsByMeetingAndEventAndHeat(meeting string, event primitive.ObjectID, heat int) ([]model.Start, error) {
	var result []model.Start
	starts, err := getStartsByBsonDocument(bson.D{{"meeting", meeting}})
	if err != nil {
		return []model.Start{}, err
	}
	for _, start := range starts {
		if start.Heat.EventId == event && start.Heat.Number == heat {
			result = append(result, start)
		}
	}
	return result, nil
}

func GetStartsByAthlete(athlete primitive.ObjectID) ([]model.Start, error) {
	return getStartsByBsonDocument(bson.D{{"athlete", athlete}})
}

func RemoveStartById(id primitive.ObjectID) error {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var start, err = GetStartById(id)
	if err != nil {
		return err
	}

	if !start.DisqualificationId.IsZero() {
		var err = RemoveDisqualificationById(start.DisqualificationId)
		if err != nil {
			return err
		}
	}

	_, err = collection.DeleteOne(ctx, bson.D{{"_id", id}})
	if err != nil {
		return err
	}
	return nil
}

func AddStart(start model.Start) (model.Start, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if !start.Disqualification.Identifier.IsZero() {
		start.DisqualificationId = start.Disqualification.Identifier
	}
	if !start.Heat.Identifier.IsZero() {
		start.HeatId = start.Heat.Identifier
	}

	start.AddedAt = time.Now()
	start.UpdatedAt = time.Now()

	r, err := collection.InsertOne(ctx, start)
	if err != nil {
		return model.Start{}, err
	}

	return GetStartById(r.InsertedID.(primitive.ObjectID))
}

func UpdateStart(start model.Start) (model.Start, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if !start.Disqualification.Identifier.IsZero() {
		start.DisqualificationId = start.Disqualification.Identifier
	}
	if !start.Heat.Identifier.IsZero() {
		start.HeatId = start.Heat.Identifier
	}
	start.UpdatedAt = time.Now()

	_, err := collection.ReplaceOne(ctx, bson.D{{"_id", start.Identifier}}, start)
	if err != nil {
		return model.Start{}, err
	}

	return GetStartById(start.Identifier)
}
