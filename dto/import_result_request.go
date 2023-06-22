package dto

import "github.com/swimresults/start-service/model"

type ImportResultRequestDto struct {
	Result model.Result `json:"result"`
	Start  model.Start  `json:"start"`
}
