package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Result struct {
	Identifier primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`                 // automatically
	Time       time.Duration      `json:"time,omitempty" bson:"time,omitempty"`               // PDF + DSV + Livetiming
	ResultType string             `json:"result_type,omitempty" bson:"result_type,omitempty"` // registration; livetiming_result; result; reaction;
	AddedAt    time.Time          `json:"added_at,omitempty" bson:"added_at,omitempty"`       // automatically
}
