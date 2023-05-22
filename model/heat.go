package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Heat struct {
	Identifier      primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Meeting         string             `json:"meeting,omitempty" bson:"meeting,omitempty"`
	Event           string             `json:"event,omitempty" bson:"event,omitempty"`
	Number          int                `json:"number,omitempty" bson:"number,omitempty"`
	StartEstimation time.Time          `json:"start_estimation,omitempty" bson:"start_estimation,omitempty"`
	StartAt         time.Time          `json:"start_at,omitempty" bson:"start_at,omitempty"`
	FinishedAt      time.Time          `json:"finished_at,omitempty" bson:"finished_at,omitempty"`
}
