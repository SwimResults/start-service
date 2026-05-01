package service

import (
	"context"
	"errors"
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

var rankingNotFoundError = "age group not found"

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

func RemoveRankingById(id primitive.ObjectID) error {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := rankingCollection.DeleteOne(ctx, bson.D{{"_id", id}})
	if err != nil {
		return err
	}
	return nil
}

func ImportRanking(group model.Ranking) (*model.Ranking, bool, error) {
	existing, err := GetRankingByMeetingAndEventAndAges(group.Meeting, group.Event, group.MinAge, group.MaxAge)
	if err != nil {
		if err.Error() != rankingNotFoundError {
			return nil, false, err
		}

		newGroup, err2 := AddRanking(group)
		if err2 != nil {
			return nil, false, err2
		}
		return &newGroup, true, nil
	}

	if group.Name != "" {
		existing.Name = group.Name
	}

	if group.Gender != "UNSET" {
		existing.Gender = group.Gender
	}

	newGroup, err := UpdateRanking(existing)
	if err != nil {
		return nil, false, err
	}
	return &newGroup, false, nil
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

	if min > max {
		a := min
		min = max
		max = a
	}

	group.Ages = []int{}
	for i := min; i <= max; i++ {
		if i < 1900 || i > 2050 {
			continue
		}
		group.Ages = append(group.Ages, i)
	}
}
