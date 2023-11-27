package dto

import "github.com/swimresults/start-service/model"
import meetingModel "github.com/swimresults/meeting-service/model"

type EventStartResultRequestDto struct {
	AgeGroup meetingModel.AgeGroup `json:"age_group"`
	Starts   []model.Start         `json:"starts"`
}
