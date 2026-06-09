package dto

import "github.com/swimresults/start-service/model"

type EventStartResultRequestDto struct {
	Ranking model.Ranking `json:"age_group"`
	Starts  []model.Start `json:"starts"`
}
