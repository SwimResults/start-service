package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/swimresults/start-service/dto"
	"github.com/swimresults/start-service/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

var heatCollection *mongo.Collection

func heatService(database *mongo.Database) {
	heatCollection = database.Collection("heat")
}

func getHeatsByBsonDocument(d primitive.D) ([]model.Heat, error) {
	return getHeatsByBsonDocumentWithOptions(d, options.FindOptions{}, true)
}

func getHeatsByBsonDocumentWithOptions(d interface{}, fOps options.FindOptions, fetchDelay bool) ([]model.Heat, error) {
	var heats []model.Heat

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := heatCollection.Find(ctx, d, &fOps)
	if err != nil {
		return []model.Heat{}, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var heat model.Heat
		cursor.Decode(&heat)

		// set delay
		if !heat.StartAt.IsZero() {
			heat.StartDelayEstimation = heat.StartAt
		} else {
			if heat.StartDelayEstimation.IsZero() {
				// no time information in current heat, calculating delay
				if fetchDelay {
					delay, e := getDelayForHeat(heat.Meeting, heat.Event, heat.Number)
					if e == nil {
						// add delay to estimation
						heat.StartDelayEstimation = heat.StartEstimation.Add(*delay)
					}
				}
			}
		}

		heats = append(heats, heat)
	}

	if err := cursor.Err(); err != nil {
		return []model.Heat{}, err
	}

	return heats, nil
}

func getHeatByBsonDocument(d interface{}) (model.Heat, error) {
	return getHeatByBsonDocumentWithOptions(d, options.FindOptions{}, true)
}

func getHeatByBsonDocumentWithOptions(d interface{}, fOps options.FindOptions, fetchDelay bool) (model.Heat, error) {
	heats, err := getHeatsByBsonDocumentWithOptions(d, fOps, fetchDelay)
	if err != nil {
		return model.Heat{}, err
	}
	if len(heats) <= 0 {
		return model.Heat{}, errors.New("no entry found")
	}

	return heats[0], nil
}

func GetCurrentHeat(meeting string) (model.Heat, error) {
	queryOptions := options.FindOptions{}
	queryOptions.SetSort(bson.D{{"start_at", -1}, {"finished_at", -1}})

	return getHeatByBsonDocumentWithOptions(
		bson.M{
			"$and": []interface{}{
				bson.M{"meeting": meeting},
				bson.M{
					"$or": []interface{}{
						bson.M{"start_at": bson.M{"$exists": true}},
						bson.M{"finished_at": bson.M{"$exists": true}},
					},
				},
			},
		},
		*options.Find().SetLimit(1).SetSort(bson.D{{"start_at", -1}, {"finished_at", -1}}),
		true,
	)
}

func GetHeats() ([]model.Heat, error) {
	return getHeatsByBsonDocument(bson.D{})
}

func GetHeatsByMeeting(id string) ([]model.Heat, error) {
	return getHeatsByBsonDocument(bson.D{{"meeting", id}})
}

func GetHeatById(id primitive.ObjectID) (model.Heat, error) {
	return getHeatByBsonDocument(bson.D{{"_id", id}})
}

func GetHeatByNumber(meeting string, event int, number int) (model.Heat, error) {
	return getHeatByBsonDocument(bson.D{{"meeting", meeting}, {"event", event}, {"number", number}})
}

func GetHeatInfoByMeetingAndEvent(meeting string, event int) (dto.EventHeatInfoDto, error) {
	var info dto.EventHeatInfoDto
	heats, err := getHeatsByBsonDocument(bson.D{{"meeting", meeting}, {"event", event}})
	if err != nil {
		return dto.EventHeatInfoDto{}, err
	}
	info.Amount = len(heats)
	return info, nil
}

func GetHeatInfoByMeeting(meeting string) (dto.MeetingHeatInfoDto, error) {
	var info dto.MeetingHeatInfoDto
	heats, err := getHeatsByBsonDocument(bson.D{{"meeting", meeting}})
	if err != nil {
		return dto.MeetingHeatInfoDto{}, err
	}
	info.Amount = len(heats)
	return info, nil
}

// returns delay of previous heat with delay information as time.Duration
// positive number, if heat is delayed
func getDelayForHeat(meeting string, event int, heatNumber int) (delay *time.Duration, err error) {

	fmt.Println("need to fetch delay")

	var t1 time.Time
	var t2 time.Time

	heat, err1 := getHeatByBsonDocumentWithOptions(
		// ((smaller event) OR (same event, smaller heat)) AND ((start_delay_estimation exists) OR (started_at exists))
		bson.M{
			"$and": []interface{}{
				bson.M{"meeting": meeting},
				bson.M{
					"$or": []interface{}{
						bson.M{
							"event": bson.M{"$lt": event},
						},
						bson.M{
							"event":  event,
							"number": bson.M{"$lt": heatNumber},
						},
					},
				},
				bson.M{
					"$or": []interface{}{
						bson.M{"start_delay_estimation": bson.M{"$exists": true}},
						bson.M{"start_at": bson.M{"$exists": true}},
					},
				},
			},
		},
		//options.FindOptions{},
		*options.Find().SetLimit(1).SetSort(bson.D{{"event", -1}, {"number", -1}}),
		false)

	if err1 != nil {
		fmt.Println(err1.Error())
		return nil, err1
	}

	if heat.Event == 0 && heat.Number == 0 {
		return nil, errors.New("invalid head found")
	}

	if heat.StartEstimation.IsZero() {
		return nil, errors.New("invalid heat found")
	}

	fmt.Printf("found heat: event %d, heat %d\n", heat.Event, heat.Number)

	t1 = heat.StartEstimation

	if !heat.StartAt.IsZero() {
		t2 = heat.StartAt
	} else if !heat.StartDelayEstimation.IsZero() {
		t2 = heat.StartDelayEstimation
	} else {
		return nil, errors.New("no delay found")
	}

	const layout = "15:04:05"

	t1s := t1.Format(layout)
	t2s := t2.Format(layout)

	t1t, _ := time.Parse(layout, t1s)
	t2t, _ := time.Parse(layout, t2s)

	de := t2t.Sub(t1t)

	return &de, nil
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
		if err.Error() == "no entry found" {
			newHeat, err2 := AddHeat(heat)
			if err2 != nil {
				return model.Heat{}, false, err2
			}
			return newHeat, true, nil
		}
		return model.Heat{}, false, err
	}

	changed := false
	if !heat.StartEstimation.IsZero() {
		existing.StartEstimation = heat.StartEstimation
		changed = true
	}
	if !heat.StartAt.IsZero() {
		existing.StartAt = heat.StartAt
		changed = true
	}
	if changed {
		existing, err = UpdateHeat(existing)
		if err != nil {
			return model.Heat{}, false, err
		}
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
