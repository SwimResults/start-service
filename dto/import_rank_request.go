package dto

import "github.com/swimresults/start-service/model"

type ImportRankRequestDto struct {
	Rank  model.Rank  `json:"rank"`
	Start model.Start `json:"start"`
}
