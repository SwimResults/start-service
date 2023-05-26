package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Disqualification struct {
	Identifier       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Reason           string             `json:"reason,omitempty" bson:"reason,omitempty"`
	AnnouncementTime time.Time          `json:"announcement_time,omitempty" bson:"announcement_time,omitempty"`
	AddedAt          time.Time          `json:"added_at,omitempty" bson:"added_at,omitempty"`
}
