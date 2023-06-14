package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Start struct {
	Identifier         primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`                               // automatically
	Meeting            string             `json:"meeting,omitempty" bson:"meeting,omitempty"`                       // import service
	Event              int                `json:"event,omitempty" bson:"event,omitempty"`                           // PDF + DSV
	HeatNumber         int                `json:"heat_number" bson:"heat,omitempty"`                                // automatically
	Heat               Heat               `json:"heat,omitempty" bson:"-"`                                          // PDF + DSV
	Lane               int                `json:"lane,omitempty" bson:"lane,omitempty"`                             // PDF + DSV
	Athlete            primitive.ObjectID `json:"athlete,omitempty" bson:"athlete,omitempty"`                       // PDF + DSV (via athlete)
	AthleteMeetingId   string             `json:"athlete_meeting_id,omitempty" bson:"athlete_meeting_id,omitempty"` // DSV
	AthleteName        string             `json:"athlete_name,omitempty" bson:"athlete_name,omitempty"`             // PDF + DSV (+ Livetiming)
	AthleteTeam        primitive.ObjectID `json:"athlete_team,omitempty" bson:"athlete_team,omitempty"`             // PDF + DSV (+ Livetiming)
	Rank               int                `json:"rank,omitempty" bson:"rank,omitempty"`                             // PDF + DSV
	Certified          bool               `json:"certified,omitempty" bson:"certified,omitempty"`                   // PDF + DSV
	Results            []Result           `json:"results,omitempty" bson:"results,omitempty"`                       // PDF + DSV + Livetiming
	DisqualificationId primitive.ObjectID `json:"-" bson:"disqualification_id,omitempty"`                           // automatically
	Disqualification   Disqualification   `json:"disqualification,omitempty" bson:"-"`                              // PDF + DSV
	AddedAt            time.Time          `json:"added_at,omitempty" bson:"added_at,omitempty"`                     // automatically
	UpdatedAt          time.Time          `json:"updated_at,omitempty" bson:"updated_at,omitempty"`                 // automatically
}
