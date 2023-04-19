package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Result struct {
	Identifier primitive.ObjectID  `json:"_id,omitempty" bson:"_id,omitempty"`
	Time       time.Duration       `json:"time,omitempty" bson:"time,omitempty"`
	ResultType string              `json:"result_type,omitempty" bson:"result_type,omitempty"`
	AddedAt    primitive.Timestamp `json:"added_at,omitempty" bson:"added_at,omitempty"`
	Start      Start               `json:"start,omitempty" bson:"start,omitempty"`
}
