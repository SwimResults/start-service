package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Start struct {
	Identifier         primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Meeting            string             `json:"meeting,omitempty" bson:"meeting,omitempty"`
	Event              int                `json:"event,omitempty" bson:"event,omitempty"`
	HeatNumber         int                `json:"heat_number" bson:"heat,omitempty"`
	Heat               Heat               `json:"heat,omitempty" bson:"-"`
	Lane               int                `json:"lane,omitempty" bson:"lane,omitempty"`
	Athlete            primitive.ObjectID `json:"athlete,omitempty" bson:"athlete,omitempty"`
	Delay              int                `json:"delay,omitempty" bson:"delay,omitempty"`
	AddedAt            time.Time          `json:"added_at,omitempty" bson:"added_at,omitempty"`
	UpdatedAt          time.Time          `json:"updated_at,omitempty" bson:"updated_at,omitempty"`
	Rank               int                `json:"rank,omitempty" bson:"rank,omitempty"`
	Certified          bool               `json:"certified,omitempty" bson:"certified,omitempty"`
	Results            []Result           `json:"results,omitempty" bson:"results,omitempty"`
	DisqualificationId primitive.ObjectID `json:"-" bson:"disqualification_id,omitempty"`
	Disqualification   Disqualification   `json:"disqualification,omitempty" bson:"-"`
}
