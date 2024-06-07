package dto

import "time"

type HeatEstimationDateRequest struct {
	Time time.Time `json:"time"`
}
