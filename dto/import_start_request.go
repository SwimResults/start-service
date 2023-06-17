package dto

import "github.com/swimresults/start-service/model"

type ImportStartRequestDto struct {
	Start model.Start `json:"start"`
}
