package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Heat struct {
	Identifier           primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`                                       // automatically
	Meeting              string             `json:"meeting,omitempty" bson:"meeting,omitempty"`                               // automatically
	Event                int                `json:"event,omitempty" bson:"event,omitempty"`                                   // PDF + DSV
	Number               int                `json:"number,omitempty" bson:"number,omitempty"`                                 // PDF + DSV
	StartEstimation      time.Time          `json:"start_estimation,omitempty" bson:"start_estimation,omitempty"`             // PDF				estimation from start list
	StartDelayEstimation time.Time          `json:"start_delay_estimation,omitempty" bson:"start_delay_estimation,omitempty"` // manual calculation of delay				estimation from delay
	StartAt              time.Time          `json:"start_at,omitempty" bson:"start_at,omitempty"`                             // automatically		actual moment when it started
	FinishedAt           time.Time          `json:"finished_at,omitempty" bson:"finished_at,omitempty"`                       // automatically		actual moment when it finished
}
