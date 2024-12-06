package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/swimresults/service-core/misc"
	"github.com/swimresults/start-service/dto"
	"github.com/swimresults/start-service/model"
	"github.com/swimresults/start-service/notification"
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

func getHeatsByBsonDocument(d interface{}) ([]model.Heat, error) {
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
					delay, e := getDelayForHeat(heat)
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

	// TODO current heat sorting by start_at AND finished_at (UNION)

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

func GetCurrentAndNextHeat(meeting string) (*dto.CurrentNextHeatDto, error) {
	var dto dto.CurrentNextHeatDto
	current, err := GetCurrentHeat(meeting)
	if err != nil {
		if err.Error() == "no entry found" {
			dto.Current = nil
		} else {
			return nil, err
		}
	} else {
		dto.Current = &current
	}

	next, err := getHeatByBsonDocumentWithOptions(
		bson.M{
			"$and": []interface{}{
				bson.M{"meeting": meeting},
				bson.M{
					"start_estimation": bson.M{"$gt": current.StartEstimation},
				},
			},
		},
		*options.Find().SetLimit(1).SetSort(bson.D{{"start_estimation", 1}}),
		true,
	)
	if err != nil {
		if err.Error() == "no entry found" {
			dto.Next = nil
		} else {
			return nil, err
		}
	} else {
		dto.Next = &next
	}

	return &dto, nil
}

func GetHeats() ([]model.Heat, error) {
	return getHeatsByBsonDocument(bson.D{})
}

func GetHeatsByMeeting(id string) ([]model.Heat, error) {
	return getHeatsByBsonDocument(bson.D{{"meeting", id}})
}

func GetHeatsByMeetingAndEvent(id string, event int) ([]model.Heat, error) {
	return getHeatsByBsonDocument(bson.D{{"meeting", id}, {"event", event}})
}

func GetHeatsByMeetingAndEvents(id string, events []int) ([]model.Heat, error) {
	return getHeatsByBsonDocument(
		bson.M{
			"$and": []interface{}{
				bson.M{"meeting": id},
				bson.M{"event": bson.M{"$in": events}},
			},
		})
}

func GetHeatById(id primitive.ObjectID) (model.Heat, error) {
	return getHeatByBsonDocument(bson.D{{"_id", id}})
}

func GetHeatByNumber(meeting string, event int, number int) (model.Heat, error) {
	return getHeatByBsonDocument(bson.D{{"meeting", meeting}, {"event", event}, {"number", number}})
}

func GetHeatsAmount() (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Count().SetHint("_id_")
	count, err := heatCollection.CountDocuments(ctx, bson.D{}, opts)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func GetHeatsAmountByMeeting(meeting string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Count().SetHint("_id_")
	count, err := heatCollection.CountDocuments(ctx, bson.D{{"meeting", meeting}}, opts)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func GetHeatInfoByMeetingAndEvent(meeting string, event int) (dto.EventHeatInfoDto, error) {
	var info dto.EventHeatInfoDto
	count, err := GetHeatsAmountByMeetingAndEvent(meeting, event)
	if err != nil {
		return dto.EventHeatInfoDto{}, err
	}
	info.Amount = count
	return info, nil
}

func GetHeatsAmountByMeetingAndEvent(meeting string, event int) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Count().SetHint("_id_")
	count, err := heatCollection.CountDocuments(ctx, bson.D{{"meeting", meeting}, {"event", event}}, opts)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func GetHeatsByMeetingForEventList(meeting string) (dto.MeetingHeatsEventListDto, error) {
	var info dto.MeetingHeatsEventListDto
	heats, err := getHeatsByBsonDocument(bson.D{{"meeting", meeting}, {"number", 1}})
	if err != nil {
		return dto.MeetingHeatsEventListDto{}, err
	}

	for _, heat := range heats {
		var infoTile dto.MeetingHeatEventListDto
		var err1 error
		infoTile.EventNumber = heat.Event
		infoTile.FirstHeat = heat
		infoTile.Amount, err1 = GetHeatsAmountByMeetingAndEvent(meeting, heat.Event)
		if err1 != nil {
			return dto.MeetingHeatsEventListDto{}, err
		}
		info.Events = append(info.Events, infoTile)
	}

	return info, nil
}

func GetHeatsByMeetingForEventListEvents(meeting string, events []int) (dto.MeetingHeatsEventListDto, error) {
	var info dto.MeetingHeatsEventListDto
	heats, err := getHeatsByBsonDocument(
		bson.M{
			"$and": []interface{}{
				bson.M{"meeting": meeting},
				bson.M{"number": 1},
				bson.M{"event": bson.M{"$in": events}},
			},
		})
	if err != nil {
		return dto.MeetingHeatsEventListDto{}, err
	}

	for _, heat := range heats {
		var infoTile dto.MeetingHeatEventListDto
		var err1 error
		infoTile.EventNumber = heat.Event
		infoTile.FirstHeat = heat
		infoTile.Amount, err1 = GetHeatsAmountByMeetingAndEvent(meeting, heat.Event)
		if err1 != nil {
			return dto.MeetingHeatsEventListDto{}, err
		}
		info.Events = append(info.Events, infoTile)
	}

	return info, nil
}

func GetHeatsWithStartWithinDurationAndNoNotification(distance time.Duration) ([]model.Heat, error) {
	now := time.Now()
	then := now.Add(distance)
	return getHeatsByBsonDocument(
		bson.M{
			"$and": []interface{}{
				bson.M{"start_delay_estimation": bson.M{"$gt": now, "$lt": then}},
				bson.M{"start_soon_notified": false},
			},
		})
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
func getDelayForHeat(heat model.Heat) (delay *time.Duration, err error) {

	fmt.Println("need to fetch delay")

	var t1 time.Time
	var t2 time.Time

	heat, err1 := getHeatByBsonDocumentWithOptions(
		// ((smaller event) OR (same event, smaller heat)) AND ((start_delay_estimation exists) OR (started_at exists))
		bson.M{
			"$and": []interface{}{
				bson.M{"meeting": heat.Meeting},
				// old way of finding previous heats, but event numbers can be not in order
				/*bson.M{
					"$or": []interface{}{
						bson.M{
							"event": bson.M{"$lt": event},
						},
						bson.M{
							"event":  event,
							"number": bson.M{"$lt": heatNumber},
						},
					},
				},*/
				// new way with time estimation…
				bson.M{
					"start_estimation": bson.M{"$lt": heat.StartEstimation},
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
		//*options.Find().SetLimit(1).SetSort(bson.D{{"event", -1}, {"number", -1}}),
		*options.Find().SetLimit(1).SetSort(bson.D{{"start_estimation", -1}}),
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

	fmt.Printf("start at: %s – estimation: %s\n", t2s, t1s)

	t1t, _ := time.Parse(layout, t1s)
	t2t, _ := time.Parse(layout, t2s)

	de := t2t.Sub(t1t)

	fmt.Printf("delay: %fmin\n", de.Minutes())

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

// ImportHeat imports a heat; returns the created or existing heat and a boolean if it was created or already present
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
	if !heat.FinishedAt.IsZero() {
		existing.FinishedAt = heat.FinishedAt
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

func UpdateHeatTimes(id primitive.ObjectID, time time.Time, timeType string) (model.Heat, error) {
	heat, err := GetHeatById(id)
	if err != nil {
		return model.Heat{}, err
	}

	switch timeType {
	case "start_delay_estimation":
		heat.StartDelayEstimation = time
		break
	case "start_at":
		heat.StartAt = time
		break
	case "finished_at":
		heat.FinishedAt = time
		break
	}

	return UpdateHeat(heat)
}

func SetHeatStartToNowByNumber(meeting string, event int, number int) (model.Heat, error) {
	heat, err := GetHeatByNumber(meeting, event, number)
	if err != nil {
		return model.Heat{}, err
	}

	heat.StartAt = misc.TimeNow()

	go notification.BroadcastHeatStart(meeting, event, number, int(heat.StartEstimation.Sub(heat.StartAt).Seconds()))

	return UpdateHeat(heat)
}

func UpdateHeatsEstimationDateByMeetingAndEvent(meeting string, events []int, t time.Time, updateTimeZone bool) ([]model.Heat, error) {
	var heats []model.Heat
	var err error
	if len(events) <= 0 {
		heats, err = GetHeatsByMeeting(meeting)
		println("changing heat date and timezone for ALL events")
	} else {
		heats, err = GetHeatsByMeetingAndEvents(meeting, events)
		println("changing heat date and timezone for events:")
		println(events)
	}

	if err != nil {
		return []model.Heat{}, err
	}

	var savedHeats []model.Heat

	for _, heat := range heats {
		t2 := heat.StartEstimation

		var timezone *time.Location
		if updateTimeZone {
			timezone = t.Location()
		} else {
			timezone = t2.Location()
		}

		heat.StartEstimation = time.Date(t.Year(), t.Month(), t.Day(), t2.Hour(), t2.Minute(), t2.Second(), t2.Nanosecond(), timezone)

		saved, err := UpdateHeat(heat)
		if err != nil {
			return []model.Heat{}, err
		}

		savedHeats = append(savedHeats, saved)
	}

	return savedHeats, nil
}

func UpdateHeatEstimationDate(id primitive.ObjectID, t time.Time) (model.Heat, error) {
	heat, err := GetHeatById(id)
	if err != nil {
		return model.Heat{}, err
	}

	t2 := heat.StartEstimation

	heat.StartEstimation = time.Date(t.Year(), t.Month(), t.Day(), t2.Hour(), t2.Minute(), t2.Second(), t2.Nanosecond(), t2.Location())

	return UpdateHeat(heat)
}
