package dto

import "github.com/swimresults/start-service/model"

type ImportRankingRequestDto struct {
	Ranking model.Ranking `json:"ranking"`
}
