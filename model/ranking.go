package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Ranking struct {
	Identifier primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Meeting    string             `json:"meeting,omitempty" bson:"meeting,omitempty"`
	Event      int                `json:"event,omitempty" bson:"event,omitempty"`     // 15
	Gender     string             `json:"gender,omitempty" bson:"gender,omitempty"`   // MALE, FEMALE, MIXED (might be null, especially for non-mixed events)
	MinAge     string             `json:"min_age,omitempty" bson:"min_age,omitempty"` // 2004
	MaxAge     string             `json:"max_age,omitempty" bson:"max_age,omitempty"` // 2002
	Ages       []int              `json:"ages,omitempty" bson:"ages,omitempty"`       // 2002, 2003, 2004
	IsYear     bool               `json:"is_year,omitempty" bson:"is_year,omitempty"` // true
	Name       string             `json:"name,omitempty" bson:"name,omitempty"`       // Jahrgänge 2002 - 2004 (might not be set)
	AddedAt    time.Time          `json:"added_at,omitempty" bson:"added_at,omitempty"`
	UpdatedAt  time.Time          `json:"updated_at,omitempty" bson:"updated_at,omitempty"`
}
