package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/swimresults/service-core/misc"
	"github.com/swimresults/start-service/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

var collection *mongo.Collection

func startService(database *mongo.Database) {
	collection = database.Collection("start")
}

func getStartsByBsonDocument(d interface{}) ([]model.Start, error) {
	var starts []model.Start

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	queryOptions := options.FindOptions{}
	queryOptions.SetSort(bson.D{{"event", 1}, {"heat_number", 1}, {"lane", 1}})

	cursor, err := collection.Find(ctx, d, &queryOptions)
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
		start.Heat, _ = GetHeatByNumber(start.Meeting, start.Event, start.HeatNumber)
		starts = append(starts, start)
	}

	if err := cursor.Err(); err != nil {
		return []model.Start{}, err
	}

	return starts, nil
}

func getStartByBsonDocument(d interface{}) (model.Start, error) {
	starts, err := getStartsByBsonDocument(d)

	if err != nil {
		return model.Start{}, err
	}

	if len(starts) > 0 {
		return starts[0], nil
	}

	return model.Start{}, errors.New("no entry found")
}

func GetStartById(id primitive.ObjectID) (model.Start, error) {
	return getStartByBsonDocument(bson.D{{"_id", id}})
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

func GetStartsByMeetingAndEvent(meeting string, event int) ([]model.Start, error) {
	return getStartsByBsonDocument(bson.D{{"meeting", meeting}, {"event", event}})
}

func GetStartsByMeetingAndEventAndHeat(meeting string, event int, heat int) ([]model.Start, error) {
	return getStartsByBsonDocument(bson.D{{"meeting", meeting}, {"event", event}, {"heat", heat}})
}

func GetStartByMeetingAndEventAndHeatAndLane(meeting string, event int, heat int, lane int) (model.Start, error) {
	return getStartByBsonDocument(
		bson.D{
			{"meeting", meeting},
			{"event", event},
			{"heat", heat},
			{"lane", lane},
		},
	)
}

func GetStartByMeetingAndEventAndAthleteMeetingId(meeting string, event int, athleteMeetingId int) (model.Start, error) {
	return getStartByBsonDocument(bson.D{
		{"meeting", meeting},
		{"event", event},
		{"athlete_meeting_id", athleteMeetingId},
	})
}

func GetStartByMeetingAndEventAndAthleteNameAndYear(meeting string, event int, athleteName string, year int) (model.Start, error) {
	if hasComma, first, last := misc.ExtractNames(athleteName); hasComma {
		athleteName = first + " " + last
	}

	return getStartByBsonDocument(bson.M{
		"$and": []interface{}{
			bson.M{"meeting": meeting},
			bson.M{"event": event},
			bson.M{"athlete_year": year},
			bson.M{
				"$or": []interface{}{
					bson.M{"name": bson.M{"$regex": athleteName, "$options": "i"}},
					bson.M{"alias": bson.M{"$regex": misc.Aliasify(athleteName), "$options": "i"}},
				},
			},
		},
	})
}

func GetStartByMeetingAndEventAndAthleteId(meeting string, event int, athleteId primitive.ObjectID) (model.Start, error) {
	return getStartByBsonDocument(bson.D{
		{"meeting", meeting},
		{"event", event},
		{"athlete", athleteId},
	})
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

	start.AddedAt = time.Now()
	start.UpdatedAt = time.Now()

	r, err := collection.InsertOne(ctx, start)
	if err != nil {
		return model.Start{}, err
	}

	return GetStartById(r.InsertedID.(primitive.ObjectID))
}

func GetStartFromImport(start model.Start) (model.Start, bool, error) {

	if start.Meeting == "" ||
		start.Event == 0 {
		return model.Start{}, false, fmt.Errorf("missing arguments"+
			"(expected: meeting; event; ..."+
			"got: '%s', '%d')",
			start.Meeting,
			start.Event)
	}

	var existing model.Start
	var err error

	// TODO: DSV Lane representation (0?)
	if start.HeatNumber != 0 && start.Lane >= 0 {
		existing, err = GetStartByMeetingAndEventAndHeatAndLane(start.Meeting, start.Event, start.HeatNumber, start.Lane)
		if err != nil {
			if err.Error() != "no entry found" {
				return model.Start{}, false, err
			}
		} else {
			return existing, true, nil
		}
	}

	if start.AthleteName == "" || start.AthleteYear == 0 {
		return model.Start{}, false, fmt.Errorf("missing arguments"+
			"(expected: athlete_name; athlete_year; ..."+
			"got: '%s', '%d')",
			start.AthleteName,
			start.AthleteYear)
	}

	if start.AthleteMeetingId != 0 {
		existing, err = GetStartByMeetingAndEventAndAthleteMeetingId(start.Meeting, start.Event, start.AthleteMeetingId)
		if err != nil {
			if err.Error() != "no entry found" {
				return model.Start{}, false, err
			}
		} else {
			return existing, true, nil
		}
	}

	if start.AthleteName != "" && start.AthleteYear != 0 {
		existing, err = GetStartByMeetingAndEventAndAthleteNameAndYear(start.Meeting, start.Event, start.AthleteName, start.AthleteYear)
		if err != nil {
			if err.Error() != "no entry found" {
				return model.Start{}, false, err
			}
		} else {
			return existing, true, nil
		}
	}

	if start.AthleteName != "" && start.AthleteYear != 0 && athleteClient != nil {
		athlete, found2, err2 := athleteClient.GetAthleteByNameAndYear(start.AthleteName, start.AthleteYear)
		if err2 != nil {
			return model.Start{}, false, err2
		}
		if found2 {
			existing, err = GetStartByMeetingAndEventAndAthleteId(start.Meeting, start.Event, athlete.Identifier)
			if err != nil {
				if err.Error() != "no entry found" {
					return model.Start{}, false, err
				}
			} else {
				return existing, true, nil
			}
		}
	}

	return model.Start{}, false, nil
}

func ImportStart(start model.Start) (*model.Start, bool, error) {
	// looks for existing:
	// 		DSV 			(event, athleteMeetingId)
	//		PDF 			(event, athlete name, year)
	//		start list PDF 	(event, heat, lane)
	// 		special:	look for athlete with given name
	//						(only as last option because of synchronous request and reliability on external service)
	// if !existing:
	// 		save athlete name and alias
	//		look up athlete
	//		look up team
	//		save
	//
	// else:
	//		update fields

	var err error
	existing, found, err := GetStartFromImport(start)

	if !found {
		if start.AthleteTeamName == "" {
			return nil, false, fmt.Errorf("missing argument"+
				"(expected: athlete_team_name; since start isn't existing ..."+
				"got: '%s')",
				start.AthleteTeamName)
		}
		// create start
		// get athleteID
		athlete, f, err3 := athleteClient.GetAthleteByNameAndYear(start.AthleteName, start.AthleteYear)
		if err3 != nil {
			return nil, false, err3
		}
		if !f {
			return nil, false, fmt.Errorf("athlete with given AthleteName '%s' was not found", start.AthleteName)
		}
		start.Athlete = athlete.Identifier

		// get teamID
		team, f2, err4 := teamClient.GetTeamByName(start.AthleteTeamName)
		if err4 != nil {
			return nil, false, err4
		}
		if !f2 {
			return nil, false, fmt.Errorf("team with given AthleteTeamName '%s' was not found", start.AthleteTeamName)
		}

		start.AthleteTeam = team.Identifier

		// set name values
		if hasComma, first, last := misc.ExtractNames(start.AthleteName); hasComma {
			start.AthleteName = first + " " + last
		}
		start.AthleteAlias = misc.Aliasify(start.AthleteName)

		// save new start
		newStart, err2 := AddStart(start)
		if err2 != nil {
			return nil, false, err2
		}
		fmt.Printf("import of start '%s/%d/%d/%d', was created\n", start.Meeting, start.Event, start.HeatNumber, start.Lane)

		return &newStart, true, nil
	}

	fmt.Printf("import of start '%s/%d/%d/%d', already present\n", start.Meeting, start.Event, start.HeatNumber, start.Lane)

	changed := false
	if existing.Certified == false && start.Certified == true {
		existing.Certified = start.Certified
		changed = true
	}
	if existing.Rank == 0 && start.Rank != 0 {
		existing.Rank = start.Rank
		changed = true
	}
	if existing.AthleteMeetingId == 0 && start.AthleteMeetingId != 0 {
		existing.AthleteMeetingId = start.AthleteMeetingId
		changed = true
	}
	if existing.AthleteName == "" && start.AthleteName != "" {
		existing.AthleteName = start.AthleteName
		changed = true
	}
	if existing.Lane == 0 && start.Lane != 0 {
		existing.Lane = start.Lane
		changed = true
	}
	if existing.HeatNumber == 0 && start.HeatNumber != 0 {
		existing.HeatNumber = start.HeatNumber
		changed = true
	}

	if changed {
		fmt.Printf("updating some values...\n")
		existing, err = UpdateStart(existing)
		if err != nil {
			return nil, false, err
		}
	}
	return &existing, false, nil
}

func ImportResult(start model.Start, result model.Result) (*model.Result, bool, error) {
	existing, found, err := GetStartFromImport(start)
	if err != nil {
		return nil, false, err
	}
	if !found {
		return nil, false, fmt.Errorf("start with given information not found")
	}
	res, err2 := UpdateStartAddResult(existing.Identifier, result)
	if err2 != nil {
		return nil, false, err2
	}
	return &res, true, nil
}

func UpdateStart(start model.Start) (model.Start, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if !start.Disqualification.Identifier.IsZero() {
		start.DisqualificationId = start.Disqualification.Identifier
	}
	start.UpdatedAt = time.Now()

	_, err := collection.ReplaceOne(ctx, bson.D{{"_id", start.Identifier}}, start)
	if err != nil {
		return model.Start{}, err
	}

	return GetStartById(start.Identifier)
}

func UpdateStartSetDisqualification(startId primitive.ObjectID, disqualificationId primitive.ObjectID) error {
	start, err := GetStartById(startId)
	if err != nil {
		return err
	}
	start.DisqualificationId = disqualificationId
	_, err2 := UpdateStart(start)
	if err2 != nil {
		return err2
	}
	return nil
}

func UpdateStartAddResult(startId primitive.ObjectID, result model.Result) (model.Result, error) {
	start, err := GetStartById(startId)
	if err != nil {
		return model.Result{}, err
	}
	result.AddedAt = time.Now()
	start.Results = append(start.Results, result)
	_, err2 := UpdateStart(start)
	if err2 != nil {
		return model.Result{}, err2
	}
	return result, nil
}
