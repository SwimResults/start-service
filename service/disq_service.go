package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/swimresults/start-service/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

var disqualificationCollection *mongo.Collection

func disqualificationService(database *mongo.Database) {
	disqualificationCollection = database.Collection("disqualification")
}

func GetDisqualifications() ([]model.Disqualification, error) {
	var disqualifications []model.Disqualification

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := disqualificationCollection.Find(ctx, bson.M{})
	if err != nil {
		return []model.Disqualification{}, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var disqualification model.Disqualification
		cursor.Decode(&disqualification)
		disqualifications = append(disqualifications, disqualification)
	}

	if err := cursor.Err(); err != nil {
		return []model.Disqualification{}, err
	}

	return disqualifications, nil
}

func GetDisqualificationById(id primitive.ObjectID) (model.Disqualification, error) {
	var disqualification model.Disqualification

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := disqualificationCollection.Find(ctx, bson.D{{"_id", id}})
	if err != nil {
		return model.Disqualification{}, err
	}
	defer cursor.Close(ctx)

	if cursor.Next(ctx) {
		cursor.Decode(&disqualification)
		return disqualification, nil
	}

	return model.Disqualification{}, errors.New("no entry with given id found")
}

func RemoveDisqualificationById(id primitive.ObjectID) error {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := disqualificationCollection.DeleteOne(ctx, bson.D{{"_id", id}})
	if err != nil {
		return err
	}
	return nil
}

func ImportDisqualification(start model.Start, disqualification model.Disqualification) (*model.Disqualification, bool, error) {
	existing, found, err := GetStartFromImport(start)
	if err != nil {
		return nil, false, err
	}
	if !found {
		return nil, false, fmt.Errorf("start with given information not found")
	}

	var newDisqualification model.Disqualification
	var err2 error
	if existing.DisqualificationId.IsZero() {
		newDisqualification, err2 = AddDisqualification(disqualification)
	} else {
		disqualification.Identifier = existing.DisqualificationId
		newDisqualification, err2 = UpdateDisqualification(disqualification)
	}
	if err2 != nil {
		return &newDisqualification, false, err2
	}
	err3 := UpdateStartSetDisqualification(existing.Identifier, newDisqualification.Identifier)
	if err3 != nil {
		return nil, true, err3
	}
	return &newDisqualification, true, nil
}

func AddDisqualification(disqualification model.Disqualification) (model.Disqualification, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	r, err := disqualificationCollection.InsertOne(ctx, disqualification)
	if err != nil {
		return model.Disqualification{}, err
	}

	return GetDisqualificationById(r.InsertedID.(primitive.ObjectID))
}

func UpdateDisqualification(disqualification model.Disqualification) (model.Disqualification, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := disqualificationCollection.ReplaceOne(ctx, bson.D{{"_id", disqualification.Identifier}}, disqualification)
	if err != nil {
		return model.Disqualification{}, err
	}

	return GetDisqualificationById(disqualification.Identifier)
}
