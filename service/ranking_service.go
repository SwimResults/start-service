package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/swimresults/start-service/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var rankingCollection *mongo.Collection

func rankingService(database *mongo.Database) {
	rankingCollection = database.Collection("ranking")
}

var rankingNotFoundError = "ranking not found"

func getRankingsByBsonDocument(d primitive.D) ([]model.Ranking, error) {
	var rankings []model.Ranking

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	queryOptions := options.FindOptions{}
	queryOptions.SetSort(bson.D{{"min_age", -1}})

	cursor, err := rankingCollection.Find(ctx, d, &queryOptions)
	if err != nil {
		return []model.Ranking{}, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var ranking model.Ranking
		cursor.Decode(&ranking)
		rankings = append(rankings, ranking)
	}

	if err := cursor.Err(); err != nil {
		return []model.Ranking{}, err
	}

	return rankings, nil
}

func getRankingByBsonDocument(d primitive.D) (model.Ranking, error) {
	groups, err := getRankingsByBsonDocument(d)
	if err != nil {
		return model.Ranking{}, err
	}

	if len(groups) <= 0 {
		return model.Ranking{}, errors.New(rankingNotFoundError)
	}

	return groups[0], nil
}

func GetRankings() ([]model.Ranking, error) {
	return getRankingsByBsonDocument(bson.D{})
}

func GetRankingsByMeeting(meeting string) ([]model.Ranking, error) {
	return getRankingsByBsonDocument(bson.D{{"meeting", meeting}})
}

func GetRankingsByMeetingAndEvent(meeting string, event int) ([]model.Ranking, error) {
	return getRankingsByBsonDocument(bson.D{{"meeting", meeting}, {"event", event}})
}

func GetRankingByMeetingAndEventAndAgesAndGender(meeting string, event int, minAge string, maxAge string, gender string) (model.Ranking, error) {
	return getRankingByBsonDocument(bson.D{{"meeting", meeting}, {"event", event}, {"min_age", minAge}, {"max_age", maxAge}, {"gender", gender}})
}

func GetRankingByMeetingAndEventAndAges(meeting string, event int, minAge string, maxAge string) (model.Ranking, error) {
	return getRankingByBsonDocument(bson.D{{"meeting", meeting}, {"event", event}, {"min_age", minAge}, {"max_age", maxAge}})
}

func GetRankingByMeetingAndEventAndName(meeting string, event int, name string) (model.Ranking, error) {
	return getRankingByBsonDocument(bson.D{{"meeting", meeting}, {"event", event}, {"name", name}})
}

func GetRankingByMeetingAndEventAndLenexId(meeting string, event int, lenexId string) (model.Ranking, error) {
	return getRankingByBsonDocument(bson.D{{"meeting", meeting}, {"event", event}, {"lenex_id", lenexId}})
}

func GetRankingByMeetingAndEventAndDsvId(meeting string, event int, dsvId string) (model.Ranking, error) {
	return getRankingByBsonDocument(bson.D{{"meeting", meeting}, {"event", event}, {"dsv_id", dsvId}})
}

func GetRankingById(id primitive.ObjectID) (model.Ranking, error) {
	rankings, err := getRankingsByBsonDocument(bson.D{{"_id", id}})
	if err != nil {
		return model.Ranking{}, err
	}

	if len(rankings) > 0 {
		return rankings[0], nil
	}

	return model.Ranking{}, errors.New("no entry with given id found")
}

