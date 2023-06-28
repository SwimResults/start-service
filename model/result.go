package model

import (
	"time"
)

type Result struct {
	Time       time.Duration `json:"time,omitempty" bson:"time,omitempty"`               // PDF + DSV + Livetiming
	ResultType string        `json:"result_type,omitempty" bson:"result_type,omitempty"` // registration; livetiming_result; result_list; reaction;
	AddedAt    time.Time     `json:"added_at,omitempty" bson:"added_at,omitempty"`       // automatically
}
