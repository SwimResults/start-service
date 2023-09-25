package dto

import "github.com/swimresults/start-service/model"

type MeetingHeatEventListDto struct {
	EventNumber int        `json:"event_number,omitempty"`
	Amount      int        `json:"amount,omitempty"`
	FirstHeat   model.Heat `json:"first_heat,omitempty"`
}
