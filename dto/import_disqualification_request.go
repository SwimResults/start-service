package dto

import "github.com/swimresults/start-service/model"

type ImportDisqualificationRequestDto struct {
	Disqualification model.Disqualification `json:"disqualification"`
	Start            model.Start            `json:"start"`
}