// GetRankingByImport tries to find an existing ranking for the given ranking in the import, bool is true if ranking exists and false if it does not exist, error is set if an error occurs during the process
func GetRankingByImport(ranking model.Ranking) (model.Ranking, bool, error) {

	var existing model.Ranking
	var err error

	if !ranking.Identifier.IsZero() {
		existing, err = GetRankingById(ranking.Identifier)
		if err != nil {
			return model.Ranking{}, false, errors.New("could not find ranking by identifier '" + ranking.Identifier.String() + "', even though it was given")
		}
		return existing, true, nil
	}

	if ranking.Meeting == "" ||
		ranking.Event == 0 {
		return model.Ranking{}, false, fmt.Errorf("missing arguments"+
			"(expected: meeting; event; ..."+
			"got: '%s', '%d')",
			ranking.Meeting,
			ranking.Event)
	}

	if ranking.LenexId != "" {
		// find by lenex id
		existing, err = GetRankingByMeetingAndEventAndLenexId(ranking.Meeting, ranking.Event, ranking.LenexId)
		if err != nil {
			if err.Error() != rankingNotFoundError {
				return model.Ranking{}, false, err
			}
		} else {
			return existing, true, nil
		}

	}

	if ranking.DsvId != "" {
		// find by dsv id
		existing, err = GetRankingByMeetingAndEventAndDsvId(ranking.Meeting, ranking.Event, ranking.DsvId)
		if err != nil {
			if err.Error() != rankingNotFoundError {
				return model.Ranking{}, false, err
			}
		} else {
			return existing, true, nil
		}
	}

	if ranking.Name != "" { // should not fail but just jump to next, name could be new
		existing, err = GetRankingByMeetingAndEventAndName(ranking.Meeting, ranking.Event, ranking.Name)
		if err != nil {
			if err.Error() != rankingNotFoundError {
				return model.Ranking{}, false, err
			}
		} else {
			return existing, true, nil
		}
	}

	existing, err = GetRankingByMeetingAndEventAndAges(ranking.Meeting, ranking.Event, ranking.MinAge, ranking.MaxAge)
	if err != nil {
		if err.Error() != rankingNotFoundError {
			return model.Ranking{}, false, err
		}

		return model.Ranking{}, false, nil
	}

	return existing, true, nil
}

func RemoveRankingById(id primitive.ObjectID) error {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := rankingCollection.DeleteOne(ctx, bson.D{{"_id", id}})
	if err != nil {
		return err
	}
	return nil
}

func ImportRanking(ranking model.Ranking) (*model.Ranking, bool, error) {
	existing, found, err := GetRankingByImport(ranking)
	if err != nil {
		return nil, false, err
	}
	if !found {
		newRanking, err2 := AddRanking(ranking)
		if err2 != nil {
			return nil, false, err2
		}
		return &newRanking, true, nil
	}

	if ranking.Name != "" {
		existing.Name = ranking.Name
	}

	if ranking.Gender != "UNSET" {
		existing.Gender = ranking.Gender
	}

	if ranking.LenexId != "" {
		existing.LenexId = ranking.LenexId
	}

	if ranking.DsvId != "" {
		existing.DsvId = ranking.DsvId
	}

	newRanking, err := UpdateRanking(existing)
	if err != nil {
		return nil, false, err
	}
	return &newRanking, false, nil
}

func AddRanking(ranking model.Ranking) (model.Ranking, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ranking.AddedAt = time.Now()
	ranking.UpdatedAt = time.Now()

	SetAgesForRanking(&ranking)

	r, err := rankingCollection.InsertOne(ctx, ranking)
	if err != nil {
		return model.Ranking{}, err
	}

	return GetRankingById(r.InsertedID.(primitive.ObjectID))
}

func UpdateRanking(ranking model.Ranking) (model.Ranking, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ranking.UpdatedAt = time.Now()

	SetAgesForRanking(&ranking)

	_, err := rankingCollection.ReplaceOne(ctx, bson.D{{"_id", ranking.Identifier}}, ranking)
	if err != nil {
		return model.Ranking{}, err
	}

	return GetRankingById(ranking.Identifier)
}

func SetAgesForRanking(group *model.Ranking) {
	if !group.IsYear {
		return
	}

	min, _ := strconv.Atoi(group.MinAge)
	max, _ := strconv.Atoi(group.MaxAge)

	if min <= 0 {
		min = 2100
	}

	if max <= 0 {
		max = 1900
	}

	group.Ages = []int{}
	for i := max; i <= min; i++ {
		group.Ages = append(group.Ages, i)
	}
}
