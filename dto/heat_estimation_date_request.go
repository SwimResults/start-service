package dto

import "time"

type HeatEstimationDateRequest struct {
	Time           time.Time `json:"time"`
	Events         []int     `json:"events"`
	UpdateTimeZone bool      `json:"update_time_zone"`
}
