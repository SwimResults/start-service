package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/swimresults/service-core/misc"
	"github.com/swimresults/start-service/dto"
	"github.com/swimresults/start-service/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var collection *mongo.Collection

func startService(database *mongo.Database) {
	collection = database.Collection("start")
}

func getStartsByBsonDocument(d interface{}) ([]model.Start, error) {

	queryOptions := options.FindOptions{}
	queryOptions.SetSort(bson.D{{"event", 1}, {"heat_number", 1}, {"lane", 1}})

	return getStartsByBsonDocumentWithOptions(d, &queryOptions)
}

func getStartsByBsonDocumentWithOptions(d interface{}, queryOptions *options.FindOptions) ([]model.Start, error) {
	var starts []model.Start

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, d, queryOptions)
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
		start.Heat, _ = GetHeatByNumberWithoutDelay(start.Meeting, start.Event, start.HeatNumber)
		starts = append(starts, start)
	}

	if err := cursor.Err(); err != nil {
		return []model.Start{}, err
	}

	return starts, nil
}

func getStartByBsonDocument(d interface{}) (model.Start, error) {

	queryOptions := options.FindOptions{}
	queryOptions.SetSort(bson.D{{"event", 1}, {"heat_number", 1}, {"lane", 1}})

	return getStartByBsonDocumentWithOptions(d, &queryOptions)
}

