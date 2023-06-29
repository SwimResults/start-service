package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Disqualification struct {
	Identifier       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`                             // automatically
	Type             string             `json:"type,omitempty" bson:"type,omitempty"`                           // PDF + DSV 	types: 'dns, disqualified, '
	Reason           string             `json:"reason,omitempty" bson:"reason,omitempty"`                       // PDF + DSV
	AnnouncementTime time.Time          `json:"announcement_time,omitempty" bson:"announcement_time,omitempty"` // PDF + DSV
	AddedAt          time.Time          `json:"added_at,omitempty" bson:"added_at,omitempty"`                   // automatically
}
