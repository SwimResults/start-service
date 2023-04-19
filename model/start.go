package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Start struct {
	Identifier       primitive.ObjectID  `json:"_id,omitempty" bson:"_id,omitempty"`
	Heat             Heat                `json:"heat,omitempty" bson:"heat,omitempty"`
	Lane             int                 `json:"lane,omitempty" bson:"lane,omitempty"`
	Athlete          primitive.ObjectID  `json:"athlete,omitempty" bson:"athlete,omitempty"`
	Delay            int                 `json:"delay,omitempty" bson:"delay,omitempty"`
	AddedAt          primitive.Timestamp `json:"added_at,omitempty" bson:"added_at,omitempty"`
	UpdatedAt        primitive.Timestamp `json:"updated_at,omitempty" bson:"updated_at,omitempty"`
	Rank             int                 `json:"rank,omitempty" bson:"rank,omitempty"`
	Certified        bool                `json:"certified,omitempty" bson:"certified,omitempty"`
	Results          []Result            `json:"results,omitempty" bson:"results,omitempty"`
	Disqualification Disqualification    `json:"disqualification,omitempty" bson:"disqualification,omitempty"`
}