func getStartByBsonDocumentWithOptions(d interface{}, queryOptions *options.FindOptions) (model.Start, error) {
	starts, err := getStartsByBsonDocumentWithOptions(d, queryOptions)

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

func GetStartsAmount() (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Count().SetHint("_id_")
	count, err := collection.CountDocuments(ctx, bson.D{}, opts)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func GetStartsAmountByMeeting(meeting string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Count().SetHint("_id_")
	count, err := collection.CountDocuments(ctx, bson.D{{"meeting", meeting}}, opts)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func GetStartsByMeeting(meeting string) ([]model.Start, error) {
	return getStartsByBsonDocument(bson.D{{"meeting", meeting}})
}

func GetStartsByMeetingAndAthlete(meeting string, athlete primitive.ObjectID) ([]model.Start, error) {
	starts, err := getStartsByBsonDocument(bson.D{{"meeting", meeting}, {"athlete", athlete}})
	if err != nil {
		return []model.Start{}, err
	}

	sort.Slice(starts, func(i, j int) bool {
		return starts[i].Heat.StartEstimation.Before(starts[j].Heat.StartEstimation)
	})
	return starts, nil
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
					bson.M{"athlete_name": bson.M{"$regex": athleteName, "$options": "i"}},
					bson.M{"athlete_alias": bson.M{"$regex": misc.Aliasify(athleteName), "$options": "i"}},
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
	starts, err := getStartsByBsonDocument(bson.D{{"athlete", athlete}})
	if err != nil {
		return []model.Start{}, err
	}

	sort.Slice(starts, func(i, j int) bool {
		return starts[i].Heat.StartEstimation.Before(starts[j].Heat.StartEstimation)
	})
	return starts, nil
}

func GetStartsByMeetingAndEventAsResults(meeting string, event int) ([]dto.EventStartResultRequestDto, error) {
	ageGroups, err := ageGroupClient.GetAgeGroupsForMeetingAndEvent(meeting, event)
	if err != nil {
		return nil, err
	}

	var results []dto.EventStartResultRequestDto

	queryOptions := options.FindOptions{}
	queryOptions.SetSort(bson.D{{"disqualification_id", 1}, {"rank", 1}})

	for _, group := range *ageGroups {
		if group.IsYear != true {
			continue
		}

		starts, err2 := getStartsByBsonDocumentWithOptions(
			bson.M{
				"$and": []interface{}{
					bson.M{"meeting": meeting},
					bson.M{"event": event},
					bson.M{"athlete_year": bson.M{"$in": group.Ages}},
				},
			}, &queryOptions)

		if err2 != nil {
			return nil, err2
		}

		result := dto.EventStartResultRequestDto{
			AgeGroup: group,
			Starts:   starts,
		}

		results = append(results, result)
	}

	return results, nil
}

func GetCurrentStarts(meeting string) ([]model.Start, error) {
	heat, err := GetCurrentHeat(meeting)
	if err != nil {
		return []model.Start{}, err
	}

	return GetStartsByMeetingAndEventAndHeat(meeting, heat.Event, heat.Number)
}

func GetStartsByMeetingStats(meeting string) ([]dto.StartsByYearAndGenderStatsDto, error) {
	if eventClient == nil {
		return nil, errors.New("event client not configured")
	}

	events, err := eventClient.GetEventsByMeetId(meeting)
	if err != nil {
		return nil, err
	}

	allowedEvents := make(map[int]string)
	for _, event := range *events {
		if event.Final.IsFinal {
			continue
		}

		if strings.TrimSpace(event.RelayDistance) != "" {
			continue
		}

		gender := strings.ToUpper(strings.TrimSpace(event.Gender))
		if gender != "MALE" && gender != "FEMALE" {
			continue
		}

		allowedEvents[event.Number] = gender
	}

	starts, err := GetStartsByMeeting(meeting)
	if err != nil {
		return nil, err
	}

	stats := make(map[int]map[string]int)
	for _, start := range starts {
		gender, ok := allowedEvents[start.Event]
		if !ok {
			continue
		}

		if start.AthleteYear <= 0 {
			continue
		}

		// Skip withdrawn starts
		if !start.DisqualificationId.IsZero() && start.Disqualification.Type == "withdrawn" {
			continue
		}

		if _, ok = stats[start.AthleteYear]; !ok {
			stats[start.AthleteYear] = map[string]int{}
		}

		stats[start.AthleteYear][gender]++
	}

	years := make([]int, 0, len(stats))
	for year := range stats {
		years = append(years, year)
	}
	sort.Ints(years)

	response := make([]dto.StartsByYearAndGenderStatsDto, 0, len(years))
	for _, year := range years {
		response = append(response, dto.StartsByYearAndGenderStatsDto{
			Year: year,
			Genders: []dto.StartsByGenderStatsDto{
				{Gender: "FEMALE", Amount: stats[year]["FEMALE"]},
				{Gender: "MALE", Amount: stats[year]["MALE"]},
			},
		})
	}

	return response, nil
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

// RemoveStartsByMeetingAndEventAndHeat removes all starts for the given meeting, event and heat.
// This is used when a heat is removed, so that no orphaned starts remain.
// It also removes the disqualifications of the starts, if they exist.
func RemoveStartsByMeetingAndEventAndHeat(meeting string, event int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var starts, err = GetStartsByMeetingAndEvent(meeting, event)
	if err != nil {
		return err
	}

	for _, start := range starts {
		if !start.DisqualificationId.IsZero() {
			var err = RemoveDisqualificationById(start.DisqualificationId)
			if err != nil {
				return err
			}
		}
	}

	_, err = collection.DeleteMany(ctx, bson.D{{"meeting", meeting}, {"event", event}})
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

	var existing model.Start
	var err error

	if !start.Identifier.IsZero() {
		existing, err = GetStartById(start.Identifier)
		if err != nil {
			debugImportStartFailure("lookup by identifier", start, err)
			return model.Start{}, false, errors.New("could not find start by identifier '" + start.Identifier.String() + "', even though it was given")
		}
		return existing, true, nil
	}

	if start.Meeting == "" ||
		start.Event == 0 {
		debugImportStartFailure("validation missing meeting or event", start, fmt.Errorf("missing arguments"))
		return model.Start{}, false, fmt.Errorf("missing arguments"+
			"(expected: meeting; event; ..."+
			"got: '%s', '%d')",
			start.Meeting,
			start.Event)
	}

	if start.HeatNumber != 0 && start.Lane >= 0 {
		existing, err = GetStartByMeetingAndEventAndHeatAndLane(start.Meeting, start.Event, start.HeatNumber, start.Lane)
		if err != nil {
			if err.Error() != "no entry found" {
				debugImportStartFailure("lookup by heat and lane", start, err)
				return model.Start{}, false, err
			}
		} else {
			return existing, true, nil
		}
	}

	if start.AthleteName == "" || start.AthleteYear == 0 {
		debugImportStartFailure("validation missing athlete name or year", start, fmt.Errorf("missing arguments"))
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
				debugImportStartFailure("lookup by athlete meeting id", start, err)
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
				debugImportStartFailure("lookup by athlete name and year", start, err)
				return model.Start{}, false, err
			}
		} else {
			return existing, true, nil
		}
	}

	if start.AthleteName != "" && start.AthleteYear != 0 && athleteClient != nil {
		athlete, found2, err2 := athleteClient.GetAthleteByNameAndYear(start.AthleteName, start.AthleteYear)
		if err2 != nil {
			debugImportStartFailure("athlete client lookup by name and year", start, err2)
			return model.Start{}, false, err2
		}
		if found2 {
			existing, err = GetStartByMeetingAndEventAndAthleteId(start.Meeting, start.Event, athlete.Identifier)
			if err != nil {
				if err.Error() != "no entry found" {
					debugImportStartFailure("lookup by athlete id", start, err)
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

	if err != nil {
		debugImportStartFailure("resolve existing start", start, err)
		return nil, false, err
	}

	if !found {
		if start.AthleteTeamName == "" {
			debugImportStartFailure("missing athlete team name", start, fmt.Errorf("missing athlete_team_name"))
			return nil, false, fmt.Errorf("missing argument"+
				"(expected: athlete_team_name; since start isn't existing ..."+
				"got: '%s')",
				start.AthleteTeamName)
		}
		// create start
		// get athleteID
		if !start.IsRelay {

			if start.Athlete.IsZero() {
				athlete, f, err3 := athleteClient.GetAthleteByNameAndYear(start.AthleteName, start.AthleteYear)
				if err3 != nil {
					debugImportStartFailure("athlete lookup while creating start", start, err3)
					return nil, false, err3
				}
				if !f {
					debugImportStartFailure("athlete not found while creating start", start, fmt.Errorf("athlete with given AthleteName '%s' was not found", start.AthleteName))
					return nil, false, fmt.Errorf("athlete with given AthleteName '%s' was not found", start.AthleteName)
				}
				start.Athlete = athlete.Identifier
			}

			// set name values
			if hasComma, first, last := misc.ExtractNames(start.AthleteName); hasComma {
				start.AthleteName = first + " " + last
			}
			start.AthleteAlias = misc.Aliasify(start.AthleteName)

		} else {
			start.Athlete = primitive.ObjectID{}
		}

		if start.AthleteTeam.IsZero() {
			// get teamID
			team, f2, err4 := teamClient.GetTeamByName(start.AthleteTeamName)
			if err4 != nil {
				debugImportStartFailure("team lookup while creating start", start, err4)
				return nil, false, err4
			}
			if !f2 {
				debugImportStartFailure("team not found while creating start", start, fmt.Errorf("team with given AthleteTeamName '%s' was not found", start.AthleteTeamName))
				return nil, false, fmt.Errorf("team with given AthleteTeamName '%s' was not found", start.AthleteTeamName)
			}

			start.AthleteTeam = team.Identifier
		}

		// save new start
		newStart, err2 := AddStart(start)
		if err2 != nil {
			debugImportStartFailure("persist new start", start, err2)
			return nil, false, err2
		}
		fmt.Printf("import of start '%s/%d/%d/%d', was created\n", start.Meeting, start.Event, start.HeatNumber, start.Lane)

		return &newStart, true, nil
	}

	fmt.Printf("import of start '%s/%d/%d/%d', already present\n", start.Meeting, start.Event, start.HeatNumber, start.Lane)

	// Normalize and enrich import values before applying updates.
	if !start.IsRelay {
		if hasComma, first, last := misc.ExtractNames(start.AthleteName); hasComma {
			start.AthleteName = first + " " + last
		}
		if start.AthleteName != "" {
			start.AthleteAlias = misc.Aliasify(start.AthleteName)
		}

		needsAthleteLookup := start.Athlete.IsZero() &&
			start.AthleteName != "" &&
			start.AthleteYear != 0 &&
			(existing.AthleteName != start.AthleteName ||
				existing.AthleteYear != start.AthleteYear ||
				existing.Athlete.IsZero())

		if needsAthleteLookup {
			if athleteClient == nil {
				debugImportStartFailure("athlete client missing during update import", start, fmt.Errorf("athlete client is not configured"))
				return nil, false, fmt.Errorf("athlete client is not configured")
			}
			athlete, f, err3 := athleteClient.GetAthleteByNameAndYear(start.AthleteName, start.AthleteYear)
			if err3 != nil {
				debugImportStartFailure("athlete lookup during update import", start, err3)
				return nil, false, err3
			}
			if !f {
				debugImportStartFailure("athlete not found during update import", start, fmt.Errorf("athlete with given AthleteName '%s' was not found", start.AthleteName))
				return nil, false, fmt.Errorf("athlete with given AthleteName '%s' was not found", start.AthleteName)
			}
			start.Athlete = athlete.Identifier
		}
	}

	needsTeamLookup := start.AthleteTeam.IsZero() &&
		start.AthleteTeamName != "" &&
		(existing.AthleteTeamName != start.AthleteTeamName || existing.AthleteTeam.IsZero())

	if needsTeamLookup {
		if teamClient == nil {
			debugImportStartFailure("team client missing during update import", start, fmt.Errorf("team client is not configured"))
			return nil, false, fmt.Errorf("team client is not configured")
		}
		team, f2, err4 := teamClient.GetTeamByName(start.AthleteTeamName)
		if err4 != nil {
			debugImportStartFailure("team lookup during update import", start, err4)
			return nil, false, err4
		}
		if !f2 {
			debugImportStartFailure("team not found during update import", start, fmt.Errorf("team with given AthleteTeamName '%s' was not found", start.AthleteTeamName))
			return nil, false, fmt.Errorf("team with given AthleteTeamName '%s' was not found", start.AthleteTeamName)
		}

		start.AthleteTeam = team.Identifier
	}

	changed := false
	if start.Certified == true && existing.Certified != start.Certified {
		existing.Certified = start.Certified
		changed = true
	}
	if start.Rank != 0 && existing.Rank != start.Rank {
		existing.Rank = start.Rank
		changed = true
	}
	if start.AthleteMeetingId != 0 && existing.AthleteMeetingId != start.AthleteMeetingId {
		existing.AthleteMeetingId = start.AthleteMeetingId
		changed = true
	}
	if start.AthleteName != "" && existing.AthleteName != start.AthleteName {
		existing.AthleteName = start.AthleteName
		changed = true
	}
	if start.AthleteAlias != "" && existing.AthleteAlias != start.AthleteAlias {
		existing.AthleteAlias = start.AthleteAlias
		changed = true
	}
	if !start.Athlete.IsZero() && existing.Athlete != start.Athlete {
		existing.Athlete = start.Athlete
		changed = true
	}
	if start.AthleteTeamName != "" && existing.AthleteTeamName != start.AthleteTeamName {
		existing.AthleteTeamName = start.AthleteTeamName
		changed = true
	}
	if !start.AthleteTeam.IsZero() && existing.AthleteTeam != start.AthleteTeam {
		existing.AthleteTeam = start.AthleteTeam
		changed = true
	}
	if start.AthleteYear != 0 && existing.AthleteYear != start.AthleteYear {
		existing.AthleteYear = start.AthleteYear
		changed = true
	}
	if start.Lane != 0 && existing.Lane != start.Lane {
		existing.Lane = start.Lane
		changed = true
	}
	if start.HeatNumber != 0 && existing.HeatNumber != start.HeatNumber {
		existing.HeatNumber = start.HeatNumber
		changed = true
	}
	if start.Points != 0 && existing.Points != start.Points {
		existing.Points = start.Points
		changed = true
	}
	if start.Rank != 0 && existing.Rank > start.Rank {
		existing.Rank = start.Rank
		changed = true
	}
	// if ranks set in start import -> update given ranks. Assumes that RankingId is set
	if len(start.Ranks) > 0 {
		err3 := addOrUpdateRanksInStart(&existing, start.Ranks)
		if err3 != nil {
			debugImportStartFailure("update ranks during import", start, err3)
			return nil, false, err3
		}
	}

	if changed {
		fmt.Printf("updating some values...\n")
		existing, err = UpdateStart(existing)
		if err != nil {
			debugImportStartFailure("persist updated start", start, err)
			return nil, false, err
		}
	}
	return &existing, false, nil
}

func debugImportStartFailure(stage string, start model.Start, err error) {
	fmt.Printf(
		"import start failed at %s: meeting=%q event=%d heat=%d lane=%d athlete_name=%q athlete_year=%d athlete_meeting_id=%d athlete_team_name=%q athlete=%s athlete_team=%s relay=%t start_id=%s reason=%v\n",
		stage,
		start.Meeting,
		start.Event,
		start.HeatNumber,
		start.Lane,
		start.AthleteName,
		start.AthleteYear,
		start.AthleteMeetingId,
		start.AthleteTeamName,
		start.Athlete.Hex(),
		start.AthleteTeam.Hex(),
		start.IsRelay,
		start.Identifier.Hex(),
		err,
	)
}

func ImportResult(start model.Start, result model.Result) (*model.Result, bool, error) {
	existing, found, err := GetStartFromImport(start)
	if err != nil {
		return nil, false, err
	}
	if !found {
		return nil, false, fmt.Errorf("start with given information not found")
	}
	res, c, err2 := UpdateStartAddOrUpdateResult(existing.Identifier, result)
	if err2 != nil {
		return nil, false, err2
	}
	return &res, c, nil
}

func ImportRank(start model.Start, rank model.Rank) (*model.Rank, bool, error) {
	existingStart, found, err := GetStartFromImport(start)
	if err != nil {
		return nil, false, err
	}
	if !found {
		return nil, false, fmt.Errorf("start with given information not found")
	}

	existingRanking, found, err3 := GetRankingByImport(rank.Ranking)
	if err3 != nil {
		return nil, false, err3
	}

	if !found {
		return nil, false, fmt.Errorf("ranking with given information not found")
	}

	rank.RankingId = existingRanking.Identifier

	res, c, err2 := UpdateStartAddOrUpdateRank(existingStart.Identifier, rank)
	if err2 != nil {
		return nil, false, err2
	}
	return &res, c, nil
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

// UpdateStartAddOrUpdateResult bool is true if the result was added, false if it was updated
func UpdateStartAddOrUpdateResult(startId primitive.ObjectID, result model.Result) (model.Result, bool, error) {
	start, err := GetStartById(startId)
	if err != nil {
		return model.Result{}, false, err
	}
	result.AddedAt = time.Now()
	found := false
	for i, r := range start.Results {
		if r.ResultType == result.ResultType && r.LapMeters == result.LapMeters {
			start.Results[i] = result
			found = true
			break
		}
	}
	if !found {
		start.Results = append(start.Results, result)
	}
	_, err2 := UpdateStart(start)
	if err2 != nil {
		return model.Result{}, false, err2
	}
	return result, !found, nil
}

// UpdateStartAddOrUpdateRank bool is true if the rank was added, false if it was updated, rank must have correct rankingId set!!
func UpdateStartAddOrUpdateRank(startId primitive.ObjectID, rank model.Rank) (model.Rank, bool, error) {
	start, err := GetStartById(startId)
	if err != nil {
		return model.Rank{}, false, err
	}

	rank.UpdatedAt = time.Now()
	found := false
	for i, r := range start.Ranks {
		if r.RankingId == rank.RankingId {
			start.Ranks[i] = rank
			found = true
			break
		}
	}
	if !found {
		rank.AddedAt = time.Now()
		start.Ranks = append(start.Ranks, rank)
	}
	_, err2 := UpdateStart(start)
	if err2 != nil {
		return model.Rank{}, false, err2
	}
	return rank, !found, nil
}

// addOrUpdateRanksInStart replaces ranks in a start but does not save the start
func addOrUpdateRanksInStart(start *model.Start, ranks []model.Rank) error {
	for _, rank := range ranks {
		if rank.RankingId.IsZero() {
			return fmt.Errorf("cannot add rank to start without ranking id set, use rank import to add new ranks  without knowing the ranking id beforehand")
		}
		rank.UpdatedAt = time.Now()
		found := false
		for i, r := range start.Ranks {
			if r.RankingId == rank.RankingId {
				start.Ranks[i] = rank
				found = true
				break
			}
		}
		if !found {
			rank.AddedAt = time.Now()
			start.Ranks = append(start.Ranks, rank)
		}
	}

	return nil
}
