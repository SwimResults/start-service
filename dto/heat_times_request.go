package dto

import "time"

type HeatTimesRequestDto struct {
	Time time.Time `json:"time,omitempty"`
	Type string    `json:"type,omitempty"`
}
