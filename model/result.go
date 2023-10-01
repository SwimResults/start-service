package model

import (
	"time"
)

type Result struct {
	Time       time.Duration `json:"time,omitempty" bson:"time,omitempty"`               // PDF + DSV + Livetiming
	ResultType string        `json:"result_type,omitempty" bson:"result_type,omitempty"` // registration; livetiming_result; result_list; reaction; lap;
	LapMeters  int           `json:"lap_meters,omitempty" bson:"lap_meters,omitempty"`   // 25; 50; 75; 100; 125; 150; 175;
	AddedAt    time.Time     `json:"added_at,omitempty" bson:"added_at,omitempty"`       // automatically
}
